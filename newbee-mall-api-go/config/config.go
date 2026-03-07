package config

// Server 结构体聚合了所有配置
type Server struct {
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	System System `mapstructure:"system" json:"system" yaml:"system"`
	Zap    Zap    `mapstructure:"zap" json:"zap" yaml:"zap"`
	Local  Local  `mapstructure:"local" json:"local" yaml:"local"`
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis"`
	Kafka  Kafka  `mapstructure:"kafka" json:"kafka" yaml:"kafka"`
}

// Redis 新增的 Redis 配置结构体
type Redis struct {
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`
}

// Kafka 新增的 Kafka 配置结构体
type Kafka struct {
	Addr  string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Topic string `mapstructure:"topic" json:"topic" yaml:"topic"`
	Group string `mapstructure:"group" json:"group" yaml:"group"`
}
