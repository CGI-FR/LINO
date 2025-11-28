package push

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cgi-fr/lino/pkg/push"

	// import Oracle connector
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
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

// DisableConstraintsStatement generate statement to deactivate constraints
func (d OracleDialect) DisableConstraintsStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	sql := &strings.Builder{}
	sql.WriteString(
		`BEGIN
		 FOR c IN(
		 SELECT c.owner, c.table_name, c.constraint_name
		 FROM user_constraints c
		 CONNECT BY PRIOR c.constraint_name = c.r_constraint_name
		 START WITH c.constraint_name IN (
			SELECT c.constraint_name
			FROM user_constraints c
		 	WHERE c.status = 'ENABLED' AND c.table_name = '`)
	if len(schemaAndTable) == 2 {
		sql.WriteString(schemaAndTable[1])
		sql.WriteString("' AND c.owner = '")
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("'")
	} else {
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("' AND c.owner = sys_context( 'userenv', 'current_schema' )")
	}
	sql.WriteString(`)
		LOOP
			dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
		END LOOP;
	END;`)
	return sql.String()
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d OracleDialect) EnableConstraintsStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	sql := &strings.Builder{}
	sql.WriteString(
		`BEGIN
		 FOR c IN(
		 SELECT c.owner, c.table_name, c.constraint_name
		 FROM user_constraints c
		 CONNECT BY PRIOR c.constraint_name = c.r_constraint_name
		 START WITH c.constraint_name IN (
			SELECT c.constraint_name
			FROM user_constraints c
		 	WHERE c.status = 'DISABLED' AND c.table_name = '`)
	if len(schemaAndTable) == 2 {
		sql.WriteString(schemaAndTable[1])
		sql.WriteString("' AND c.owner = '")
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("'")
	} else {
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("' AND c.owner = sys_context( 'userenv', 'current_schema' )")
	}
	sql.WriteString(`)
		LOOP
			dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
		END LOOP;
	END;`)
	return sql.String()
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

// UpsertStatement
func (d OracleDialect) UpsertStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
	sql := &strings.Builder{}
	sql.WriteString("MERGE INTO ")
	sql.WriteString(tableName)
	sql.WriteString(" target USING (SELECT ")

	for i, col := range selectValues {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(d.Placeholder(i + 1))
		sql.WriteString(" AS \"")
		sql.WriteString(col.name)
		sql.WriteString("\"")
	}
	sql.WriteString(" FROM dual) source ON (")

	for i, pk := range primaryKeys {
		if i > 0 {
			sql.WriteString(" AND ")
		}
		sql.WriteString(fmt.Sprintf("target.\"%s\" = source.\"%s\"", pk, pk))
	}
	sql.WriteString(") WHEN MATCHED THEN UPDATE SET ")

	first := true
	for _, col := range selectValues {
		if isAPrimaryKey(col.name, primaryKeys) {
			continue
		}
		if !first {
			sql.WriteString(", ")
		}
		sql.WriteString(fmt.Sprintf("target.\"%s\" = source.\"%s\"", col.name, col.name))
		first = false
	}

	sql.WriteString(" WHEN NOT MATCHED THEN INSERT (")
	for i, col := range selectValues {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(fmt.Sprintf("\"%s\"", col.name))
	}
	sql.WriteString(") VALUES (")
	for i, col := range selectValues {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(fmt.Sprintf("source.\"%s\"", col.name))
	}
	sql.WriteString(")")

	return sql.String(), selectValues, nil
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

		errColumn := appendColumnToSQL(column, sql, d, index)
		if errColumn != nil {
			return "", nil, errColumn
		}

		if index+1 < len(selectValues) {
			sql.WriteString(", ")
		}
	}
	if len(whereValues) > 0 {
		sql.WriteString(" WHERE ")
	} else {
		return "", nil, &push.Error{
			Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName),
		}
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
func (d OracleDialect) ConvertValue(from push.Value, descriptor ValueDescriptor) push.Value {
	if descriptor.column != nil && (descriptor.column.Import() == "file" || descriptor.column.Import() == "blob") {
		//nolint:gocritic
		switch v := from.(type) {
		case []byte:
			return go_ora.Blob{Data: v, Valid: true}
		}
	}
	if descriptor.column != nil && descriptor.column.Export() == "numeric" {
		switch v := from.(type) {
		case json.Number:
			return v.String()
		}
	}
	// FIXME: Workaround to parse time from json
	aTime, err := time.Parse("2006-01-02T15:04:05.999Z07:00", fmt.Sprintf("%v", from))
	if err != nil {
		return from
	} else {
		return aTime
	}
}

func (d OracleDialect) CanDisableIndividualConstraints() bool {
	return true
}

func (d OracleDialect) ReadConstraintsStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	sql := &strings.Builder{}
	sql.WriteString(
		`SELECT c.owner || '.' || c.table_name table_name, c.constraint_name
		 FROM user_constraints c
		 CONNECT BY PRIOR c.constraint_name = c.r_constraint_name
		 START WITH c.constraint_name IN (
			SELECT c.constraint_name
			FROM user_constraints c
		 	WHERE c.status = 'ENABLED' AND c.table_name = '`)
	if len(schemaAndTable) == 2 {
		sql.WriteString(schemaAndTable[1])
		sql.WriteString("' AND c.owner = '")
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("'")
	} else {
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("' AND c.owner = sys_context( 'userenv', 'current_schema' )")
	}
	sql.WriteString(") ORDER BY c.constraint_type DESC") // disable FK then PK then others
	return sql.String()
}

func (d OracleDialect) DisableConstraintStatement(tableName string, constraintName string) string {
	sql := &strings.Builder{}
	sql.WriteString("ALTER TABLE ")
	sql.WriteString(tableName)
	sql.WriteString(" DISABLE CONSTRAINT ")
	sql.WriteString(constraintName)
	return sql.String()
}

func (d OracleDialect) EnableConstraintStatement(tableName string, constraintName string) string {
	sql := &strings.Builder{}
	sql.WriteString("ALTER TABLE ")
	sql.WriteString(tableName)
	sql.WriteString(" ENABLE CONSTRAINT ")
	sql.WriteString(constraintName)
	return sql.String()
}

func (d OracleDialect) SupportPreserve() []string {
	return []string{
		string(push.PreserveNothing),
		string(push.PreserveNull),
		string(push.PreserveBlank),
	}
}

// BlankTest implements SQLDialect.
func (d OracleDialect) BlankTest(column string) string {
	return fmt.Sprintf("TRIM(%s) IS NULL", column)
}

func (d OracleDialect) EmptyTest(column string) string {
	return fmt.Sprintf("%s IS NULL", column)
}
