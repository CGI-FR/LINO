package sequence

import (
	"fmt"

	// import Oracle connector
	_ "github.com/godror/godror"

	"github.com/cgi-fr/lino/pkg/sequence"
)

// NewOracleExtractorFactory creates a new oracle extractor factory.
func NewOracleUpdatorFactory() *OracleUpdatorFactory {
	return &OracleUpdatorFactory{}
}

// OracleExtractorFactory exposes methods to create new Oracle extractors.
type OracleUpdatorFactory struct{}

// New return a Oracle extractor
func (e *OracleUpdatorFactory) New(url string, schema string) sequence.Updator {
	return NewSQLUpdator(url, schema, OracleDialect{})
}

type OracleDialect struct{}

func (d OracleDialect) SequencesSQL(schema string) string {
	if schema == "" {
		return fmt.Sprintf("SELECT sequence_name FROM user_sequences")
	} else {
		return fmt.Sprintf("SELECT sequence_name FROM dba_sequences WHERE sequence_owner='%s'", schema)
	}
}

func (d OracleDialect) UpdateSequenceSQL(schema string, sequence string, tableName string, column string) string {
	if schema != "" {
		tableName = schema + "." + tableName
		sequence = schema + "." + sequence
	}
	return fmt.Sprintf(`
			DECLARE
				last_val NUMBER;
				next_val NUMBER;
			BEGIN
				SELECT MAX(%s) INTO next_val FROM %s;
				IF next_val > 0 THEN
					SELECT %s.nextval INTO last_val FROM DUAL;
					EXECUTE IMMEDIATE 'ALTER SEQUENCE %s INCREMENT BY -' || last_val || ' MINVALUE 0';
					SELECT %s.nextval INTO last_val FROM DUAL;
					EXECUTE IMMEDIATE 'ALTER SEQUENCE %s INCREMENT BY ' || next_val;
					SELECT %s.nextval INTO last_val FROM DUAL;
					EXECUTE IMMEDIATE 'ALTER SEQUENCE %s INCREMENT BY 1 MINVALUE 1';
				END IF;
			END;`,
		column, tableName, sequence, sequence, sequence, sequence, sequence, sequence)
}

func (d OracleDialect) StatusSequenceSQL(schema string, sequence string) string {
	if schema == "" {
		return fmt.Sprintf("SELECT last_number FROM user_sequences WHERE sequence_name = '%s'", sequence)
	} else {
		return fmt.Sprintf(
			"SELECT last_number FROM dba_sequences WHERE sequence_name = '%s' AND sequence_owner='%s'",
			sequence, schema,
		)
	}
}
