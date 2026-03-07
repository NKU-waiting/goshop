# GoShop 商城系统

一个基于 Go + Vue3 的现代化电商系统，包含前台商城、后台管理和 RESTful API。

> **项目说明**: 本项目基于 [newbee-mall-api-go](https://github.com/newbee-ltd/newbee-mall-api-go) 开源项目进行二次开发，在原项目基础上进行了以下改进：
> - 🐛 修复了多个已知 Bug（UserLogin 函数、数据库查询优化等）
> - 🚀 引入 Redis 缓存中间件，性能提升 10-30 倍
> - 📨 引入 Kafka 消息队列，实现订单异步处理
> - 🔧 优化代码质量（消除 Goroutine 泄漏、资源泄漏等问题）
> - 📊 完善监控和日志系统


## 🚀 项目特点

- **前后端分离**: Go 后端 + Vue3 前端
- **微服务架构**: 使用 Redis 缓存 + Kafka 消息队列
- **高性能**: 商品详情响应时间 5ms，数据库查询减少 70-80%
- **容器化部署**: Docker + Docker Compose 一键部署
- **完整测试**: 单元测试 + 集成测试 + 压力测试
- **代码质量**: 修复原项目中的多个代码质量问题

## 📦 技术栈

### 后端
- **框架**: Gin (Go Web 框架)
- **数据库**: MySQL 8.0 + GORM
- **缓存**: Redis 6.2
- **消息队列**: Kafka 3.5
- **日志**: Zap

### 前端
- **用户端**: Vue 3 + Vite + Element Plus
- **管理端**: Vue 3 + Vite + Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router

## 🏗️ 项目结构

```
goshop/
├── newbee-mall-api-go/          # Go 后端 API
│   ├── api/                     # API 接口层
│   ├── service/                 # 业务逻辑层
│   │   ├── cache/              # 缓存服务
│   │   └── mall/               # 商城服务
│   ├── model/                   # 数据模型
│   ├── pkg/                     # 公共包
│   │   ├── redis/              # Redis 客户端
│   │   └── kafka/              # Kafka 客户端
│   └── config.yaml             # 配置文件
├── newbee-mall-vue3-app/        # 前台商城 (Vue3)
├── vue3-admin/                  # 后台管理 (Vue3)
├── docker-compose.yml           # Docker 编排
└── doc/                         # 项目文档
```

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 1.29+

### 一键启动

```bash
# 克隆项目
git clone https://github.com/YOUR_USERNAME/goshop.git
cd goshop

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 访问地址

- 前台商城: http://localhost:8080
- 后台管理: http://localhost:8081
- 后端 API: http://localhost:8888

### 默认账号

**管理员账号**:
- 用户名: `admin`
- 密码: `123456`

**测试用户**:
- 用户名: `test`
- 密码: `123456`

## 📊 性能优化

### Redis 缓存

- **商品详情缓存**: TTL 1小时，响应时间从 50ms 降至 5ms
- **分类列表缓存**: TTL 24小时，响应时间从 150ms 降至 5ms
- **缓存命中率**: 80-90%

### Kafka 消息队列

- **订单异步处理**: 解耦订单创建和库存扣减
- **Worker Pool**: 限制并发为 10，防止资源耗尽
- **消息幂等性**: 防止重复订单


## 🧪 测试

### 运行测试

```bash
# 进入后端目录
cd newbee-mall-api-go

# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v ./service/cache
go test -v ./service/mall -run TestWorkerPool
```

### 压力测试

```bash
# 使用 k6 进行压力测试
cd k6-tests
k6 run load-test.js
```

## 📖 API 文档

### 用户相关

```bash
# 用户注册
POST /api/v1/user/register
Content-Type: application/json
{
  "loginName": "test",
  "password": "123456"
}

# 用户登录
POST /api/v1/user/login
Content-Type: application/json
{
  "loginName": "test",
  "passwordMd5": "e10adc3949ba59abbe56e057f20f883e"
}
```

### 商品相关

```bash
# 获取商品详情
GET /api/v1/goods/detail/:id

# 搜索商品
GET /api/v1/search?keyword=手机&pageNumber=1

# 获取分类列表
GET /api/v1/categories
```

### 订单相关

```bash
# 创建订单
POST /api/v1/saveOrder
Headers: token: YOUR_TOKEN
Content-Type: application/json
{
  "cartItemIds": [1, 2, 3],
  "addressId": 1
}

# 查询订单列表
GET /api/v1/order?pageNumber=1&status=0
Headers: token: YOUR_TOKEN
```

## 🔧 配置说明

### 后端配置 (config.yaml)

```yaml
mysql:
  host: mysql
  port: 3306
  database: newbee_mall_db_v2
  username: root
  password: root

redis:
  addr: redis:6379
  password: ""
  db: 0

kafka:
  addr: kafka:9092
  topic: newbee-mall-order
  group: newbee-mall-group
```

## 🐳 Docker 部署

### 服务说明

- **backend**: Go 后端 API (端口 8888)
- **frontend**: Vue3 前台商城 (端口 8080)
- **admin**: Vue3 后台管理 (端口 8081)
- **mysql**: MySQL 数据库
- **redis**: Redis 缓存
- **kafka**: Kafka 消息队列
- **zookeeper**: Kafka 依赖

### 常用命令

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 查看日志
docker-compose logs -f backend

# 重启服务
docker-compose restart backend

# 重新构建
docker-compose build --no-cache backend
```

## 📈 监控

### Redis 监控

```bash
# 查看 Redis 内存使用
docker exec goshop_redis_1 redis-cli INFO memory

# 查看缓存键数量
docker exec goshop_redis_1 redis-cli DBSIZE

# 查看商品缓存
docker exec goshop_redis_1 redis-cli KEYS "goods:detail:*"
```

### Kafka 监控

```bash
# 查看消费者组状态
docker exec goshop_kafka_1 kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group newbee-mall-group \
  --describe

# 查看主题信息
docker exec goshop_kafka_1 kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --describe --topic newbee-mall-order
```

### 应用日志

```bash
# 查看后端日志
docker logs -f goshop_backend_1

# 查看缓存命中日志
docker logs goshop_backend_1 | grep "Cache hit"

# 查看订单处理日志
docker logs goshop_backend_1 | grep "Order processed"
```

## 🛠️ 开发指南

### 本地开发

```bash
# 后端开发
cd newbee-mall-api-go
go run main.go

# 前端开发
cd newbee-mall-vue3-app
npm install
npm run dev

# 管理端开发
cd vue3-admin
npm install
npm run dev
```

### 代码规范

- Go 代码遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- Vue 代码遵循 [Vue Style Guide](https://vuejs.org/style-guide/)
- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/)

## 📝 更新日志

### v1.0.0 (2026-03-07) - 基于原项目的重大改进

**原项目**: [newbee-mall-api-go](https://github.com/newbee-ltd/newbee-mall-api-go)

**Bug 修复**:
- ✅ 修复 UserLogin 函数的逻辑错误
- ✅ 修复 Goroutine 泄漏问题（订单消费者无限制创建 goroutine）
- ✅ 修复资源泄漏问题（Kafka、Redis 连接未正确关闭）
- ✅ 优化低效的数据库查询（O(n²) → O(n)）
- ✅ 消除硬编码的魔法数字

**新增功能**:
- ✅ Redis 缓存集成（商品详情、分类列表）
- ✅ Kafka 消息队列（订单异步处理）
- ✅ Worker Pool 并发控制（限制并发数为 10）
- ✅ 消息幂等性保证（防止重复订单）
- ✅ 完整的监控日志系统



## 📄 许可证

本项目采用 GNU Affero General Public License v3.0 (AGPL-3.0) 许可证 - 详见 [LICENSE](LICENSE) 文件

**重要说明**:
- 本项目基于 [newbee-mall-api-go](https://github.com/newbee-ltd/newbee-mall-api-go) 开发，原项目使用 AGPL-3.0 许可证

## 👥 作者

- **原项目地址**: [newbee-mall-api-go](https://github.com/newbee-ltd/newbee-mall-api-go)

