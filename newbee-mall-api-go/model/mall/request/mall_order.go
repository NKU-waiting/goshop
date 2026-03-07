package request

import (
	"main.go/model/mall"
	mallRes "main.go/model/mall/response"
)

type PaySuccessParams struct {
	OrderNo string `json:"orderNo"`
	PayType int    `json:"payType"`
}

type OrderSearchParams struct {
	Status     string `form:"status"`
	PageNumber int    `form:"pageNumber"`
}

type SaveOrderParam struct {
	CartItemIds []int `json:"cartItemIds"`
	AddressId   int   `json:"addressId"`
}

// OrderEvent 定义了创建订单时发送到 Kafka 的消息结构
type OrderEvent struct {
	OrderNo             string                       `json:"orderNo"`
	UserId              int                          `json:"userId"`
	Address             mall.MallUserAddress         `json:"address"`
	ShoppingCartItemIds []int                        `json:"shoppingCartItemIds"`
	Items               []mallRes.CartItemResponse `json:"items"`
}
