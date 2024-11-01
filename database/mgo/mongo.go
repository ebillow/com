package mgo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"server/com/log"
	"server/com/util"
	"sync"
	"time"
)

type ITask interface {
	Run(db *mongo.Database)
}

func createClient(connStr string) (cli *mongo.Client, err error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Panicf("init mongodb %s err:%v", connStr, err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Panicf("connect mongodb %s err:%v", connStr, err)
		return nil, err
	}

	//ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	//err = client.Ping(ctx, nil)
	//if err != nil {
	//	log.Errorf("ping mongo %s error:%v", connStr, err)
	//	return nil, err
	//}
	return client, nil
}

// --------------------------------
type taskMgr struct {
	tasks     chan ITask
	closeChan chan struct{}
	Cnt       sync.WaitGroup
	client    *mongo.Client
}

func (m *taskMgr) init(connStr string, dbName string) error {
	var err error
	m.client, err = createClient(connStr)
	if err != nil {
		return err
	}
	m.closeChan = make(chan struct{})
	m.tasks = make(chan ITask, 10000)
	go m.run(dbName)

	return nil
}

func (m *taskMgr) Post(t ITask) {
	m.Cnt.Add(1)
	m.tasks <- t
}

func safeRun(t ITask, db *mongo.Database) {
	defer func() {
		if err := recover(); err != nil {
			util.PrintStack("mongo task run panic err:%v", err)
		}
	}()

	t.Run(db)
}

func (m *taskMgr) run(dbName string) {
	defer func() {
		log.Info("db closed")
		_ = m.client.Disconnect(context.Background())
	}()

	db := m.client.Database(dbName)
	for {
		select {
		case t := <-m.tasks:
			safeRun(t, db)
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
	db    chan *mongo.Database
}

func (m *DB) Init(addrs string, usr string, password, dbName string, size int, options string) error {
	//mongodb://[username:password@]host1[:port1][,host2[:port2],…[,hostN[:portN]]][/[database][?options]]
	//mongodb://dcjt:dcjt123dalgurak@127.0.0.1:27017/admin
	//mongodb://rwuser:<password>@192.168.xx.xx:8635,192.168.xx.xx:8635/test?authSource=admin
	//aws
	//readPreference           = "secondaryPreferred"
	//connectionStringTemplate = "mongodb://%s:%s@%s/sample-database?replicaSet=rs0&readpreference=%s"
	//connectionURI := fmt.Sprintf(connectionStringTemplate, username, password, clusterEndpoint, readPreference)

	up := ""
	if usr != "" {
		up = usr + ":" + password + "@"
	}
	connStr := fmt.Sprintf("mongodb://%s%s/admin?%s", up, addrs, options)
	//}
	log.Infof("connStr:%s", connStr)
	m.rTask = &taskMgr{}
	err := m.rTask.init(connStr, dbName)
	if err != nil {
		return err
	}
	m.wTask = &taskMgr{}
	err = m.wTask.init(connStr, dbName)
	if err != nil {
		return err
	}
	m.createSyncConn(connStr, size, dbName)
	return nil
}

func (m *DB) createSyncConn(connStr string, size int, dbName string) error {
	m.db = make(chan *mongo.Database, size)

	for i := 0; i < size; i++ {
		cli, err := createClient(connStr)
		if err != nil {
			return err
		}
		m.db <- cli.Database(dbName)
	}
	return nil
}

func (m *DB) SyncExe(f func(db *mongo.Database)) {
	db := <-m.db
	defer func() {
		if err := recover(); err != nil {
			util.PrintStack("sync exe panic", err)
		}
	}()
	f(db)
	m.db <- db
}

func (m *DB) Read(t ITask) {
	m.rTask.Post(t)
}

func (m *DB) Write(t ITask) {
	m.wTask.Post(t)
}

func (m *DB) Close() {
	if m.rTask != nil {
		m.rTask.close()
	}
	if m.wTask != nil {
		m.wTask.Cnt.Wait()
		m.wTask.close()
	}
}

var mgr DB

func Init(addrs string, usr string, password string, dbName string, syncSize int, options string) error {
	return mgr.Init(addrs, usr, password, dbName, syncSize, options)
}

func Read(t ITask) {
	mgr.Read(t)
}

func Write(t ITask) {
	mgr.Write(t)
}

func Close() {
	mgr.Close()
}

func SyncExe(f func(cli *mongo.Database)) {
	mgr.SyncExe(f)
}
