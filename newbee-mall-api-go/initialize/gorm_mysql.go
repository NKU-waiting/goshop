package initialize

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"main.go/global"
	"main.go/initialize/internal"
)

func GormMysql() *gorm.DB {
	m := global.GVA_CONFIG.Mysql
	if m.Dbname == "" {
		global.GVA_LOG.Error("Database name is empty")
		return nil
	}
	dsn := m.Dsn()
	global.GVA_LOG.Info("Connecting to database with DSN: " + dsn)
	mysqlConfig := mysql.Config{
		DSN:                       dsn, // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config()); err != nil {
		global.GVA_LOG.Error("Failed to connect to database: " + err.Error())
		return nil
	} else {
		global.GVA_LOG.Info("Database connected successfully")
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}
