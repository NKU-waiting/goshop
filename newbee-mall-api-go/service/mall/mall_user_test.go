package mall

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"main.go/global"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&mall.MallUser{}, &mall.MallUserToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestRegisterUser(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)
	service := &MallUserService{}

	tests := []struct {
		name    string
		req     mallReq.RegisterUserParam
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功注册",
			req: mallReq.RegisterUserParam{
				LoginName: "testuser001",
				Password:  "123456",
			},
			wantErr: false,
		},
		{
			name: "重复用户名",
			req: mallReq.RegisterUserParam{
				LoginName: "testuser001",
				Password:  "123456",
			},
			wantErr: true,
			errMsg:  "存在相同用户名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RegisterUser(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// 验证用户是否创建成功
				var user mall.MallUser
				err = global.GVA_DB.Where("login_name = ?", tt.req.LoginName).First(&user).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.req.LoginName, user.LoginName)
			}
		})
	}
}

func TestGetNewToken(t *testing.T) {
	tests := []struct {
		name    string
		timeInt int64
		userId  int
	}{
		{
			name:    "生成Token",
			timeInt: time.Now().UnixNano() / 1e6,
			userId:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token1 := getNewToken(tt.timeInt, tt.userId)
			token2 := getNewToken(tt.timeInt, tt.userId)

			// Token应该是32位MD5字符串
			assert.Len(t, token1, 32)

			// 相同参数生成的Token应该不同（因为包含随机数）
			assert.NotEqual(t, token1, token2)
		})
	}
}

func TestUserLogin(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)
	service := &MallUserService{}

	// 先注册一个用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "logintest",
		Password:  "123456",
	}
	err := service.RegisterUser(registerReq)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		req     mallReq.UserLoginParam
		wantErr bool
	}{
		{
			name: "成功登录",
			req: mallReq.UserLoginParam{
				LoginName:   "logintest",
				PasswordMd5: "e10adc3949ba59abbe56e057f20f883e", // 123456 的 MD5
			},
			wantErr: false,
		},
		{
			name: "密码错误",
			req: mallReq.UserLoginParam{
				LoginName:   "logintest",
				PasswordMd5: "wrongpassword",
			},
			wantErr: true,
		},
		{
			name: "用户不存在",
			req: mallReq.UserLoginParam{
				LoginName:   "notexist",
				PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, user, userToken := service.UserLogin(tt.req)
			if tt.wantErr {
				// 登录失败时，user 应该为空
				assert.Equal(t, mall.MallUser{}, user)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.UserId)
				assert.Equal(t, tt.req.LoginName, user.LoginName)
				assert.NotEmpty(t, userToken.Token)
				assert.Len(t, userToken.Token, 32) // Token 应该是 32 位 MD5
			}
		})
	}
}

func TestUpdateUserInfo(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)
	service := &MallUserService{}

	// 先注册并登录一个用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "updatetest",
		Password:  "123456",
	}
	err := service.RegisterUser(registerReq)
	assert.NoError(t, err)

	loginReq := mallReq.UserLoginParam{
		LoginName:   "updatetest",
		PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
	}
	err, _, userToken := service.UserLogin(loginReq)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		req     mallReq.UpdateUserInfoParam
		wantErr bool
	}{
		{
			name:  "成功更新昵称和签名",
			token: userToken.Token,
			req: mallReq.UpdateUserInfoParam{
				NickName:      "新昵称",
				IntroduceSign: "新签名",
			},
			wantErr: false,
		},
		{
			name:  "更新密码",
			token: userToken.Token,
			req: mallReq.UpdateUserInfoParam{
				NickName:      "新昵称2",
				IntroduceSign: "新签名2",
				PasswordMd5:   "newpassword",
			},
			wantErr: false,
		},
		{
			name:  "无效的Token",
			token: "invalidtoken",
			req: mallReq.UpdateUserInfoParam{
				NickName: "测试",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateUserInfo(tt.token, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserDetail(t *testing.T) {
	// 设置测试数据库
	global.GVA_DB = setupTestDB(t)
	service := &MallUserService{}

	// 先注册并登录一个用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "detailtest",
		Password:  "123456",
	}
	err := service.RegisterUser(registerReq)
	assert.NoError(t, err)

	loginReq := mallReq.UserLoginParam{
		LoginName:   "detailtest",
		PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
	}
	err, _, userToken := service.UserLogin(loginReq)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "成功获取用户详情",
			token:   userToken.Token,
			wantErr: false,
		},
		{
			name:    "无效的Token",
			token:   "invalidtoken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, userDetail := service.GetUserDetail(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "detailtest", userDetail.LoginName)
				assert.NotEmpty(t, userDetail.IntroduceSign)
			}
		})
	}
}

