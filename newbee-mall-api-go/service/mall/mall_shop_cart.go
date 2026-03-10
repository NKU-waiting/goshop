package mall

import (
	"errors"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"main.go/constants"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
	mallRes "main.go/model/mall/response"
	"main.go/model/manage"
	"main.go/utils"
)

type MallShopCartService struct {
}

// GetMyShoppingCartItems 不分页
func (m *MallShopCartService) GetMyShoppingCartItems(token string) (err error, cartItems []mallRes.CartItemResponse) {
	var userToken mall.MallUserToken
	var shopCartItems []mall.MallShoppingCartItem
	var goodsInfos []manage.MallGoodsInfo
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), cartItems
	}
	global.GVA_DB.Where("user_id=? and is_deleted = 0", userToken.UserId).Find(&shopCartItems)
	var goodsIds []int
	for _, shopcartItem := range shopCartItems {
		goodsIds = append(goodsIds, shopcartItem.GoodsId)
	}
	global.GVA_DB.Where("goods_id in ?", goodsIds).Find(&goodsInfos)
	goodsMap := make(map[int]manage.MallGoodsInfo)
	for _, goodsInfo := range goodsInfos {
		goodsMap[goodsInfo.GoodsId] = goodsInfo
	}
	for _, v := range shopCartItems {
		var cartItem mallRes.CartItemResponse
		copier.Copy(&cartItem, &v)
		if _, ok := goodsMap[v.GoodsId]; ok {
			goodsInfo := goodsMap[v.GoodsId]
			cartItem.GoodsName = goodsInfo.GoodsName
			cartItem.GoodsCoverImg = goodsInfo.GoodsCoverImg
			cartItem.SellingPrice = goodsInfo.SellingPrice
		}
		cartItems = append(cartItems, cartItem)
	}

	return
}

func (m *MallShopCartService) SaveMallCartItem(token string, req mallReq.SaveCartItemParam) (err error) {
	if req.GoodsCount < 1 {
		return errors.New("商品数量不能小于 1 ！")

	}
	if req.GoodsCount > constants.MaxGoodsCountPerItem {
		return errors.New("超出单个商品的最大购买数量！")
	}
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var shopCartItem mall.MallShoppingCartItem
		// 是否已存在商品
		if err = tx.Where("user_id = ? and goods_id = ? and is_deleted = 0", userToken.UserId, req.GoodsId).First(&shopCartItem).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在则新建
				var goodsInfo manage.MallGoodsInfo
				if err = tx.Where("goods_id = ?", req.GoodsId).First(&goodsInfo).Error; err != nil {
					return errors.New("商品不存在")
				}
				var total int64
				tx.Where("user_id = ? and is_deleted = 0", userToken.UserId).Count(&total)
				if total >= constants.MaxCartItems {
					return errors.New("超出购物车最大容量！")
				}
				shopCartItem.GoodsId = req.GoodsId
				shopCartItem.GoodsCount = req.GoodsCount
				shopCartItem.UserId = userToken.UserId
				shopCartItem.CreateTime = common.JSONTime{Time: time.Now()}
				shopCartItem.UpdateTime = common.JSONTime{Time: time.Now()}
				return tx.Save(&shopCartItem).Error
			}
			return err
		} else {
			// 已存在则修改
			totalCount := shopCartItem.GoodsCount + req.GoodsCount
			if totalCount < 1 {
				return errors.New("商品数量不能小于 1 ！")
			}
			if totalCount > constants.MaxGoodsCountPerItem {
				return errors.New("超出单个商品的最大购买数量！")
			}
			shopCartItem.GoodsCount = totalCount
			shopCartItem.UpdateTime = common.JSONTime{Time: time.Now()}
			return tx.Save(&shopCartItem).Error
		}
	})
	return
}

