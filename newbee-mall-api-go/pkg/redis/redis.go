package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

// InitRedis 初始化 Redis 客户端
func InitRedis(addr, password string, db int) {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,     // Redis 地址
		Password: password, // Redis 密码，没有则为空
		DB:       db,       // 默认数据库
		PoolSize: 10,       // 连接池大小
		PoolTimeout: 5 * time.Second, // 连接池超时时间
	})

	// 通过向 Redis 发送 PING 命令来检查连接
	pong, err := Client.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
	fmt.Println("Redis connected successfully:", pong)
}

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(Ctx, key, value, expiration).Err()
}

// Get 获取键值
func Get(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

// Del 删除键
func Del(key string) error {
	return Client.Del(Ctx, key).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	count, err := Client.Exists(Ctx, key).Result()
	return count > 0, err
}

// Incrby 对指定键的值进行原子性递增
func Incrby(key string, value int64) (int64, error) {
	return Client.IncrBy(Ctx, key, value).Result()
}

// Decrby 对指定键的值进行原子性递减
func Decrby(key string, value int64) (int64, error) {
	return Client.DecrBy(Ctx, key, value).Result()
}

// HSet 设置哈希表中字段的值
func HSet(key, field string, value interface{}) error {
	return Client.HSet(Ctx, key, field, value).Err()
}

// HGet 获取哈希表中字段的值
func HGet(key, field string) (string, error) {
	return Client.HGet(Ctx, key, field).Result()
}

// HGetAll 获取哈希表中所有字段和值
func HGetAll(key string) (map[string]string, error) {
	return Client.HGetAll(Ctx, key).Result()
}

// LPush 将一个或多个值插入到列表头部
func LPush(key string, values ...interface{}) error {
	return Client.LPush(Ctx, key, values...).Err()
}

// RPop 移除并返回列表的最后一个元素
func RPop(key string) (string, error) {
	return Client.RPop(Ctx, key).Result()
}

// SAdd 将一个或多个成员元素加入到集合中
func SAdd(key string, members ...interface{}) error {
	return Client.SAdd(Ctx, key, members...).Err()
}

// SIsMember 判断成员元素是否是集合的成员
func SIsMember(key string, member interface{}) (bool, error) {
	return Client.SIsMember(Ctx, key, member).Result()
}

// ScriptLoad 将脚本加载到 Redis 缓存中，但并不立即执行
func ScriptLoad(script string) (string, error) {
	return Client.ScriptLoad(Ctx, script).Result()
}

// EvalSha 执行 Lua 脚本
func EvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return Client.EvalSha(Ctx, sha1, keys, args...).Result()
}

// Close 关闭 Redis 连接
func Close() error {
	if Client != nil {
		fmt.Println("Closing Redis connection...")
		err := Client.Close()
		if err != nil {
			fmt.Printf("Error closing Redis: %v\n", err)
			return err
		}
		fmt.Println("Redis connection closed successfully")
	}
	return nil
}
