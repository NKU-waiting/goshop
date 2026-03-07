package mall

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/copier"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/common/enum"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
	mallRes "main.go/model/mall/response"
	"main.go/model/manage"
	"main.go/pkg/kafka"
	"main.go/utils"
)

type MallOrderService struct {
}

// SaveOrder 将订单信息发送到 Kafka
func (m *MallOrderService) SaveOrder(token string, userAddress mall.MallUserAddress, myShoppingCartItems []mallRes.CartItemResponse) (err error, orderNo string) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), ""
	}

	var itemIds []int
	var goodsIds []int
	for _, cartItem := range myShoppingCartItems {
		itemIds = append(itemIds, cartItem.CartItemId)
		goodsIds = append(goodsIds, cartItem.GoodsId)
	}

	var newBeeMallGoods []manage.MallGoodsInfo
	global.GVA_DB.Where("goods_id in ?", goodsIds).Find(&newBeeMallGoods)
	// 简单检查商品是否存在和上架状态
	if len(newBeeMallGoods) != len(goodsIds) {
		return errors.New("部分商品不存在或已下架"), ""
	}
	for _, mallGoods := range newBeeMallGoods {
		if mallGoods.GoodsSellStatus != enum.GOODS_UNDER.Code() {
			return errors.New("包含已下架商品，无法生成订单"), ""
		}
	}

	// 生成订单号
	orderNo = utils.GenOrderNo()

	// 创建订单事件
	orderEvent := mallReq.OrderEvent{
		OrderNo:             orderNo,
		UserId:              userToken.UserId,
		Address:             userAddress,
		ShoppingCartItemIds: itemIds,
		Items:               myShoppingCartItems,
	}

	// 将事件序列化为 JSON
	jsonValue, err := json.Marshal(orderEvent)
	if err != nil {
		return errors.New("创建订单失败: 无法序列化订单事件"), ""
	}

	// 发送消息到 Kafka
	err = kafka.SendMessage(context.Background(), []byte(orderNo), jsonValue)
	if err != nil {
		return errors.New("创建订单失败: 无法发送消息到队列"), ""
	}

	// 立即返回，告知用户订单正在处理中
	return nil, orderNo
}

// PaySuccess 支付订单
func (m *MallOrderService) PaySuccess(orderNo string, payType int) (err error) {
	var mallOrder manage.MallOrder
	err = global.GVA_DB.Where("order_no = ? and is_deleted=0 ", orderNo).First(&mallOrder).Error
	if mallOrder != (manage.MallOrder{}) {
		if mallOrder.OrderStatus != 0 {
			return errors.New("订单状态异常！")
		}
		mallOrder.OrderStatus = enum.ORDER_PAID.Code()
		mallOrder.PayType = payType
		mallOrder.PayStatus = 1
		mallOrder.PayTime = common.JSONTime{time.Now()}
		mallOrder.UpdateTime = common.JSONTime{time.Now()}
		err = global.GVA_DB.Save(&mallOrder).Error
	}
	return
}

// FinishOrder 完结订单
func (m *MallOrderService) FinishOrder(token string, orderNo string) (err error) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	var mallOrder manage.MallOrder
	if err = global.GVA_DB.Where("order_no=? and is_deleted = 0", orderNo).First(&mallOrder).Error; err != nil {
		return errors.New("未查询到记录！")
	}
	if mallOrder.UserId != userToken.UserId {
		return errors.New("禁止该操作！")
	}
	mallOrder.OrderStatus = enum.ORDER_SUCCESS.Code()
	mallOrder.UpdateTime = common.JSONTime{time.Now()}
	err = global.GVA_DB.Save(&mallOrder).Error
	return
}

// CancelOrder 关闭订单
func (m *MallOrderService) CancelOrder(token string, orderNo string) (err error) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	var mallOrder manage.MallOrder
	if err = global.GVA_DB.Where("order_no=? and is_deleted = 0", orderNo).First(&mallOrder).Error; err != nil {
		return errors.New("未查询到记录！")
	}
	if mallOrder.UserId != userToken.UserId {
		return errors.New("禁止该操作！")
	}
	if utils.NumsInList(mallOrder.OrderStatus, []int{enum.ORDER_SUCCESS.Code(),
		enum.ORDER_CLOSED_BY_MALLUSER.Code(), enum.ORDER_CLOSED_BY_EXPIRED.Code(), enum.ORDER_CLOSED_BY_JUDGE.Code()}) {
		return errors.New("订单状态异常！")
	}
	mallOrder.OrderStatus = enum.ORDER_CLOSED_BY_MALLUSER.Code()
	mallOrder.UpdateTime = common.JSONTime{time.Now()}
	err = global.GVA_DB.Save(&mallOrder).Error
	return
}