func (m *MallShopCartService) UpdateMallCartItem(token string, req mallReq.UpdateCartItemParam) (err error) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}

	var shopCartItem mall.MallShoppingCartItem
	if err = global.GVA_DB.Where("cart_item_id=? and is_deleted = 0", req.CartItemId).First(&shopCartItem).Error; err != nil {
		return errors.New("未查询到记录！")
	}

	if shopCartItem.UserId != userToken.UserId {
		return errors.New("禁止该操作！")
	}

	// 如果数量小于等于0，则删除
	if req.GoodsCount <= 0 {
		return m.deleteMallCartItemInternal(req.CartItemId, userToken.UserId)
	}

	// 超出单个商品的最大数量
	if req.GoodsCount > constants.MaxGoodsCountPerItem {
		return errors.New("超出单个商品的最大购买数量！")
	}

	shopCartItem.GoodsCount = req.GoodsCount
	shopCartItem.UpdateTime = common.JSONTime{Time: time.Now()}
	err = global.GVA_DB.Save(&shopCartItem).Error
	return
}

func (m *MallShopCartService) DeleteMallCartItem(token string, id int) (err error) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	return m.deleteMallCartItemInternal(id, userToken.UserId)
}

// 内部函数，用于删除购物车项目
func (m *MallShopCartService) deleteMallCartItemInternal(cartItemId int, userId int) error {
	var shopCartItem mall.MallShoppingCartItem
	if err := global.GVA_DB.Where("cart_item_id = ? and is_deleted = 0", cartItemId).First(&shopCartItem).Error; err != nil {
		// 如果记录本身就不存在，也认为是成功的
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if shopCartItem.UserId != userId {
		return errors.New("禁止该操作！")
	}

	return global.GVA_DB.Model(&mall.MallShoppingCartItem{}).Where("cart_item_id = ?", cartItemId).Update("is_deleted", 1).Error
}

func (m *MallShopCartService) GetCartItemsForSettle(token string, cartItemIds []int) (err error, cartItemRes []mallRes.CartItemResponse) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), cartItemRes
	}
	var shopCartItems []mall.MallShoppingCartItem
	err = global.GVA_DB.Where("cart_item_id in (?) and user_id = ? and is_deleted = 0", cartItemIds, userToken.UserId).Find(&shopCartItems).Error
	if err != nil {
		return
	}
	_, cartItemRes = getMallShoppingCartItemVOS(shopCartItems)
	//购物车算价
	priceTotal := 0
	for _, cartItem := range cartItemRes {
		priceTotal = priceTotal + cartItem.GoodsCount*cartItem.SellingPrice
	}
	return
}

// 购物车数据转换
func getMallShoppingCartItemVOS(cartItems []mall.MallShoppingCartItem) (err error, cartItemsRes []mallRes.CartItemResponse) {
	var goodsIds []int
	for _, cartItem := range cartItems {
		goodsIds = append(goodsIds, cartItem.GoodsId)
	}
	var newBeeMallGoods []manage.MallGoodsInfo
	err = global.GVA_DB.Where("goods_id in ?", goodsIds).Find(&newBeeMallGoods).Error
	if err != nil {
		return
	}

	newBeeMallGoodsMap := make(map[int]manage.MallGoodsInfo)
	for _, goodsInfo := range newBeeMallGoods {
		newBeeMallGoodsMap[goodsInfo.GoodsId] = goodsInfo
	}
	for _, cartItem := range cartItems {
		var cartItemRes mallRes.CartItemResponse
		copier.Copy(&cartItemRes, &cartItem)
		// 是否包含key
		if _, ok := newBeeMallGoodsMap[cartItemRes.GoodsId]; ok {
			newBeeMallGoodsTemp := newBeeMallGoodsMap[cartItemRes.GoodsId]
			cartItemRes.GoodsCoverImg = newBeeMallGoodsTemp.GoodsCoverImg
			goodsName := utils.SubStrLen(newBeeMallGoodsTemp.GoodsName, 28)
			cartItemRes.GoodsName = goodsName
			cartItemRes.SellingPrice = newBeeMallGoodsTemp.SellingPrice
			cartItemsRes = append(cartItemsRes, cartItemRes)
		}
	}
	return
}
