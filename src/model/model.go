package model

import (
	"awesomeProject/src/common"
	"database/sql"
	"fmt"
	"time"
)

const (
	ConnectionOptions = "host=127.0.0.1 user=postgres password=postgres sslmode=disable"
	DB1Name, DB2Name  = "go_db1", "go_db2"
	DbDriver          = "postgres"
	AppPort           = ":8000"
)

var DB *sql.DB

type Entity struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
	Date string `json:"date"`
}

var Sequence int

func NextId() *int {
	*&Sequence++
	return &Sequence
}

func InitData() {
	tableName := "public.t1"
	table1 := []Entity{
		{*NextId(), "text1", time.Now().String()},
		{*NextId(), "text2", time.Now().Add(time.Hour * 24).String()},
	}
	InitTable(DB, &tableName, &table1)

	rows, err := DB.Query("SELECT * FROM " + tableName)
	_ = common.TryCatch(err, &common.ErrMsg{}, false, DB)

	db2 := common.CreateDbConnection(DB2Name, DbDriver, ConnectionOptions)
	defer db2.Close()
	tableName = "public.t2"
	InitTable(db2, &tableName, nil)

	var entity Entity
	for rows.Next() {
		_ = common.TryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), &common.ErrMsg{}, false, DB, db2)
		_ = common.RunQuery(db2, fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", tableName), false, entity.Id, entity.Text, entity.Date)
	}
}

func InitTable(db *sql.DB, tableName *string, entities *[]Entity) {
	_ = common.RunQuery(db, fmt.Sprintf("DROP TABLE IF EXISTS %s;", *tableName), false)
	_ = common.RunQuery(db, fmt.Sprintf("CREATE TABLE %s(id int, text VARCHAR(255), date VARCHAR(60));", *tableName), false)
	if entities != nil {
		for _, entity := range *entities {
			_ = common.RunQuery(db, fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", *tableName), false, entity.Id, entity.Text, entity.Date)
		}
	}
	fmt.Println()
}
