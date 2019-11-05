package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	connectionOptions = "host=127.0.0.1 user=postgres password=postgres sslmode=disable"
	db1Name, db2Name  = "go_db1", "go_db2"
	dbDriver          = "postgres"
	appPort           = ":8000"
)

type Entity struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
	Date string `json:"date"`
}

var Sequence int

func NextId() int {
	Sequence++
	return Sequence
}
func InitData() {
	db1 := createDbConnection(db1Name)
	defer db1.Close()
	tableName := "public.t1"
	table1 := []Entity{
		{NextId(), "text1", time.Now().String()},
		{NextId(), "text2", time.Now().Add(time.Hour * 24).String()},
	}
	initTable(db1, &tableName, &table1)

	rows, err := db1.Query("SELECT * FROM " + tableName)
	_ = tryCatch(err, ErrMsg{}, false, db1)

	db2 := createDbConnection(db2Name)
	defer db2.Close()
	tableName = "public.t2"
	initTable(createDbConnection(db2Name), &tableName, nil)

	var entity Entity
	for rows.Next() {
		_ = tryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), ErrMsg{}, false, db1, db2)
		_ = runQuery(db2, fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", tableName), false, entity.Id, entity.Text, entity.Date)
	}
}

func main() {
	//InitData() //tables init and their data create
	db := createDbConnection(db1Name)
	defer db.Close()
	query := "SELECT max(t1.id) FROM t1"
	rows, err := db.Query(query)
	_ = tryCatch(err, ErrMsg{"Error from query:\n" + query, nil, Default}, false, db)
	if rows.Next() {
		_ = tryCatch(rows.Scan(&Sequence), ErrMsg{}, false, db)
	}
	r := mux.NewRouter()
	r.HandleFunc("/entity", createEntity).Methods("POST")
	r.HandleFunc("/entity", getEntities).Methods("GET")
	r.HandleFunc("/entity/{id}", getEntity).Methods("GET")
	r.HandleFunc("/entity/{id}", updateEntity).Methods("PUT")
	r.HandleFunc("/entity/{id}", deleteEntity).Methods("DELETE")
	println("\nStart listener on port" + appPort + " ...\n")
	log.Fatal(http.ListenAndServe(appPort, r))
}

func initTable(db *sql.DB, tableName *string, entities *[]Entity) {
	_ = runQuery(db, fmt.Sprintf("DROP TABLE IF EXISTS %s;", *tableName), false)
	_ = runQuery(db, fmt.Sprintf("CREATE TABLE %s(id int, text VARCHAR(255), date VARCHAR(60));", *tableName), false)
	if entities != nil {
		for _, entity := range *entities {
			_ = runQuery(db, fmt.Sprintf("INSERT INTO %s(id, text, date) VALUES($1, $2, $3);", *tableName), false, entity.Id, entity.Text, entity.Date)
		}
	}
	fmt.Println()
}

func runQuery(db *sql.DB, query string, throwUp bool, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	err = tryCatch(err, ErrMsg{"Error from query:\n" + query, args, Representation}, throwUp, db)
	if err == nil {
		if args != nil {
			query = fmt.Sprintf(query+"\t←—→\t %v", args)
		}
		fmt.Println(query)
		return nil
	} else {
		return err
	}
}

func createDbConnection(dbName string) *sql.DB {
	dbName = fmt.Sprintf("dbname=%s", dbName)
	db, err := sql.Open(dbDriver, fmt.Sprintf(connectionOptions+" %s", dbName))
	_ = tryCatch(err, ErrMsg{}, false, db)
	_ = tryCatch(db.Ping(), ErrMsg{}, false, db)
	return db
}

func getEntities(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	db := createDbConnection(db1Name)
	defer db.Close()
	query := "SELECT * FROM t1"
	rows, err := db.Query(query)
	errMsg := ""
	if err = tryCatch(err, ErrMsg{"Error from query:\n" + query, nil, Default}, true, db); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entity Entity
	var entities []Entity
	for rows.Next() {
		if err = tryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), ErrMsg{}, true, db); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
		entities = append(entities, entity)
		fmt.Printf("%#v\n", entity)
	}
	jsonEntities, err := json.Marshal(entities)
	if err != nil {
		errMsg = "Error on marshal entities."
		http.Error(writer, errMsg, http.StatusInternalServerError)
	} else {
		jsonResponse := append([]byte{'d', 'a', 't', 'a', ':'}, jsonEntities...)
		_, _ = writer.Write(jsonResponse)
	}
}

func getEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	db := createDbConnection(db1Name)
	defer db.Close()
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := db.Query(query)
	errMsg := ""
	if err = tryCatch(err, ErrMsg{"Error from query:\n" + query, nil, Default}, true, db); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entity Entity
	if rows.Next() {
		if err = tryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), ErrMsg{}, true, db); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
		fmt.Printf("%#v\n", entity)
		jsonEntity, err := json.Marshal(entity)
		if err != nil {
			http.Error(writer, "Error on marshal entity.", http.StatusInternalServerError)
		} else {
			jsonResponse := append([]byte{'d', 'a', 't', 'a', ':'}, jsonEntity...)
			_, _ = writer.Write(jsonResponse)
		}
	} else {
		errMsg = "Bad request. Invalid parameter: id"
		http.Error(writer, errMsg, http.StatusBadRequest)
	}
}

//TODO: transactional ↓

func createEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	errMsg := ""
	var entity Entity
	if err := json.NewDecoder(request.Body).Decode(&entity); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusBadRequest)
		println(errMsg)
		return
	}
	entity.Id = NextId()
	entity.Date = time.Now().String()
	db := createDbConnection(db1Name)
	defer db.Close()
	if err := runQuery(db, "INSERT INTO t1(id, text, date) VALUES($1, $2, $3);", true, entity.Id, entity.Text, entity.Date); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	jsonEntity, err := json.Marshal(entity)
	if err != nil {
		errMsg = "Error on marshal entity."
		http.Error(writer, errMsg, http.StatusInternalServerError)
	} else {
		jsonResponse := append([]byte{'d', 'a', 't', 'a', ':'}, jsonEntity...)
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write(jsonResponse)
	}
}

func updateEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	db := createDbConnection(db1Name)
	defer db.Close()
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := db.Query(query)
	errMsg := ""
	if err = tryCatch(err, ErrMsg{"Error from query:\n" + query, nil, Default}, true, db); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entityOld Entity
	var entity Entity
	if err := json.NewDecoder(request.Body).Decode(&entity); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusBadRequest)
		println(errMsg)
		return
	}
	entity.Id = int(id)
	if rows.Next() { //Обновить, если существует (дата создания не обновляется)
		if err = tryCatch(rows.Scan(&entityOld.Id, &entityOld.Text, &entityOld.Date), ErrMsg{}, true, db); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
		fmt.Printf("%#v\n", entityOld)
		entity.Date = entityOld.Date
		if err := runQuery(db, "UPDATE t1 SET text=$2 where id=$1;", true, id, entity.Text); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
	} else { //Иначе создать новую запись
		entity.Date = time.Now().String()
		if err := runQuery(db, "INSERT INTO t1(id, text, date) VALUES($1, $2, $3);", true, entity.Id, entity.Text, entity.Date); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
	}
	fmt.Printf("%#v\n", entity)
	jsonEntity, err := json.Marshal(entity)
	if err != nil {
		http.Error(writer, errMsg, http.StatusInternalServerError)
	} else {
		jsonResponse := append([]byte{'d', 'a', 't', 'a', ':'}, jsonEntity...)
		_, _ = writer.Write(jsonResponse)
	}
}

func deleteEntity(writer http.ResponseWriter, request *http.Request) {
	db := createDbConnection(db1Name)
	defer db.Close()
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := db.Query(query)
	errMsg := ""
	if err = tryCatch(err, ErrMsg{"Error from query:\n" + query, nil, Default}, true, db); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	if rows.Next() {
		if err := runQuery(db, "DELETE FROM t1 WHERE id=$1;", true, id); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	} else {
		errMsg = "Bad request. Invalid parameter: id"
		http.Error(writer, errMsg, http.StatusBadRequest)
	}
}

func tryCatch(err error, errMessage ErrMsg, throwUp bool, toClose ...*sql.DB) error {
	if err != nil {
		for _, resource := range toClose {
			resource.Close()
		}
		if errMessage.message != "" {
			message := errMessage.message
			if errMessage.args != nil {
				message = fmt.Sprintf(message+"\t\targs: %"+errMessage.argsViewType.get()+"v", errMessage.args)
			}
			println(message)
		}
		println()
		if !throwUp {
			panic(err)
		}
		return err
	}
	return nil
}

type ErrMsg struct {
	message      string
	args         interface{}
	argsViewType ArgsViewType
}

type ArgsViewType int

func (viewType ArgsViewType) get() string {
	return [...]string{"", "+", "#"}[viewType]
}

const (
	Default ArgsViewType = iota
	WithStructNames
	Representation
)
