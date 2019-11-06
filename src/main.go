package main

import (
	"awesomeProject/src/common"
	"awesomeProject/src/controller"
	"awesomeProject/src/model"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	model.DB = common.CreateDbConnection(model.DB1Name, model.DbDriver, model.ConnectionOptions)
	model.InitData() //tables init and their data create
	defer model.DB.Close()
	query := fmt.Sprintf("SELECT max(t.id) FROM %s t", model.Table1)
	rows, err := model.DB.Query(query)
	_ = common.TryCatch(err, &common.ErrMsg{"Error from query:\n" + query, nil, common.Default}, false, model.DB)
	if rows.Next() {
		_ = common.TryCatch(rows.Scan(&model.Sequence), &common.ErrMsg{}, false, model.DB)
	}
	r := mux.NewRouter()
	r.HandleFunc("/entity", controller.CreateEntity).Methods("POST")
	r.HandleFunc("/entity", controller.GetEntities).Methods("GET")
	r.HandleFunc("/entity/{id}", controller.GetEntity).Methods("GET")
	r.HandleFunc("/entity/{id}", controller.UpdateEntity).Methods("PUT")
	r.HandleFunc("/entity/{id}", controller.DeleteEntity).Methods("DELETE")
	println("\nStart listener on port" + model.AppPort + " ...\n")
	log.Fatal(http.ListenAndServe(model.AppPort, r))
}
