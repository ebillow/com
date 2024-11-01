package orm

import "gorm.io/gorm"

var orm Orm

type DbCfg struct {
	Driver string
	Usr    string
	Pswd   string
	Host   string
	Db     string
}

// Init 初始化
func Init(cfg *DbCfg, asyncTaskSize int32, syncSessCnt int32, migrate func(*gorm.DB) error) bool {
	return orm.Init(cfg, asyncTaskSize, syncSessCnt, migrate)
}

// Close 关闭
func Close() {
	orm.Close()
}

// Post 投递异步任务
func Post(t IOrmTask) {
	orm.Post(t)
}

// Execute 执行同步任务
func Execute(f func(sess *gorm.DB) error) error {
	return orm.Execute(f)
}
