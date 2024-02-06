package table

import (
	"database/sql"
	"fmt"
)

func openDB() {
	db, err := sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable host=source port=5432 password=sakila")
	if err != nil {
		fmt.Println("Open DB: ", err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("Ping: ", err.Error())
	}

	tableName := "film"
	rows, err := db.Query(fmt.Sprintf("SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = '%s'", tableName))
	defer rows.Close()
	fmt.Printf("Informations sur les colonnes de la table '%s':\n", tableName)
	for rows.Next() {
		var columnName, dataType string
		var characterMaxLength, numericPrecision, numericScale sql.NullInt64

		err := rows.Scan(&columnName, &dataType, &characterMaxLength)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Column: %s, Type: %s", columnName, dataType)

		// Afficher la longueur si elle est disponible
		if characterMaxLength.Valid {
			fmt.Printf(", Length: %d", characterMaxLength.Int64)
		}

		// Afficher la précision et l'échelle si elles sont disponibles
		if numericPrecision.Valid && numericScale.Valid {
			fmt.Printf(", Precision: %d, Scale: %d", numericPrecision.Int64, numericScale.Int64)
		}

		fmt.Println()
	}
}

// func printRows(rows *sql.Rows) {
// 	colTypes, err := rows.ColumnTypes()
// 	if err != nil {
// 		fmt.Println("Get Col Info: ", err.Error())
// 	}
// 	for _, s := range colTypes {
// 		fmt.Println(s)
// 		// fmt.Println("Col", s.Name(), "est type:", s.DatabaseTypeName())
// 		// length, err := s.Length()
// 		// if err {
// 		// 	fmt.Println(s.DatabaseTypeName(), "N'a Pas de length")
// 		// } else {
// 		// 	fmt.Println("et length est: ", length)
// 		// }
// 	}
// }
