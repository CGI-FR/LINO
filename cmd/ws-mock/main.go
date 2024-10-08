package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	log.SetFlags(2)

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

// run starts a http.Server for the passed in address
// with all requests handled by echoServer.
func run() error {
	if len(os.Args) < 2 {
		return errors.New("please provide an address to listen on as the first argument")
	}

	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())

	s := &http.Server{
		Handler: echoServer{
			logf: log.Printf,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)

	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}

// echoServer is the WebSocket echo server implementation.
// It ensures the client speaks the echo subprotocol and
// only allows one message every 100ms with a 10 message burst.
type echoServer struct {
	// logf controls where logs are sent.
	logf func(f string, v ...interface{})
}

func (s echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if len(auth) > 0 {
		log.Printf("receive header auth  %v", auth)
		if userpass, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic ")); err != nil {
			s.logf("failed to decode basic auth %s %v", r.RemoteAddr, err)
			return
		} else {
			creds := strings.Split(string(userpass), ":")
			if creds[0] != "user" {
				s.logf("invalid user %s %s", creds[0], r.RemoteAddr)
				return
			}
			if creds[1] != "password" {
				s.logf("invalid password %s %s", creds[1], r.RemoteAddr)
				return
			}
		}
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"echo"},
	})
	if err != nil {
		s.logf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	/*
		if c.Subprotocol() != "lino" {
			c.Close(websocket.StatusPolicyViolation, "client must speak the lino subprotocol not "+c.Subprotocol())
			return
		}
	*/

	for {
		err = lino(r.Context(), c)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}

		if err != nil {
			s.logf("failed to echo with %v: %v", r.RemoteAddr, err)
			return
		}
	}
}

// lino reads from the WebSocket connection and then writes
// the received message back to it.
// The entire function has 10s to complete.
func lino(ctx context.Context, c *websocket.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	v := map[string]interface{}{}

	if err := wsjson.Read(ctx, c, &v); err != nil {
		return err
	}

	log.Printf("receive message  %v", v)

	var err error

	switch v["action"] {
	case "ping":
		err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": nil, "next": false})

	case "extract_tables":
		err = wsjson.Write(ctx, c, map[string]interface{}{
			"id":    v["id"],
			"error": nil,
			"next":  false,
			"payload": []map[string]interface{}{
				{
					"name": "ACT",
					"keys": []string{
						"ACTNO",
					},
				},
				{
					"name": "CATALOG",
					"keys": []string{
						"NAME",
					},
				},
				{
					"name": "CUSTOMER",
					"keys": []string{
						"CID",
					},
				},
				{
					"name": "DEPARTMENT",
					"keys": []string{
						"DEPTNO",
					},
				},
				{
					"name": "EMPLOYEE",
					"keys": []string{
						"EMPNO",
					},
				},
				{
					"name": "EMP_PHOTO",
					"keys": []string{
						"EMPNO",
						"PHOTO_FORMAT",
					},
				},
				{
					"name": "EMP_RESUME",
					"keys": []string{
						"EMPNO",
						"RESUME_FORMAT",
					},
				},
				{
					"name": "INVENTORY",
					"keys": []string{
						"PID",
					},
				},
				{
					"name": "PRODUCT",
					"keys": []string{
						"PID",
					},
				},
				{
					"name": "PRODUCTSUPPLIER",
					"keys": []string{
						"PID",
						"SID",
					},
				},
				{
					"name": "PROJACT",
					"keys": []string{
						"ACSTDATE",
						"ACTNO",
						"PROJNO",
					},
				},
				{
					"name": "PROJECT",
					"keys": []string{
						"PROJNO",
					},
				},
				{
					"name": "PURCHASEORDER",
					"keys": []string{
						"POID",
					},
				},
				{
					"name": "SUPPLIERS",
					"keys": []string{
						"SID",
					},
				},
			},
		},
		)
	case "extract_relations":
		err = relations(ctx, c, v)
	case "pull_open":
		var shouldReturn bool
		var returnValue error
		err, shouldReturn, returnValue = pull(v, err, ctx, c)
		if shouldReturn {
			return returnValue
		}

	case "push_open":
		schema := v["payload"].(map[string]interface{})["schema"]
		tables := v["payload"].(map[string]interface{})["tables"].([]interface{})
		mode := v["payload"].(map[string]interface{})["mode"]
		disableConstraints := v["payload"].(map[string]interface{})["disable_constraints"].(bool)
		log.Printf("schema  %v", schema)
		log.Printf("mode  %v", mode)
		log.Printf("disableConstraints  %v", disableConstraints)
		if disableConstraints {
			err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": "constraint is locked", "next": false})
		} else {
			for _, table := range tables {
				log.Printf("table  %v", table)
			}
			err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": nil, "next": false})
		}

	case "push_data":
		table := v["payload"].(map[string]interface{})["table"].(string)
		row := v["payload"].(map[string]interface{})["row"].(map[string]interface{})
		log.Printf("insert %v in table  %s", row, table)
		err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": nil, "next": false})
	case "push_commit":
		err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": nil, "next": false})
	case "push_close":
		err = wsjson.Write(ctx, c, map[string]interface{}{"id": v["id"], "error": nil, "next": false})
	}
	return err
}

