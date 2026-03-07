package mall

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
	"main.go/model/manage"
)

func TestSaveMallCartItem(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)

	// 迁移购物车和商品表
	global.GVA_DB.AutoMigrate(&mall.MallShoppingCartItem{}, &manage.MallGoodsInfo{})

	userService := &MallUserService{}
	cartService := &MallShopCartService{}

	// 创建测试用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "carttest",
		Password:  "123456",
	}
	err := userService.RegisterUser(registerReq)
	assert.NoError(t, err)

	// 登录获取 token
	loginReq := mallReq.UserLoginParam{
		LoginName:   "carttest",
		PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
	}
	err, _, userToken := userService.UserLogin(loginReq)
	assert.NoError(t, err)

	// 创建测试商品
	testGoods := manage.MallGoodsInfo{
		GoodsId:       10001,
		GoodsName:     "测试商品",
		GoodsCoverImg: "test.jpg",
		SellingPrice:  100,
		StockNum:      100,
		GoodsSellStatus: 0,
		CreateTime:    common.JSONTime{Time: time.Now()},
	}
	err = global.GVA_DB.Create(&testGoods).Error
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		req     mallReq.SaveCartItemParam
		wantErr bool
		errMsg  string
	}{
		{
			name:  "成功添加商品到购物车",
			token: userToken.Token,
			req: mallReq.SaveCartItemParam{
				GoodsId:    10001,
				GoodsCount: 2,
			},
			wantErr: false,
		},
		{
			name:  "商品数量小于1",
			token: userToken.Token,
			req: mallReq.SaveCartItemParam{
				GoodsId:    10001,
				GoodsCount: 0,
			},
			wantErr: true,
			errMsg:  "商品数量不能小于 1",
		},
		{
			name:  "商品数量超过5",
			token: userToken.Token,
			req: mallReq.SaveCartItemParam{
				GoodsId:    10001,
				GoodsCount: 6,
			},
			wantErr: true,
			errMsg:  "超出单个商品的最大购买数量",
		},
		{
			name:  "无效的Token",
			token: "invalidtoken",
			req: mallReq.SaveCartItemParam{
				GoodsId:    10001,
				GoodsCount: 1,
			},
			wantErr: true,
			errMsg:  "不存在的用户",
		},
		{
			name:  "商品不存在",
			token: userToken.Token,
			req: mallReq.SaveCartItemParam{
				GoodsId:    99999,
				GoodsCount: 1,
			},
			wantErr: true,
			errMsg:  "商品为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cartService.SaveMallCartItem(tt.token, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetMyShoppingCartItems(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)

	// 迁移相关表
	global.GVA_DB.AutoMigrate(&mall.MallShoppingCartItem{}, &manage.MallGoodsInfo{})

	userService := &MallUserService{}
	cartService := &MallShopCartService{}

	// 创建测试用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "cartlisttest",
		Password:  "123456",
	}
	err := userService.RegisterUser(registerReq)
	assert.NoError(t, err)

	// 登录获取 token
	loginReq := mallReq.UserLoginParam{
		LoginName:   "cartlisttest",
		PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
	}
	err, user, userToken := userService.UserLogin(loginReq)
	assert.NoError(t, err)

	// 创建测试商品
	testGoods := manage.MallGoodsInfo{
		GoodsId:       10002,
		GoodsName:     "测试商品2",
		GoodsCoverImg: "test2.jpg",
		SellingPrice:  200,
		StockNum:      50,
		GoodsSellStatus: 0,
		CreateTime:    common.JSONTime{Time: time.Now()},
	}
	err = global.GVA_DB.Create(&testGoods).Error
	assert.NoError(t, err)

	// 添加商品到购物车
	cartItem := mall.MallShoppingCartItem{
		UserId:     user.UserId,
		GoodsId:    10002,
		GoodsCount: 3,
		IsDeleted:  0,
		CreateTime: common.JSONTime{Time: time.Now()},
	}
	err = global.GVA_DB.Create(&cartItem).Error
	assert.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "成功获取购物车列表",
			token:     userToken.Token,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "无效的Token",
			token:     "invalidtoken",
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, cartItems := cartService.GetMyShoppingCartItems(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cartItems, tt.wantCount)
				if tt.wantCount > 0 {
					assert.Equal(t, "测试商品2", cartItems[0].GoodsName)
					assert.Equal(t, 3, cartItems[0].GoodsCount)
				}
			}
		})
	}
}
