package push

import (
	"fmt"
	"time"

	"makeit.imfr.cgi.com/lino/pkg/push"

	// import Oracle connector
	"github.com/godror/godror"
	_ "github.com/godror/godror"
)

// OracleDataDestinationFactory exposes methods to create new Oracle extractors.
type OracleDataDestinationFactory struct {
	logger push.Logger
}

// NewOracleDataDestinationFactory creates a new Oracle datadestination factory.
func NewOracleDataDestinationFactory(l push.Logger) *OracleDataDestinationFactory {
	return &OracleDataDestinationFactory{l}
}

// New return a Oracle pusher
func (e *OracleDataDestinationFactory) New(url string) push.DataDestination {
	return NewSQLDataDestination(url, OracleDialect{}, e.logger)
}

// OracleDialect inject oracle variations
type OracleDialect struct{}

// Placeholde return the variable format for postgres
func (d OracleDialect) Placeholder(position int) string {
	return fmt.Sprintf(":v%d", position)
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d OracleDialect) EnableConstraintsStatement(tableName string) string {
	return fmt.Sprintf(`BEGIN
	FOR c IN
	(SELECT c.owner, c.table_name, c.constraint_name
	 FROM user_constraints c, user_tables t
	 WHERE c.table_name = t.table_name
	 AND c.table_name = '%s'
	 AND c.status = 'DISABLED'
	 ORDER BY c.constraint_type)
	LOOP
	  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" enable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, tableName)
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d OracleDialect) DisableConstraintsStatement(tableName string) string {
	return fmt.Sprintf(`BEGIN
	FOR c IN
	(SELECT c.owner, c.table_name, c.constraint_name
	 FROM user_constraints c, user_tables t
	 WHERE c.table_name = t.table_name
	 AND c.table_name = '%s'
	 AND c.status = 'ENABLED'
	 AND NOT (t.iot_type IS NOT NULL AND c.constraint_type = 'P')
	 ORDER BY c.constraint_type DESC)
	LOOP
	  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, tableName)
}

// TruncateStatement generate statement to truncat table content
func (d OracleDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)
}

// IsDuplicateError check if error is a duplicate error
func (d OracleDialect) IsDuplicateError(err error) bool {
	// ORA-00001
	if oarErr, ok := err.(*godror.OraErr); ok {
		return oarErr.Code() == 1
	}
	return false
}

// ConvertValue before load
func (d OracleDialect) ConvertValue(from push.Value) push.Value {
	// FIXME: Workaround to parse time from json
	aTime, err := time.Parse("2006-01-02T15:04:05.999Z07:00", fmt.Sprintf("%v", from))
	if err != nil {
		return from
	} else {
		return aTime
	}
}
