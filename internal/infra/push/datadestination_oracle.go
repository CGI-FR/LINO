package push

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgi-fr/lino/pkg/push"

	// import Oracle connector
	_ "github.com/godror/godror"
)

// OracleDataDestinationFactory exposes methods to create new Oracle extractors.
type OracleDataDestinationFactory struct{}

// NewOracleDataDestinationFactory creates a new Oracle datadestination factory.
func NewOracleDataDestinationFactory() *OracleDataDestinationFactory {
	return &OracleDataDestinationFactory{}
}

// New return a Oracle pusher
func (e *OracleDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, OracleDialect{})
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

// InsertStatement generate insert statement
func (d OracleDialect) InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor) {
	protectedColumns := []string{}
	for _, value := range selectValues {
		protectedColumns = append(protectedColumns, fmt.Sprintf("\"%s\"", value.name))
	}

	sql := &strings.Builder{}
	sql.WriteString("INSERT INTO ")
	sql.WriteString(tableName)
	sql.WriteString("(")
	sql.WriteString(strings.Join(protectedColumns, ","))
	sql.WriteString(") VALUES (")
	for i := 1; i <= len(selectValues); i++ {
		sql.WriteString(d.Placeholder(i))
		if i < len(selectValues) {
			sql.WriteString(", ")
		}
	}
	sql.WriteString(")")

	return sql.String(), selectValues
}

// UpdateStatement
func (d OracleDialect) UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
	sql := &strings.Builder{}
	sql.WriteString("UPDATE ")
	sql.WriteString(tableName)
	sql.WriteString(" SET ")

	for index, column := range selectValues {
		// don't update primary key, except if it's in whereValues
		if isAPrimaryKey(column.name, primaryKeys) {
			isInWhere := false
			for _, pk := range whereValues {
				if column.name == pk.name {
					isInWhere = true
					break
				}
			}
			if !isInWhere {
				continue
			}
		}

		headers = append(headers, column)

		sql.WriteString(column.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(index + 1))
		if index+1 < len(selectValues) {
			sql.WriteString(", ")
		}
	}
	if len(whereValues) > 0 {
		sql.WriteString(" WHERE ")
	} else {
		return "", nil, &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}
	for index, pk := range whereValues {
		headers = append(headers, pk)

		sql.WriteString(pk.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(len(selectValues) + index + 1))
		if index+1 < len(whereValues) {
			sql.Write([]byte(" AND "))
		}
	}

	return sql.String(), headers, nil
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
