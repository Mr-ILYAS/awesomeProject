package common

import (
	"database/sql"
	"fmt"
)

type ErrMsg struct {
	Message      string
	Args         interface{}
	ArgsViewType ArgsViewType
}

type ArgsViewType int

func (viewType ArgsViewType) Get() string {
	return [...]string{"", "+", "#"}[viewType]
}

const (
	Default ArgsViewType = iota
	WithStructNames
	Representation
)

func TryCatch(err error, errMessage *ErrMsg, throwUp bool, toClose ...*sql.DB) error {
	if err != nil {
		if errMessage.Message != "" {
			message := errMessage.Message
			if errMessage.Args != nil {
				message = fmt.Sprintf(message+"\t\targs: %"+errMessage.ArgsViewType.Get()+"v", errMessage.Args)
			}
			println(message)
		}
		println()
		if !throwUp {
			for _, resource := range toClose {
				resource.Close()
			}
			panic(err)
		}
		return err
	}
	return nil
}

func RunQuery(db *sql.DB, query string, throwUp bool, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	err = TryCatch(err, &ErrMsg{"Error from query:\n" + query, args, Representation}, throwUp, db)
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

func CreateDbConnection(dbName string, driver string, connectionOptions string) *sql.DB {
	dbName = fmt.Sprintf("dbname=%s", dbName)
	db, err := sql.Open(driver, fmt.Sprintf(connectionOptions+" %s", dbName))
	_ = TryCatch(err, &ErrMsg{}, false)
	_ = TryCatch(db.Ping(), &ErrMsg{}, false, db)
	return db
}
