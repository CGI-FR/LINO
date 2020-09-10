package id

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"

	"makeit.imfr.cgi.com/lino/pkg/id"
)

// GraphVizExporter export to SVG graph representation and open browser.
type GraphVizExporter struct{}

// NewGraphVizExporter create a new GraphVizExporter
func NewGraphVizExporter() *GraphVizExporter {
	return &GraphVizExporter{}
}

// Export to a temporary svg file and open it.
func (e *GraphVizExporter) Export(ep id.PullerPlan) *id.Error {
	dotexe, err := exec.LookPath("dot")
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	graphName := "G"
	graphViz := gographviz.NewGraph()
	err = graphViz.SetName(graphName)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}
	err = graphViz.SetDir(true)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}
	for i := uint(0); i < ep.Tables().Len(); i++ {
		err = graphViz.AddNode(graphName, strconv.Quote(ep.Tables().Table(i).Name()), nil)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}
	}
	for i := uint(0); i < ep.Len(); i++ {
		if ep.Step(i).Tables().Len() > 1 {
			compName := strconv.Quote("cluster" + fmt.Sprint(i))
			err = graphViz.AddSubGraph(graphName, compName, nil)
			if err != nil {
				return &id.Error{Description: err.Error()}
			}
			for j := uint(0); j < ep.Step(i).Tables().Len(); j++ {
				table := ep.Step(i).Tables().Table(j)
				err = graphViz.AddNode(compName, strconv.Quote(table.Name()), nil)
				if err != nil {
					return &id.Error{Description: err.Error()}
				}
			}
		}
	}
	for i := uint(0); i < ep.Relations().Len(); i++ {
		rel := ep.Relations().Relation(i)
		var relname string
		switch {
		case rel.LookUpChild() && rel.LookUpParent():
			relname = `↔`
		case rel.LookUpChild():
			relname = `→`
		case rel.LookUpParent():
			relname = `←`
		}

		err = graphViz.AddEdge(strconv.Quote(rel.Parent().Name()), strconv.Quote(rel.Child().Name()), true, map[string]string{"label": relname})
		if err != nil {
			return &id.Error{Description: err.Error()}
		}
	}

	file := filepath.Join(os.TempDir(), "lino-graph-export.svg")

	cmd := exec.Command(dotexe, "-Tsvg", "-o", file)
	cmd.Stdin = strings.NewReader(graphViz.String())
	if err := cmd.Run(); err != nil {
		return &id.Error{Description: err.Error()}
	}

	err = open(file)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	return nil
}

func open(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}
	return nil
}
