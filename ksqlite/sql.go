// sql
package ksqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func GetStmt(fileName string, query string) (Stmt *sql.Stmt, err error) {
	var db *sql.DB
	if db, err = sql.Open("sqlite3", fileName); err != nil {
		err = fmt.Errorf("ksqlite.GetStmt: open db err= %v", err.Error())
		return
	}
	//if Stmt, err = db.Prepare("insert into t_32_48 (DevID, sec, sat, cred, lat, lon) values(?, ?, ?, ?, ?, ?)"); err != nil {
	if Stmt, err = db.Prepare(query); err != nil {
		err = fmt.Errorf("Create_32_48_Stmt: prepare err= %v", err.Error())
		Stmt = nil
		return
	}
	return
}
