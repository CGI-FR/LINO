package push

import (
	"fmt"
	"strings"
	"time"

	"makeit.imfr.cgi.com/lino/pkg/push"

	// import Oracle connector

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
func (e *OracleDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, OracleDialect{}, e.logger)
}

// OracleDialect inject oracle variations
type OracleDialect struct{}

// Placeholde return the variable format for postgres
func (d OracleDialect) Placeholder(position int) string {
	return fmt.Sprintf(":v%d", position)
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d OracleDialect) EnableConstraintsStatement(tableName string) string {
	tableComposants := strings.Split(tableName, ".")
	switch len(tableComposants) {
	case 1:
		return fmt.Sprintf(`BEGIN
	FOR c IN
	(SELECT c.owner, c.table_name, c.constraint_name
	 FROM user_constraints c, user_tables t
	 WHERE c.table_name = t.table_name
	 AND c.owner = sys_context( 'userenv', 'current_schema' )
	 AND c.table_name = '%s'
	 AND c.status = 'DISABLED'
	 ORDER BY c.constraint_type)
	LOOP
	  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" enable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, tableName)

	case 2:
		return fmt.Sprintf(`BEGIN
		FOR c IN
		(SELECT c.owner, c.table_name, c.constraint_name
		 FROM user_constraints c, user_tables t
		 WHERE c.table_name = t.table_name
		 AND c.owner = '%s'
		 AND c.table_name = '%s'
		 AND c.status = 'DISABLED'
		 ORDER BY c.constraint_type)
		LOOP
		  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" enable constraint ' || c.constraint_name);
		END LOOP;
	  END;
	  `, tableComposants[0], tableComposants[0])
	default:
		return ""
	}
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d OracleDialect) DisableConstraintsStatement(tableName string) string {
	tableComposants := strings.Split(tableName, ".")
	switch len(tableComposants) {
	case 1:
		return fmt.Sprintf(`BEGIN
	FOR c IN
		(SELECT c.owner, c.table_name, c.constraint_name
		FROM user_constraints c, user_tables t
		WHERE c.table_name = t.table_name
		AND c.owner = sys_context( 'userenv', 'current_schema' )
		AND c.table_name = '%s'
		AND c.status = 'ENABLED'
		AND NOT (t.iot_type IS NOT NULL AND c.constraint_type = 'P')
		ORDER BY c.constraint_type DESC)
	LOOP
		dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, tableName)
	case 2:
		return fmt.Sprintf(`BEGIN
	FOR c IN
		(SELECT c.owner, c.table_name, c.constraint_name
		FROM user_constraints c, user_tables t
		WHERE c.table_name = t.table_name
		AND c.owner = '%s'
		AND c.table_name = '%s'
		AND c.status = 'ENABLED'
		AND NOT (t.iot_type IS NOT NULL AND c.constraint_type = 'P')
		ORDER BY c.constraint_type DESC)
	LOOP
		dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
	END LOOP;
	END;
	`, tableComposants[0], tableComposants[1])
	default:
		return ""
	}
}

// TruncateStatement generate statement to truncat table content
func (d OracleDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s", tableName)
}

func (d OracleDialect) InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string {
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s);", tableName, strings.Join(columns, ","), strings.Join(values, ","))
}

func (d OracleDialect) UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string) (string, *push.Error) {
	sql := &strings.Builder{}
	sql.Write([]byte("UPDATE "))
	sql.Write([]byte(tableName))
	sql.Write([]byte(" SET "))
	for index, column := range columns {
		sql.Write([]byte(column))
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, uValues[index])
		if index+1 < len(columns) {
			sql.Write([]byte(", "))
		}
	}
	if len(primaryKeys) > 0 {
		sql.Write([]byte(" WHERE "))
	} else {
		return "", &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}
	for index, pk := range primaryKeys {
		sql.Write([]byte(pk))
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, pValues[index])
		if index+1 < len(primaryKeys) {
			sql.Write([]byte(" AND "))
		}
	}
	return sql.String(), nil
}

// IsDuplicateError check if error is a duplicate error
func (d OracleDialect) IsDuplicateError(err error) bool {
	// ORA-00001
	return strings.Contains(err.Error(), "ORA-00001")
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
