package mall

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
	mallRes "main.go/model/mall/response"
	"main.go/utils"
)

type MallUserService struct {
}

// RegisterUser 注册用户
func (m *MallUserService) RegisterUser(req mallReq.RegisterUserParam) (err error) {
	if !errors.Is(global.GVA_DB.Where("login_name =?", req.LoginName).First(&mall.MallUser{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同用户名")
	}

	return global.GVA_DB.Create(&mall.MallUser{
		LoginName:     req.LoginName,
		PasswordMd5:   utils.MD5V([]byte(req.Password)),
		IntroduceSign: "随新所欲，蜂富多彩",
		CreateTime:    common.JSONTime{Time: time.Now()},
	}).Error

}

// UpdateUserInfo 更新用户信息
func (m *MallUserService) UpdateUserInfo(token string, req mallReq.UpdateUserInfoParam) (err error) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	var userInfo mall.MallUser
	err = global.GVA_DB.Where("user_id =?", userToken.UserId).First(&userInfo).Error
	// 若密码为空字符，则表明用户不打算修改密码，使用原密码保存
	if !(req.PasswordMd5 == "") {
		userInfo.PasswordMd5 = utils.MD5V([]byte(req.PasswordMd5))
	}
	userInfo.NickName = req.NickName
	userInfo.IntroduceSign = req.IntroduceSign
	err = global.GVA_DB.Where("user_id =?", userToken.UserId).UpdateColumns(&userInfo).Error
	return
}

func (m *MallUserService) GetUserDetail(token string) (err error, userDetail mallRes.MallUserDetailResponse) {
	var userToken mall.MallUserToken
	err = global.GVA_DB.Where("token =?", token).First(&userToken).Error
	if err != nil {
		return errors.New("不存在的用户"), userDetail
	}
	var userInfo mall.MallUser
	err = global.GVA_DB.Where("user_id =?", userToken.UserId).First(&userInfo).Error
	if err != nil {
		return errors.New("用户信息获取失败"), userDetail
	}
	err = copier.Copy(&userDetail, &userInfo)
	return
}

// UserLogin 用户登陆
func (m *MallUserService) UserLogin(params mallReq.UserLoginParam) (err error, user mall.MallUser, userToken mall.MallUserToken) {
	err = global.GVA_DB.Where("login_name=? AND password_md5=?", params.LoginName, params.PasswordMd5).First(&user).Error
	if err != nil {
		return err, user, userToken
	}

	// 生成新的 token
	newToken := getNewToken(time.Now().UnixNano()/1e6, int(user.UserId))

	nowDate := time.Now()
	// 48小时过期
	expireTime, _ := time.ParseDuration("48h")
	expireDate := nowDate.Add(expireTime)

	// 设置 token 信息
	userToken.UserId = user.UserId
	userToken.Token = newToken
	userToken.UpdateTime = nowDate
	userToken.ExpireTime = expireDate

	// Save() 会根据主键自动判断 INSERT 或 UPDATE
	err = global.GVA_DB.Save(&userToken).Error

	return err, user, userToken
}

func getNewToken(timeInt int64, userId int) (token string) {
	var build strings.Builder
	build.WriteString(strconv.FormatInt(timeInt, 10))
	build.WriteString(strconv.Itoa(userId))
	build.WriteString(utils.GenValidateCode(6))
	return utils.MD5V([]byte(build.String()))
}
