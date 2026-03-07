package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"main.go/pkg/redis"
)

type TestGoods struct {
	GoodsId   int    `json:"goods_id"`
	GoodsName string `json:"goods_name"`
	Price     int    `json:"price"`
}

func TestCacheService(t *testing.T) {
	// 检查 Redis 是否已初始化
	if redis.Client == nil {
		t.Skip("跳过测试：Redis 未初始化")
		return
	}

	cacheService := &CacheService{}

	t.Run("商品详情缓存-设置和获取", func(t *testing.T) {
		// 准备测试数据
		testGoods := TestGoods{
			GoodsId:   1001,
			GoodsName: "测试商品",
			Price:     9999,
		}

		// 设置缓存
		err := cacheService.SetGoodsDetail(testGoods.GoodsId, testGoods)
		if err != nil {
			t.Logf("设置缓存失败（可能Redis未启动）: %v", err)
			t.Skip("跳过测试：Redis未启动")
		}

		// 获取缓存
		var result TestGoods
		found, err := cacheService.GetGoodsDetail(testGoods.GoodsId, &result)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, testGoods.GoodsId, result.GoodsId)
		assert.Equal(t, testGoods.GoodsName, result.GoodsName)
		assert.Equal(t, testGoods.Price, result.Price)

		// 删除缓存
		err = cacheService.DeleteGoodsDetail(testGoods.GoodsId)
		assert.NoError(t, err)

		// 验证已删除
		found, err = cacheService.GetGoodsDetail(testGoods.GoodsId, &result)
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("商品详情缓存-不存在的商品", func(t *testing.T) {
		var result TestGoods
		found, err := cacheService.GetGoodsDetail(99999, &result)
		if err != nil {
			t.Logf("Redis未启动: %v", err)
			t.Skip("跳过测试：Redis未启动")
		}
		assert.NoError(t, err)
		assert.False(t, found)
	})
}
