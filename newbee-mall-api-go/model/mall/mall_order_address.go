package mall

// MallOrderAddress 订单地址表
type MallOrderAddress struct {
	OrderId       int    `json:"orderId" gorm:"primaryKey;column:order_id;comment:订单id;type:bigint"`
	UserName      string `json:"userName" gorm:"column:user_name;comment:收货人姓名;type:varchar(30);"`
	UserPhone     string `json:"userPhone" gorm:"column:user_phone;comment:收货人手机号;type:varchar(11);"`
	ProvinceName  string `json:"provinceName" gorm:"column:province_name;comment:省;type:varchar(32);"`
	CityName      string `json:"cityName" gorm:"column:city_name;comment:城;type:varchar(32);"`
	RegionName    string `json:"regionName" gorm:"column:region_name;comment:区;type:varchar(32);"`
	DetailAddress string `json:"detailAddress" gorm:"column:detail_address;comment:收件详细地址;type:varchar(64);"`
}

// TableName MallOrderAddress 表名
func (MallOrderAddress) TableName() string {
	return "tb_newbee_mall_order_address"
}
