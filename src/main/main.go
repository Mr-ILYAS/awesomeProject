package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

const connectionOptions = "host=127.0.0.1 user=postgres password=postgres sslmode=disable "

type Entity struct {
	id   int
	text string
	date time.Time
}

func main() {
	dbName := "go_db1"
	db := createDbConnection(&dbName)
	defer db.Close()
	tableName := "t1"
	table1 := Entity{1, "text", time.Now()}
	initTable(db, &tableName, &table1)

	rows, err := db.Query("SELECT * from " + tableName)
	if err != nil {
		panic(err)
	}

	dbName = "go_db2"
	db = createDbConnection(&dbName)
	defer db.Close()
	tableName = "t2"
	initTable(db, &tableName, nil)

	var table2 Entity
	for rows.Next() {
		err := rows.Scan(&table2.id, &table2.text, &table2.date)
		if err != nil {
			panic(err)
		}
	}
	query := fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", tableName)
	fmt.Println(table2.id, table2.text, table2.date)
	if _, err := db.Exec(query, table2.id, table2.text, table2.date); err != nil {
		println("\nError from query:\n" + query)
		panic(err)
	}
	println(query)
}

func initTable(db *sql.DB, tableName *string, entity *Entity) {
	runQuery(db, fmt.Sprintf("DROP TABLE IF EXISTS %s;", *tableName))
	runQuery(db, fmt.Sprintf("CREATE TABLE %s(id int, text varchar(255), date Timestamp);", *tableName))

	if entity != nil {
		//hard "runQuery" with params:
		query := fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", *tableName)
		_, err := db.Exec(query, entity.id, entity.text, entity.date)
		if err != nil {
			println("\nError from query:\n" + query)
			panic(err)
		}
		println(query)
	}
	println()
}

func runQuery(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		println("\nError from query:\n" + query)
		panic(err)
	}
	println(query)
}

func createDbConnection(dbName *string) *sql.DB {
	if dbName != nil {
		*dbName = fmt.Sprintf("dbname=%s", *dbName)
	} else {
		*dbName = ""
	}
	db, err := sql.Open("postgres", fmt.Sprintf(connectionOptions+"%s", *dbName))
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	return db
}