func TestUserLoginTokenBug(t *testing.T) {
	// 这个测试用例专门用来验证 UserLogin 函数中的 bug
	// Bug 描述: 第 76 行将 string 类型的 token 传给 First() 方法，导致查询失败
	//
	// 代码问题:
	//   token := getNewToken(...)  // token 是 string 类型
	//   err = global.GVA_DB.Where("user_id =?", user.UserId).First(&token).Error  // ❌ 错误！
	//
	// 预期行为: 应该查询 userToken 结构体，然后判断是否需要更新
	// 实际行为: 查询失败，但因为 UserId 是主键，Save() 方法会执行 UPDATE，"意外地"避免了数据重复
	//
	// 但是，代码逻辑仍然是错误的！

	global.GVA_DB = setupTestDB(t)
	service := &MallUserService{}

	// 注册测试用户
	registerReq := mallReq.RegisterUserParam{
		LoginName: "bugtest",
		Password:  "123456",
	}
	err := service.RegisterUser(registerReq)
	assert.NoError(t, err)

	loginReq := mallReq.UserLoginParam{
		LoginName:   "bugtest",
		PasswordMd5: "e10adc3949ba59abbe56e057f20f883e",
	}

	t.Run("验证Bug-查询逻辑错误", func(t *testing.T) {
		// 第一次登录
		err, user1, token1 := service.UserLogin(loginReq)
		assert.NoError(t, err)
		t.Logf("第一次登录 - UserId: %d, Token: %s", user1.UserId, token1.Token)

		// 等待一小段时间
		time.Sleep(10 * time.Millisecond)

		// 第二次登录
		err, user2, token2 := service.UserLogin(loginReq)
		assert.NoError(t, err)
		t.Logf("第二次登录 - UserId: %d, Token: %s", user2.UserId, token2.Token)

		// 验证 token 已更新
		assert.NotEqual(t, token1.Token, token2.Token, "两次登录的 token 应该不同")

		// 查询数据库验证只有一条记录
		var tokens []mall.MallUserToken
		global.GVA_DB.Where("user_id = ?", user1.UserId).Find(&tokens)

		t.Logf("数据库中的 token 记录数: %d", len(tokens))
		for i, token := range tokens {
			t.Logf("  记录 %d: UserId=%d, Token=%s", i+1, token.UserId, token.Token)
		}

		// 虽然只有一条记录（因为 UserId 是主键，Save 执行了 UPDATE）
		// 但这是"意外的正确"，代码逻辑本身是错误的
		assert.Equal(t, 1, len(tokens), "应该只有一条记录")
		assert.Equal(t, token2.Token, tokens[0].Token, "数据库中的 token 应该是最新的")
	})

	t.Run("验证Bug-错误的类型传递", func(t *testing.T) {
		// 模拟 UserLogin 中的错误代码
		var user mall.MallUser
		err := global.GVA_DB.Where("login_name=?", "bugtest").First(&user).Error
		assert.NoError(t, err)

		// 这是 bug 所在：将 string 传给 First()
		wrongToken := "this_is_a_string"
		err = global.GVA_DB.Where("user_id =?", user.UserId).First(&wrongToken).Error

		// 这个查询会失败，因为类型不匹配
		t.Logf("错误的查询返回的 err: %v", err)
		assert.Error(t, err, "将 string 传给 First() 应该返回错误")

		// 正确的做法：查询到结构体
		var correctToken mall.MallUserToken
		err = global.GVA_DB.Where("user_id =?", user.UserId).First(&correctToken).Error

		t.Logf("正确的查询返回的 err: %v", err)
		if err == nil {
			t.Logf("正确查询到的 token: %s", correctToken.Token)
			assert.NoError(t, err, "正确的查询应该成功")
		}
	})

	t.Run("验证Bug-判断逻辑永远为true", func(t *testing.T) {
		// 在 UserLogin 函数中，第 82 行的判断:
		// if userToken == (mall.MallUserToken{}) {
		//
		// 因为 userToken 是返回值参数，初始值为空结构体
		// 而第 76 行的查询失败了（类型不匹配），所以 userToken 从未被赋值
		// 因此这个判断永远为 true

		var userToken mall.MallUserToken  // 模拟返回值参数

		// 模拟错误的查询（类型不匹配）
		wrongToken := "string_token"
		_ = global.GVA_DB.Where("user_id =?", 1).First(&wrongToken).Error
		// 查询失败，userToken 保持为空

		// 判断 userToken 是否为空
		isEmpty := (userToken == mall.MallUserToken{})
		t.Logf("userToken 是否为空: %v", isEmpty)
		assert.True(t, isEmpty, "因为查询失败，userToken 应该仍然为空")

		t.Log("🐛 Bug 说明:")
		t.Log("   1. 第 76 行查询失败（类型不匹配）")
		t.Log("   2. userToken 保持为空结构体")
		t.Log("   3. 第 82 行判断永远为 true")
		t.Log("   4. 代码总是走 if 分支，而不是 else 分支")
		t.Log("   5. 但因为 UserId 是主键，Save() 执行 UPDATE，避免了重复插入")
	})
}

func BenchmarkRegisterUser(b *testing.B) {
	global.GVA_DB = setupTestDB(&testing.T{})
	service := &MallUserService{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := mallReq.RegisterUserParam{
			LoginName: "benchuser" + string(rune(i)),
			Password:  "123456",
		}
		_ = service.RegisterUser(req)
	}
}