// GetOrderDetailByOrderNo 获取订单详情
func (m *MallOrderService) GetOrderDetailByOrderNo(token string, orderNo string) (err error, orderDetail mallRes.MallOrderDetailVO) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), orderDetail
	}
	var mallOrder manage.MallOrder
	if err = global.GVA_DB.Where("order_no=? and is_deleted = 0", orderNo).First(&mallOrder).Error; err != nil {
		return errors.New("未查询到记录！"), orderDetail
	}
	if mallOrder.UserId != userToken.UserId {
		return errors.New("禁止该操作！"), orderDetail
	}
	var orderItems []manage.MallOrderItem
	err = global.GVA_DB.Where("order_id = ?", mallOrder.OrderId).Find(&orderItems).Error
	if err != nil || len(orderItems) <= 0 {
		return errors.New("订单项不存在！"), orderDetail
	}

	var newBeeMallOrderItemVOS []mallRes.NewBeeMallOrderItemVO
	copier.Copy(&newBeeMallOrderItemVOS, &orderItems)
	copier.Copy(&orderDetail, &mallOrder)
	// 订单状态前端显示为中文
	_, OrderStatusStr := enum.GetNewBeeMallOrderStatusEnumByStatus(orderDetail.OrderStatus)
	_, payTapStr := enum.GetNewBeeMallOrderStatusEnumByStatus(orderDetail.PayType)
	orderDetail.OrderStatusString = OrderStatusStr
	orderDetail.PayTypeString = payTapStr
	orderDetail.NewBeeMallOrderItemVOS = newBeeMallOrderItemVOS

	return
}

// MallOrderListBySearch 搜索订单
func (m *MallOrderService) MallOrderListBySearch(token string, pageNumber int, status string) (err error, list []mallRes.MallOrderResponse, total int64) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), list, total
	}
	// 根据搜索条件查询
	var newBeeMallOrders []manage.MallOrder
	db := global.GVA_DB.Model(&newBeeMallOrders)

	if status != "" {
		db.Where("order_status = ?", status)
	}
	err = db.Where("user_id =? and is_deleted=0 ", userToken.UserId).Count(&total).Error
	//这里前段没有做滚动加载，直接显示全部订单
	//limit := 5
	offset := 5 * (pageNumber - 1)
	err = db.Offset(offset).Order(" order_id desc").Find(&newBeeMallOrders).Error

	var orderListVOS []mallRes.MallOrderResponse
	if total > 0 {
		//数据转换 将实体类转成vo
		copier.Copy(&orderListVOS, &newBeeMallOrders)
		//设置订单状态中文显示值
		for _, newBeeMallOrderListVO := range orderListVOS {
			_, statusStr := enum.GetNewBeeMallOrderStatusEnumByStatus(newBeeMallOrderListVO.OrderStatus)
			newBeeMallOrderListVO.OrderStatusString = statusStr
		}
		// 返回订单id
		var orderIds []int
		for _, order := range newBeeMallOrders {
			orderIds = append(orderIds, order.OrderId)
		}
		//获取OrderItem
		var orderItems []manage.MallOrderItem
		if len(orderIds) > 0 {
			global.GVA_DB.Where("order_id in ?", orderIds).Find(&orderItems)
			// 优化：使用 O(n) 复杂度构建 map
			itemByOrderIdMap := make(map[int][]manage.MallOrderItem)
			for _, orderItem := range orderItems {
				itemByOrderIdMap[orderItem.OrderId] = append(
					itemByOrderIdMap[orderItem.OrderId],
					orderItem,
				)
			}
			//封装每个订单列表对象的订单项数据
			for _, newBeeMallOrderListVO := range orderListVOS {
				if _, ok := itemByOrderIdMap[newBeeMallOrderListVO.OrderId]; ok {
					orderItemListTemp := itemByOrderIdMap[newBeeMallOrderListVO.OrderId]
					var newBeeMallOrderItemVOS []mallRes.NewBeeMallOrderItemVO
					copier.Copy(&newBeeMallOrderItemVOS, &orderItemListTemp)
					newBeeMallOrderListVO.NewBeeMallOrderItemVOS = newBeeMallOrderItemVOS
					_, OrderStatusStr := enum.GetNewBeeMallOrderStatusEnumByStatus(newBeeMallOrderListVO.OrderStatus)
					newBeeMallOrderListVO.OrderStatusString = OrderStatusStr
					list = append(list, newBeeMallOrderListVO)
				}
			}
		}
	}
	return err, list, total
}
