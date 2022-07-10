package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB
var sensitiveKeys = map[string]bool{"aws_access_key_id": true, "aws_secret_access_key": true}

func initDB(connStr string) {

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	sqlTable := `
	CREATE TABLE IF NOT EXISTS guilds (
		id SERIAL PRIMARY KEY,
		guild_id TEXT NOT NULL,
		region TEXT NOT NULL,
		aws_access_key_id TEXT,
		aws_secret_access_key TEXT,
		UNIQUE (guild_id, region)
	);`

	_, err = db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}

}

func queryDB(args map[string]interface{}) []map[string]interface{} {

	sqlWhere := make([]string, 0, len(args))
	sqlArgs := make([]interface{}, 0, len(args))
	i := 1
	for k, v := range args {
		sqlWhere = append(sqlWhere, fmt.Sprintf("%s = $%s", k, strconv.Itoa(i)))
		sqlArgs = append(sqlArgs, v)
		i += 1
	}

	sqlStatement := fmt.Sprintf("SELECT * FROM guilds WHERE %s;", strings.Join(sqlWhere, " AND "))

	rows, _ := db.Query(sqlStatement, sqlArgs...)
	if rows == nil {
		return nil
	}
	cols, _ := rows.Columns()

	defer rows.Close()
	var data []map[string]interface{}

	for rows.Next() {
		entry := make(map[string]interface{})
		columns := make([]string, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		rows.Scan(columnPointers...)

		for i, colName := range cols {
			entry[colName] = columns[i]
		}
		data = append(data, entry)
	}

	return data
}

func insertDB(args map[string]interface{}) error {

	sqlCols := make([]string, 0, len(args))
	sqlVals := make([]string, 0, len(args))
	sqlArgs := make([]interface{}, 0, len(args))

	i := 1
	for k, v := range args {
		sqlCols = append(sqlCols, k)
		sqlVals = append(sqlVals, "$"+strconv.Itoa(i))
		sqlArgs = append(sqlArgs, v)
		i += 1
	}

	sqlStatement := fmt.Sprintf("INSERT INTO guilds (%s) VALUES (%s) RETURNING id;", strings.Join(sqlCols, ", "), strings.Join(sqlVals, ", "))

	var id int
	err := db.QueryRow(sqlStatement, sqlArgs...).Scan(&id)
	if err != nil {
		log.Println(err)
	}

	return err
}

func deleteDB(args map[string]interface{}) {

	sqlWhere := make([]string, 0, len(args))
	sqlArgs := make([]interface{}, 0, len(args))

	i := 1
	for k, v := range args {
		sqlWhere = append(sqlWhere, fmt.Sprintf("%s = $%s", k, strconv.Itoa(i)))
		sqlArgs = append(sqlArgs, v)
		i += 1
	}

	sqlStatement := fmt.Sprintf("DELETE FROM guilds WHERE %s;", strings.Join(sqlWhere, " AND "))

	res, err := db.Exec(sqlStatement, sqlArgs...)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Result", res)
}

func saveCredsToDB(args map[string]interface{}) error {
	encryptMap := make(map[string]interface{}, len(args))

	for k, v := range args {
		switch {
		case sensitiveKeys[k]:
			encryptMap[k] = encrypt(v.(string))
		default:
			encryptMap[k] = v
		}
	}
	err := insertDB(encryptMap)

	return err
}

func getCredsFromDB(args map[string]interface{}) []map[string]interface{} {
	queryMap := make(map[string]interface{})
	for k, v := range args {
		if k == "guild_id" || k == "region" {
			queryMap[k] = v
		}
	}
	data := queryDB(queryMap)

	returnData := make([]map[string]interface{}, len(data))

	for _, d := range data {
		decryptMap := make(map[string]interface{}, len(d))

		for k, v := range d {
			switch {
			case sensitiveKeys[k]:
				decryptMap[k] = decrypt(v.(string))
			default:
				decryptMap[k] = v
			}
		}
		returnData = append(returnData, decryptMap)
	}

	return returnData
}