func relations(ctx context.Context, c *websocket.Conn, v map[string]interface{}) error {
	err := wsjson.Write(ctx, c, map[string]interface{}{
		"id":    v["id"],
		"error": nil,
		"next":  false,
		"payload": []map[string]interface{}{
			{
				"name": "FK_EMP_PHOTO",
				"parent": map[string]interface{}{
					"name": "EMP_PHOTO",
					"keys": []string{
						"EMPNO",
					},
				},
				"child": map[string]interface{}{
					"name": "EMPLOYEE",
					"keys": []string{
						"PK_EMPLOYEE",
					},
				},
			},

			{
				"name": "FK_EMP_RESUME",
				"parent": map[string]interface{}{
					"name": "EMP_RESUME",
					"keys": []string{
						"EMPNO",
					},
				},
				"child": map[string]interface{}{
					"name": "EMPLOYEE",
					"keys": []string{
						"PK_EMPLOYEE",
					},
				},
			},
			{
				"name": "FK_PO_CUST",
				"parent": map[string]interface{}{
					"name": "PURCHASEORDER",
					"keys": []string{
						"CUSTID",
					},
				},
				"child": map[string]interface{}{
					"name": "CUSTOMER",
					"keys": []string{
						"PK_CUSTOMER",
					},
				},
			},
			{
				"name": "FK_PROJECT_1",
				"parent": map[string]interface{}{
					"name": "PROJECT",
					"keys": []string{
						"DEPTNO",
					},
				},
				"child": map[string]interface{}{
					"name": "DEPARTMENT",
					"keys": []string{
						"PK_DEPARTMENT",
					},
				},
			},
			{
				"name": "FK_PROJECT_2",
				"parent": map[string]interface{}{
					"name": "PROJECT",
					"keys": []string{
						"RESPEMP",
					},
				},
				"child": map[string]interface{}{
					"name": "EMPLOYEE",
					"keys": []string{
						"PK_EMPLOYEE",
					},
				},
			},
			{
				"name": "RDE",
				"parent": map[string]interface{}{
					"name": "DEPARTMENT",
					"keys": []string{
						"MGRNO",
					},
				},
				"child": map[string]interface{}{
					"name": "EMPLOYEE",
					"keys": []string{
						"PK_EMPLOYEE",
					},
				},
			},
			{
				"name": "RED",
				"parent": map[string]interface{}{
					"name": "EMPLOYEE",
					"keys": []string{
						"WORKDEPT",
					},
				},
				"child": map[string]interface{}{
					"name": "DEPARTMENT",
					"keys": []string{
						"PK_DEPARTMENT",
					},
				},
			},
			{
				"name": "REPAPA",
				"parent": map[string]interface{}{
					"name": "EMPPROJACT",
					"keys": []string{
						"PROJNO,ACTNO,EMSTDATE",
					},
				},
				"child": map[string]interface{}{
					"name": "PROJACT",
					"keys": []string{
						"PK_PROJACT",
					},
				},
			},
			{
				"name": "ROD",
				"parent": map[string]interface{}{
					"name": "DEPARTMENT",
					"keys": []string{
						"ADMRDEPT",
					},
				},
				"child": map[string]interface{}{
					"name": "DEPARTMENT",
					"keys": []string{
						"PK_DEPARTMENT",
					},
				},
			},
			{
				"name": "RPAA",
				"parent": map[string]interface{}{
					"name": "ACT",
					"keys": []string{
						"ACTNO",
					},
				},
				"child": map[string]interface{}{
					"name": "ACT",
					"keys": []string{
						"PK_ACT",
					},
				},
			},
			{
				"name": "RPAP",
				"parent": map[string]interface{}{
					"name": "PROJACT",
					"keys": []string{
						"PROJNO",
					},
				},
				"child": map[string]interface{}{
					"name": "PROJECT",
					"keys": []string{
						"PK_PROJECT",
					},
				},
			},
			{
				"name": "RPP",
				"parent": map[string]interface{}{
					"name": "PROJECT",
					"keys": []string{
						"MAJPROJ",
					},
				},
				"child": map[string]interface{}{
					"name": "PROJECT",
					"keys": []string{
						"PROJNO",
					},
				},
			},
		},
	},
	)
	return err
}

func pull(v map[string]interface{}, err error, ctx context.Context, c *websocket.Conn) (error, bool, error) {
	switch v["payload"].(map[string]interface{})["table"] {
	case "PROJECT":

		for i := 0; i < 1000; i++ {
			err = wsjson.Write(ctx, c, map[string]interface{}{
				"id":    v["id"],
				"error": nil,
				"next":  true,
				"payload": json.RawMessage(
					[]byte(fmt.Sprintf("{\"PROJNO\": %d, \"MAJPROJ\":2, \"DEPTNO\":1}", i)),
				),
			},
			)
		}
	case "DEPARTMENT":
		err = wsjson.Write(ctx, c, map[string]interface{}{
			"id":    v["id"],
			"error": nil,
			"next":  true,
			"payload": json.RawMessage(
				[]byte("{\"DEPTNO\": 1}"),
			),
		},
		)
	}

	if err != nil {
		return nil, true, err
	}

	err = wsjson.Write(ctx, c, map[string]interface{}{
		"id":    v["id"],
		"error": nil,
		"next":  false,
	},
	)
	return err, false, nil
}
