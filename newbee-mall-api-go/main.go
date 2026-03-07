package main

import (
	"main.go/core"
	"main.go/global"
	"main.go/initialize"
	"main.go/pkg/kafka"
	"main.go/pkg/redis"
	"main.go/service/mall"
)

func main() {

	global.GVA_VP = core.Viper()      // 初始化Viper
	global.GVA_LOG = core.Zap()       // 初始化zap日志库
	global.GVA_DB = initialize.Gorm() // gorm连接数据库
	redis.InitRedis(global.GVA_CONFIG.Redis.Addr, global.GVA_CONFIG.Redis.Password, global.GVA_CONFIG.Redis.DB)
	kafka.InitProducer(global.GVA_CONFIG.Kafka.Addr, global.GVA_CONFIG.Kafka.Topic)

	// 启动 Kafka 订单消费者
	go mall.RunOrderConsumer()

	core.RunWindowsServer()
}
