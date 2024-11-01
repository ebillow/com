package db

import (
	"database/sql"
	"fmt"
	"server/com/database/orm"
	"testing"
)

var cfg = &orm.DbCfg{
	Driver: "mysql",
	Usr:    "root",
	Pswd:   "123",
	Host:   "192.168.0.126:3306",
	Db:     "",
}

func TestDB_Init(t *testing.T) {
	conStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
		cfg.Usr, cfg.Pswd, cfg.Host, cfg.Db)
	c, err := sql.Open(cfg.Driver, conStr)
	if err != nil {
		t.Error("conn db err:v", err)
	}
	r, err := c.Exec("CREATE DATABASE  IF NOT EXISTS game")
	if err != nil {
		t.Error(err)
	}
	r, err = c.Exec("use yw")
	if err != nil {
		t.Error(err)
	}
	r, err = c.Exec("show tables")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(r)

	stmt, err := c.Prepare("INSERT INTO log_login VALUES (?,?,?,?,?,?),(?,?,?,?,?,?)")
	if err != nil {
		t.Error(err)
	}
	_, err = stmt.Exec("111333s", "name", "accc", "aaa", "ip", "timeaa", "2223333666", "name1", "accc2", "aaa2", "ip2", "timeaa2")
	if err != nil {
		t.Error(err)
	}
}
