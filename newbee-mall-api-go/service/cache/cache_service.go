package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"main.go/pkg/redis"
)

// CacheService 缓存服务
type CacheService struct{}

// 缓存键前缀
const (
	GoodsDetailPrefix  = "goods:detail:"
	CategoryListPrefix = "category:list:"
	CarouselPrefix     = "carousel:list"
)

// 缓存过期时间
const (
	GoodsDetailTTL  = 1 * time.Hour
	CategoryListTTL = 24 * time.Hour
	CarouselTTL     = 1 * time.Hour
)

// GetGoodsDetail 获取商品详情缓存
func (c *CacheService) GetGoodsDetail(goodsId int, result interface{}) (bool, error) {
	key := fmt.Sprintf("%s%d", GoodsDetailPrefix, goodsId)
	data, err := redis.Get(key)
	if err != nil {
		return false, nil
	}
	
	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetGoodsDetail 设置商品详情缓存
func (c *CacheService) SetGoodsDetail(goodsId int, data interface{}) error {
	key := fmt.Sprintf("%s%d", GoodsDetailPrefix, goodsId)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return redis.Set(key, jsonData, GoodsDetailTTL)
}

// DeleteGoodsDetail 删除商品详情缓存
func (c *CacheService) DeleteGoodsDetail(goodsId int) error {
	key := fmt.Sprintf("%s%d", GoodsDetailPrefix, goodsId)
	return redis.Del(key)
}

// GetCategoryList 获取分类列表缓存
func (c *CacheService) GetCategoryList(result interface{}) (bool, error) {
	key := CategoryListPrefix + "all"
	data, err := redis.Get(key)
	if err != nil {
		return false, nil
	}

	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetCategoryList 设置分类列表缓存
func (c *CacheService) SetCategoryList(data interface{}) error {
	key := CategoryListPrefix + "all"
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return redis.Set(key, jsonData, CategoryListTTL)
}

// DeleteCategoryList 删除分类列表缓存
func (c *CacheService) DeleteCategoryList() error {
	key := CategoryListPrefix + "all"
	return redis.Del(key)
}
