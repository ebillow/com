package orm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// IOrmTask 使用gorm的异步任务接口
type IOrmTask interface {
	Run(*gorm.DB)
}

// Orm gorm封装
type Orm struct {
	tasks    chan IOrmTask
	syncSess chan *gorm.DB
	sess     *gorm.DB
	cfg      *DbCfg

	closeChan chan struct{}
}

// Init 初始化
const charset = "utf8mb4"

func (m *Orm) Init(cfg *DbCfg, asyncTaskSize int32, syncSessCnt int32, migrate func(*gorm.DB) error) bool {
	m.cfg = cfg
	conStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True",
		cfg.Usr, cfg.Pswd, cfg.Host, "", charset)
	log.Infof("connStr: %s", conStr)
	sess, err := gorm.Open(mysql.Open(conStr), &gorm.Config{})
	if err != nil {
		log.Panicf("connect db %s error:%v", conStr, err)
		return false
	}

	sess = sess.Set("gorm:db_options", "CHARSET="+charset)

	err = sess.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci ", cfg.Db)).Error
	if err != nil {
		return false
	}

	err = sess.Exec(fmt.Sprintf("use %s", cfg.Db)).Error
	if err != nil {
		return false
	}

	log.Info("start init callback")
	err = migrate(sess)
	if err != nil {
		log.Panicf("migrate error:%v", err)
		return false
	}

	conStrWithDb := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True",
		cfg.Usr, cfg.Pswd, cfg.Host, cfg.Db, charset)

	m.sess = sess
	m.tasks = make(chan IOrmTask, asyncTaskSize)
	m.closeChan = make(chan struct{})
	m.syncSess = make(chan *gorm.DB, syncSessCnt)

	go m.run(m.sess)

	for i := 0; i < cap(m.syncSess); i++ {
		sess, err := gorm.Open(mysql.Open(conStr), &gorm.Config{})
		if err != nil {
			log.Panicf("connect db %s error:%v", conStrWithDb, err)
			return false
		}
		m.syncSess <- sess
	}

	log.Infof("db init to :%v", m.cfg)
	return true
}

// Close 关闭
func (m *Orm) Close() {
	log.Info("gorm close")
	close(m.closeChan)
}

// Post 投递异步任务
func (m *Orm) Post(t IOrmTask) {
	m.tasks <- t
}

// Execute 执行同步任务
func (m *Orm) Execute(f func(sess *gorm.DB) error) error {
	sess := <-m.syncSess
	defer func() {
		m.syncSess <- sess
	}()

	return f(sess)
}

func (m *Orm) run(conn *gorm.DB) {
	defer func() {
		conn.Close()
		for v := range m.syncSess {
			v.Close()
		}
	}()

	for {
		select {
		case t := <-m.tasks:
			t.Run(conn)
		case <-m.closeChan:
			return
		}
	}
}
