package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"server/com/log"
	"server/com/util"
	"sync"
	"time"
)

type IDBTask interface {
	Run(*sql.DB)
}

// --------------------------------
type taskMgr struct {
	tasks     chan IDBTask
	Db        *sql.DB
	closeChan chan struct{}
	Cnt       sync.WaitGroup
}

func (m *taskMgr) init(connStr string) error {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Errorf("connect db %s error:%v", connStr, err)
		return err
	}
	m.Db = db
	//	con.db.SetMaxOpenConns(100)
	//	con.db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	err = m.Db.Ping()
	if err != nil {
		log.Errorf("ping db %s error:%v", connStr, err)
		return err
	}

	m.closeChan = make(chan struct{})
	m.tasks = make(chan IDBTask, 1000)
	go m.run(m.Db)

	return nil
}

func (m *taskMgr) Post(t IDBTask) {
	m.Cnt.Add(1)
	m.tasks <- t
}

func safeRun(t IDBTask, conn *sql.DB) {
	defer func() {
		if err := recover(); err != nil {
			util.PrintStack("db task run panic", err)
		}
	}()

	t.Run(conn)
}

func (m *taskMgr) run(conn *sql.DB) {
	defer func() {
		log.Info("db closed")
		conn.Close()
	}()
	for {
		select {
		case t := <-m.tasks:
			safeRun(t, conn)
			m.Cnt.Done()
		case <-m.closeChan:
			return
		}
	}
}

func (m *taskMgr) close() {
	close(m.closeChan)
}

// ---------------------------------------------------------
type DB struct {
	rTask *taskMgr
	wTask *taskMgr
	conns chan *sql.DB
}

func (m *DB) Init(usr, pswd string, host string, dbName string) error {
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",
		usr, pswd, host, dbName, "utf8mb4")
	m.rTask = &taskMgr{}
	err := m.rTask.init(connStr)
	if err != nil {
		return err
	}
	log.Infof("init mysql %s", connStr)
	m.wTask = &taskMgr{}
	err = m.wTask.init(connStr)
	if err != nil {
		return err
	}
	m.createSyncConn(connStr, 1)
	return nil
}

func (m *DB) createSyncConn(connStr string, size int) error {
	m.conns = make(chan *sql.DB, size)

	for i := 0; i < size; i++ {
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			log.Errorf("connect db %s error:%v", connStr, err)
			return err
		}
		m.conns <- db
	}
	return nil
}

func (m *DB) SyncExe(f func(conn *sql.DB)) {
	conn := <-m.conns
	if err := recover(); err != nil {
		util.PrintStack("sync exe panic", err)
	}
	f(conn)
	m.conns <- conn
}

func (m *DB) Read(t IDBTask) {
	m.rTask.Post(t)
}

func (m *DB) Write(t IDBTask) {
	m.wTask.Post(t)
}

// unInit 关闭时需等待任务执行完毕
func (m *DB) Close() {
	m.rTask.close()
	m.wTask.Cnt.Wait()
	m.wTask.close()
}

var mgr DB

func Init(usr, pswd string, host string, dbName string) error {
	return mgr.Init(usr, pswd, host, dbName)
}

func Read(t IDBTask) {
	mgr.Read(t)
}

func Write(t IDBTask) {
	mgr.Write(t)
}

func Close() {
	mgr.Close()
}

func SyncExe(f func(conn *sql.DB)) {
	mgr.SyncExe(f)
}
