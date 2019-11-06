package controller

import (
	"awesomeProject/src/common"
	"awesomeProject/src/model"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func GetEntities(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	query := "SELECT * FROM t1"
	rows, err := model.DB.Query(query)
	errMsg := ""
	if err = common.TryCatch(err, &common.ErrMsg{"Error from query:\n" + query, nil, common.Default}, true, model.DB); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entity model.Entity
	var entities []model.Entity
	for rows.Next() {
		if err = common.TryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), &common.ErrMsg{}, true, model.DB); err != nil {
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

func GetEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := model.DB.Query(query)
	errMsg := ""
	if err = common.TryCatch(err, &common.ErrMsg{"Error from query:\n" + query, nil, common.Default}, true, model.DB); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entity model.Entity
	if rows.Next() {
		if err = common.TryCatch(rows.Scan(&entity.Id, &entity.Text, &entity.Date), &common.ErrMsg{}, true, model.DB); err != nil {
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

func CreateEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	errMsg := ""
	var entity model.Entity
	if err := json.NewDecoder(request.Body).Decode(&entity); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusBadRequest)
		println(errMsg)
		return
	}
	entity.Id = *model.NextId()
	entity.Date = time.Now().String()
	if err := common.RunQuery(model.DB, "INSERT INTO t1(id, text, date) VALUES($1, $2, $3);", true, entity.Id, entity.Text, entity.Date); err != nil {
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

func UpdateEntity(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := model.DB.Query(query)
	errMsg := ""
	if err = common.TryCatch(err, &common.ErrMsg{"Error from query:\n" + query, nil, common.Default}, true, model.DB); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	var entityOld model.Entity
	var entity model.Entity
	if err := json.NewDecoder(request.Body).Decode(&entity); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusBadRequest)
		println(errMsg)
		return
	}
	entity.Id = int(id)
	if rows.Next() { //Обновить, если существует (дата создания не обновляется)
		if err = common.TryCatch(rows.Scan(&entityOld.Id, &entityOld.Text, &entityOld.Date), &common.ErrMsg{}, true, model.DB); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
		fmt.Printf("%#v\n", entityOld)
		entity.Date = entityOld.Date
		if err := common.RunQuery(model.DB, "UPDATE t1 SET text=$2 where id=$1;", true, id, entity.Text); err != nil {
			errMsg = err.Error()
			http.Error(writer, errMsg, http.StatusInternalServerError)
			println(errMsg)
			return
		}
	} else { //Иначе создать новую запись
		entity.Date = time.Now().String()
		if err := common.RunQuery(model.DB, "INSERT INTO t1(id, text, date) VALUES($1, $2, $3);", true, entity.Id, entity.Text, entity.Date); err != nil {
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

func DeleteEntity(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(request)["id"], 10, 0)
	query := fmt.Sprintf("SELECT * FROM t1 WHERE id=%d", id)
	rows, err := model.DB.Query(query)
	errMsg := ""
	if err = common.TryCatch(err, &common.ErrMsg{"Error from query:\n" + query, nil, common.Default}, true, model.DB); err != nil {
		errMsg = err.Error()
		http.Error(writer, errMsg, http.StatusInternalServerError)
		println(errMsg)
		return
	}
	fmt.Println(query)
	if rows.Next() {
		if err := common.RunQuery(model.DB, "DELETE FROM t1 WHERE id=$1;", true, id); err != nil {
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
