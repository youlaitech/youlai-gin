# 🚀 Youlai-Gin 企业级权限管理系统

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin Version](https://img.shields.io/badge/Gin-1.9+-00ADD8?style=flat&logo=go)](https://gin-gonic.com/)
[![GORM Version](https://img.shields.io/badge/GORM-1.25+-00ADD8?style=flat&logo=go)](https://gorm.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**基于 Gin + GORM + Redis 的前后端分离权限管理系统（Go 版本）**

[在线预览](http://admin.youlai.tech) | [前端仓库](https://github.com/youlaitech/vue3-element-admin) | [Java 版本](https://github.com/youlaitech/youlai-boot) | [文档地址](./vue3-element-admin-docs)

</div>

---

## 📖 项目简介

`youlai-gin` 是一款基于 Go 语言的企业级权限管理系统，采用前后端分离架构，提供完善的 RBAC 权限控制、数据权限、操作日志、文件上传、Excel 导入导出、WebSocket 实时通知等功能。

### ✨ 核心特性

- 🔐 **完善的权限体系**：RBAC 权限控制 + 数据权限（全部、部门及以下、仅本部门、仅本人、自定义）
- 🎯 **RESTful API**：标准 RESTful 风格接口设计，与 Java 版本（youlai-boot）保持 100% 接口协议一致
- 📊 **数据驱动**：基于 GORM 的数据访问层，支持多种数据库（MySQL、PostgreSQL）
- 🚀 **高性能缓存**：Redis 缓存用户权限、配置信息，提升系统性能
- 📝 **完整日志**：操作日志记录、访问趋势统计、访问统计分析
- 📁 **文件存储**：支持本地存储、阿里云 OSS、腾讯云 COS、七牛云 Kodo
- 📄 **Excel 处理**：基于 excelize 的导入导出功能，支持用户批量导入
- 🔔 **实时通知**：WebSocket 推送系统通知公告
- 📚 **API 文档**：集成 Swagger 自动生成接口文档
- 🏗️ **分层架构**：清晰的代码分层（Handler -> Service -> Repository）
- 🛡️ **安全可靠**：JWT 认证、密码加密、防重复提交、接口限流

### 🎨 前端项目

- **Vue3 版本**：[vue3-element-admin](https://github.com/youlaitech/vue3-element-admin)
- **技术栈**：Vue3 + TypeScript + Element Plus + Vite + Pinia

---

## 🏗️ 系统架构

### 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 编程语言 |
| Gin | 1.9+ | Web 框架 |
| GORM | 1.25+ | ORM 框架 |
| Redis | 7.0+ | 缓存数据库 |
| MySQL | 8.0+ | 关系型数据库 |
| JWT | - | 身份认证 |
| Swaggo | 1.16+ | API 文档生成 |
| Zap | 1.27+ | 日志框架 |
| Viper | 1.18+ | 配置管理 |
| Excelize | 2.8+ | Excel 处理 |
| Gorilla WebSocket | 1.5+ | WebSocket 支持 |

### 项目结构

```
youlai-gin/
├── cmd/                    # 命令行入口（可选）
├── configs/                # 配置文件目录
│   ├── config.dev.yaml    # 开发环境配置
│   ├── config.prod.yaml   # 生产环境配置
│   └── config.test.yaml   # 测试环境配置
├── docs/                   # Swagger 文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/               # 内部代码（不对外暴露）
│   ├── auth/              # 认证模块
│   ├── database/          # 数据库初始化
│   ├── middleware/        # 中间件
│   ├── platform/          # 平台服务层
│   │   └── file/         # 文件管理
│   ├── router/            # 路由注册
│   └── system/            # 系统管理层
│       ├── user/         # 用户管理
│       ├── role/         # 角色管理
│       ├── menu/         # 菜单管理
│       ├── dept/         # 部门管理
│       ├── dict/         # 字典管理
│       ├── config/       # 系统配置
│       ├── notice/       # 通知公告
│       └── log/          # 日志管理
├── pkg/                    # 公共包（可对外暴露）
│   ├── auth/              # JWT 认证
│   ├── common/            # 通用结构
│   ├── config/            # 配置加载
│   ├── context/           # 上下文处理
│   ├── database/          # 数据库工具
│   ├── errs/              # 错误定义
│   ├── excel/             # Excel 工具
│   ├── logger/            # 日志工具
│   ├── redis/             # Redis 客户端
│   ├── response/          # 响应封装
│   ├── storage/           # 文件存储
│   ├── utils/             # 工具函数
│   ├── validator/         # 参数校验
│   └── websocket/         # WebSocket
├── scripts/                # 脚本文件
│   ├── sql/               # 数据库脚本
│   └── deploy/            # 部署脚本
├── uploads/                # 文件上传目录（本地存储）
├── .gitignore             # Git 忽略文件
├── go.mod                 # Go 模块依赖
├── go.sum                 # 依赖校验文件
├── main.go                # 程序入口
└── README.md              # 项目说明
```

---

## 🚀 快速开始

### 环境要求

| 软件 | 版本要求 | 说明 |
|------|---------|------|
| Go | 1.21+ | [下载地址](https://go.dev/dl/) |
| MySQL | 8.0+ | 关系型数据库 |
| Redis | 7.0+ | 缓存数据库 |
| Node.js | 18+ | 前端开发环境（可选） |

### 1️⃣ 克隆项目

```bash
# 克隆后端项目
git clone https://github.com/youlaitech/youlai-gin.git
cd youlai-gin

# 克隆前端项目（可选）
git clone https://github.com/youlaitech/vue3-element-admin.git
```

### 2️⃣ 安装依赖

```bash
# 下载 Go 模块依赖
go mod download

# 或者使用 tidy 自动整理依赖
go mod tidy
```

### 3️⃣ 数据库初始化

```bash
# 1. 创建数据库
mysql -u root -p
CREATE DATABASE youlai DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

# 2. 导入数据库脚本
mysql -u root -p youlai < scripts/sql/youlai.sql

# 或使用数据库管理工具（Navicat、DBeaver 等）导入
```

**初始账号密码：**
- 超级管理员：`root` / `123456`
- 普通管理员：`admin` / `123456`

### 4️⃣ 修改配置

编辑配置文件 `configs/config.dev.yaml`：

```yaml
# 服务配置
server:
  port: 8000

# 数据库配置
database:
  host: localhost
  port: 3306
  username: root
  password: 123456
  database: youlai
  charset: utf8mb4
  parseTime: true
  loc: Local

# Redis 配置
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# JWT 配置
security:
  jwt:
    secret: your-secret-key-change-in-production
    expiration: 7200  # 2小时
    refreshExpiration: 604800  # 7天

# 文件存储配置
storage:
  type: local  # local, aliyun
  local:
    path: ./uploads
    urlPrefix: http://localhost:8000/uploads
  # aliyun:
  #   endpoint: oss-cn-hangzhou.aliyuncs.com
  #   accessKeyId: your-access-key-id
  #   accessKeySecret: your-access-key-secret
  #   bucketName: your-bucket-name
```

### 5️⃣ 生成 Swagger 文档

```bash
# 安装 swag 工具（首次使用）
go install github.com/swaggo/swag/cmd/swag@latest

# 生成 Swagger 文档
swag init -g main.go -o ./docs

# 输出信息：
# 2024/12/07 10:20:00 Generate swagger docs....
# 2024/12/07 10:20:00 Generate general API Info, search dir:./
# 2024/12/07 10:20:00 create docs.go at docs/docs.go
# 2024/12/07 10:20:00 create swagger.json at docs/swagger.json
# 2024/12/07 10:20:00 create swagger.yaml at docs/swagger.yaml
```

### 6️⃣ 启动项目

#### 开发环境

```bash
# 方式一：直接运行
go run main.go

# 方式二：使用 Air 热重载（推荐）
# 安装 Air
go install github.com/cosmtrek/air@latest

# 启动
air

# 输出信息：
# 2024/12/07 10:20:00 服务启动在 :8000 [环境: dev]
```

#### 生产环境

```bash
# 1. 编译项目
go build -o youlai-gin main.go

# 2. 设置环境变量
export APP_ENV=prod

# 3. 启动服务
./youlai-gin

# 或使用 nohup 后台运行
nohup ./youlai-gin > app.log 2>&1 &
```

### 7️⃣ 访问项目

| 服务 | 地址 | 说明 |
|------|------|------|
| 后端 API | http://localhost:8000 | 后端接口服务 |
| Swagger 文档 | http://localhost:8000/swagger/index.html | API 接口文档 |
| 前端项目 | http://localhost:3000 | Vue3 前端项目 |

---

## 📚 API 文档

### Swagger 使用

#### 1. 添加注解

在 Handler 函数上添加 Swagger 注解：

```go
// GetUserPage 用户分页列表
// @Summary 用户分页列表
// @Tags 用户管理
// @Produce json
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param username query string false "用户名"
// @Success 200 {object} response.Response{data=common.PageResult}
// @Router /api/v1/users/page [get]
func GetUserPage(c *gin.Context) {
    // 业务逻辑
}
```

#### 2. 生成文档

```bash
swag init -g main.go -o ./docs
```

#### 3. 访问文档

浏览器访问：http://localhost:8000/swagger/index.html

### 常用注解说明

| 注解 | 说明 | 示例 |
|------|------|------|
| @Summary | 接口简介 | `@Summary 用户分页列表` |
| @Description | 详细描述 | `@Description 查询用户分页列表，支持多条件筛选` |
| @Tags | 接口分组 | `@Tags 用户管理` |
| @Accept | 请求格式 | `@Accept json` |
| @Produce | 响应格式 | `@Produce json` |
| @Param | 参数说明 | `@Param id path int true "用户ID"` |
| @Success | 成功响应 | `@Success 200 {object} response.Response` |
| @Failure | 失败响应 | `@Failure 400 {object} response.Response` |
| @Router | 路由信息 | `@Router /api/v1/users/{id} [get]` |
| @Security | 安全认证 | `@Security Bearer` |

---

## 🔧 配置说明

### 配置文件

项目使用 Viper 管理配置，支持多环境配置：

```
configs/
├── config.dev.yaml     # 开发环境（默认）
├── config.test.yaml    # 测试环境
└── config.prod.yaml    # 生产环境
```

### 环境切换

通过环境变量 `APP_ENV` 切换配置：

```bash
# 开发环境（默认）
export APP_ENV=dev

# 测试环境
export APP_ENV=test

# 生产环境
export APP_ENV=prod
```

### 完整配置示例

```yaml
# 服务器配置
server:
  port: 8000
  mode: debug  # debug, release

# 数据库配置
database:
  host: localhost
  port: 3306
  username: root
  password: 123456
  database: youlai
  charset: utf8mb4
  parseTime: true
  loc: Local
  maxIdleConns: 10
  maxOpenConns: 100
  connMaxLifetime: 3600

# Redis 配置
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  poolSize: 10

# JWT 安全配置
security:
  jwt:
    secret: your-secret-key-change-in-production
    expiration: 7200        # Access Token 过期时间（秒）
    refreshExpiration: 604800  # Refresh Token 过期时间（秒）

# 日志配置
logger:
  level: debug  # debug, info, warn, error
  filename: logs/app.log
  maxSize: 100  # MB
  maxAge: 30    # 天
  maxBackups: 10
  compress: true

# 文件存储配置
storage:
  type: local  # local, aliyun
  
  # 本地存储
  local:
    path: ./uploads
    urlPrefix: http://localhost:8000/uploads
  
  # 阿里云 OSS
  aliyun:
    endpoint: oss-cn-hangzhou.aliyuncs.com
    accessKeyId: your-access-key-id
    accessKeySecret: your-access-key-secret
    bucketName: your-bucket-name
    domain: https://your-domain.com
```

---

## 🎯 核心功能

### 1. 权限管理

#### RBAC 权限模型

```
用户 (User) → 角色 (Role) → 菜单/按钮权限 (Menu)
```

- **用户管理**：用户增删改查、状态管理、密码重置、Excel 导入导出
- **角色管理**：角色配置、权限分配、数据权限设置
- **菜单管理**：菜单树、按钮权限、动态路由
- **部门管理**：部门树形结构、数据权限范围

#### 数据权限

支持 5 种数据权限范围：

| 权限范围 | 说明 | DataScope 值 |
|---------|------|--------------|
| 全部数据 | 不限制数据范围 | 0 |
| 部门及以下 | 本部门及子部门数据 | 1 |
| 仅本部门 | 只看本部门数据 | 2 |
| 仅本人 | 只看自己的数据 | 3 |
| 自定义 | 指定部门数据 | 4 |

### 2. 系统管理

- **字典管理**：系统字典维护、字典项管理
- **系统配置**：系统参数配置、Redis 缓存管理
- **操作日志**：记录用户操作、访问趋势统计、访问量分析
- **通知公告**：系统通知发布、WebSocket 实时推送、已读/未读管理

### 3. 文件管理

支持多种存储方式：

```go
// 本地存储
storage.type = "local"

// 阿里云 OSS
storage.type = "aliyun"
```

**功能特性：**
- 单文件上传
- 批量文件上传
- 图片上传（带格式、大小限制）
- 文件删除
- 自动生成唯一文件名
- 支持自定义存储路径

### 4. Excel 导入导出

基于 `excelize` 实现：

```go
// 导出用户列表
GET /api/v1/users/export

// 下载导入模板
GET /api/v1/users/template

// 导入用户数据
POST /api/v1/users/import
```

**支持功能：**
- 自定义表头
- 数据验证
- 错误行提示
- 批量导入

### 5. WebSocket 通知

实时推送系统通知：

```go
// WebSocket 连接
ws://localhost:8000/api/v1/ws

// 消息格式
{
  "type": "notice",
  "title": "系统通知",
  "content": "您有新的消息",
  "data": { ... }
}
```

---

## 🧪 开发指南

### 代码规范

#### 1. 目录命名

- 使用小写字母和下划线
- 包名使用小写字母
- 文件名使用小写字母和下划线

#### 2. 分层规范

```
Handler  -> Service -> Repository -> Model
  |           |           |
请求处理   业务逻辑   数据访问
```

**示例：**

```go
// Handler 层：处理 HTTP 请求
func GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := service.GetUserByID(id)
    response.Success(c, user)
}

// Service 层：业务逻辑
func GetUserByID(id string) (*model.User, error) {
    return repository.GetUserByID(id)
}

// Repository 层：数据访问
func GetUserByID(id string) (*model.User, error) {
    var user model.User
    err := database.DB.Where("id = ?", id).First(&user).Error
    return &user, err
}
```

#### 3. 错误处理

使用统一的错误定义：

```go
// 业务错误
return errs.BadRequest("参数错误")
return errs.NotFound("用户不存在")
return errs.Unauthorized("未登录")
return errs.Forbidden("无权限")
return errs.SystemError("系统错误")

// 自定义错误
return errs.New(40001, "自定义错误")
```

#### 4. 响应格式

统一的响应结构：

```go
// 成功响应
response.Success(c, data)

// 分页响应
response.Success(c, &common.PageResult{
    List:  list,
    Total: total,
})

// 错误响应
response.BadRequest(c, "参数错误")
response.Unauthorized(c, "未登录")
response.Forbidden(c, "无权限")
```

### 新增模块

#### 1. 创建目录结构

```bash
internal/system/example/
├── handler/
│   └── example_handler.go
├── model/
│   ├── entity.go
│   ├── form.go
│   └── vo.go
├── repository/
│   └── example_repo.go
├── service/
│   └── example_service.go
└── router.go
```

#### 2. 定义实体

```go
// model/entity.go
package model

type Example struct {
    ID         int64  `gorm:"primaryKey" json:"id"`
    Name       string `gorm:"size:100" json:"name"`
    Status     int    `gorm:"default:1" json:"status"`
    CreateTime string `gorm:"autoCreateTime" json:"createTime"`
    UpdateTime string `gorm:"autoUpdateTime" json:"updateTime"`
}

func (Example) TableName() string {
    return "sys_example"
}
```

#### 3. 实现 Repository

```go
// repository/example_repo.go
package repository

import (
    "youlai-gin/internal/database"
    "youlai-gin/internal/system/example/model"
)

func GetList() ([]model.Example, error) {
    var list []model.Example
    err := database.DB.Find(&list).Error
    return list, err
}
```

#### 4. 实现 Service

```go
// service/example_service.go
package service

import (
    "youlai-gin/internal/system/example/model"
    "youlai-gin/internal/system/example/repository"
)

func GetList() ([]model.Example, error) {
    return repository.GetList()
}
```

#### 5. 实现 Handler

```go
// handler/example_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "youlai-gin/internal/system/example/service"
    "youlai-gin/pkg/response"
)

// GetList 列表查询
// @Summary 列表查询
// @Tags 示例管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/examples [get]
func GetList(c *gin.Context) {
    list, err := service.GetList()
    if err != nil {
        response.HandleError(c, err)
        return
    }
    response.Success(c, list)
}
```

#### 6. 注册路由

```go
// router.go
package example

import (
    "github.com/gin-gonic/gin"
    "youlai-gin/internal/system/example/handler"
)

func RegisterRoutes(router *gin.RouterGroup) {
    exampleGroup := router.Group("/examples")
    {
        exampleGroup.GET("", handler.GetList)
    }
}
```

#### 7. 在 system/router.go 注册

```go
import "youlai-gin/internal/system/example"

func RegisterRoutes(r *gin.RouterGroup) {
    // ...
    example.RegisterRoutes(r)
}
```

---

## 📦 部署指南

### Docker 部署

#### 1. 创建 Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o youlai-gin main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/youlai-gin .
COPY --from=builder /app/configs ./configs

ENV TZ=Asia/Shanghai
ENV APP_ENV=prod

EXPOSE 8000
CMD ["./youlai-gin"]
```

#### 2. 构建镜像

```bash
docker build -t youlai-gin:1.0.0 .
```

#### 3. 运行容器

```bash
docker run -d \
  --name youlai-gin \
  -p 8000:8000 \
  -e APP_ENV=prod \
  -e DATABASE_HOST=mysql \
  -e DATABASE_PASSWORD=your-password \
  -e REDIS_HOST=redis \
  youlai-gin:1.0.0
```

### Docker Compose 部署

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: youlai-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: youlai
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./scripts/sql:/docker-entrypoint-initdb.d
    command: --default-authentication-plugin=mysql_native_password

  redis:
    image: redis:7-alpine
    container_name: youlai-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  backend:
    build: .
    container_name: youlai-gin
    ports:
      - "8000:8000"
    environment:
      APP_ENV: prod
      DATABASE_HOST: mysql
      DATABASE_PASSWORD: root123456
      REDIS_HOST: redis
    depends_on:
      - mysql
      - redis

volumes:
  mysql-data:
  redis-data:
```

启动服务：

```bash
docker-compose up -d
```

### 生产环境部署

#### 1. 编译发布

```bash
# 编译 Linux 版本
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o youlai-gin main.go

# 编译 Windows 版本
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o youlai-gin.exe main.go
```

#### 2. 使用 Systemd 管理

创建 `/etc/systemd/system/youlai-gin.service`：

```ini
[Unit]
Description=Youlai-Gin Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www
WorkingDirectory=/data/youlai-gin
Environment="APP_ENV=prod"
ExecStart=/data/youlai-gin/youlai-gin
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

管理服务：

```bash
# 启动服务
systemctl start youlai-gin

# 开机自启
systemctl enable youlai-gin

# 查看状态
systemctl status youlai-gin

# 查看日志
journalctl -u youlai-gin -f
```

#### 3. Nginx 反向代理

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket 支持
    location /api/v1/ws {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    # 静态文件
    location /uploads/ {
        alias /data/youlai-gin/uploads/;
    }
}
```

---

## 🔒 安全建议

### 生产环境配置

1. **修改默认密码**
   ```sql
   UPDATE sys_user SET password = MD5('new-password') WHERE username = 'root';
   ```

2. **修改 JWT Secret**
   ```yaml
   security:
     jwt:
       secret: $(openssl rand -base64 32)
   ```

3. **启用 HTTPS**
   ```nginx
   server {
       listen 443 ssl http2;
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
   }
   ```

4. **配置防火墙**
   ```bash
   # 只允许必要端口
   firewall-cmd --zone=public --add-port=80/tcp --permanent
   firewall-cmd --zone=public --add-port=443/tcp --permanent
   firewall-cmd --reload
   ```

5. **数据库权限最小化**
   ```sql
   CREATE USER 'youlai'@'%' IDENTIFIED BY 'strong-password';
   GRANT SELECT, INSERT, UPDATE, DELETE ON youlai.* TO 'youlai'@'%';
   ```

---

## 🤝 参与贡献

欢迎参与项目贡献！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

### 开发规范

- 遵循 Go 代码规范
- 完善的单元测试
- 清晰的注释文档
- 及时更新 Swagger 文档

---

## 📄 License

本项目基于 [MIT](LICENSE) 协议开源。

---

## 💬 联系方式

- **项目地址**：[https://github.com/youlaitech/youlai-gin](https://github.com/youlaitech/youlai-gin)
- **前端项目**：[https://github.com/youlaitech/vue3-element-admin](https://github.com/youlaitech/vue3-element-admin)
- **在线预览**：[http://admin.youlai.tech](http://admin.youlai.tech)
- **技术文档**：[查看文档](./vue3-element-admin-docs)
- **问题反馈**：[提交 Issue](https://github.com/youlaitech/youlai-gin/issues)

---

## ⭐ Star History

如果这个项目对你有帮助，请点个 Star ⭐️ 支持一下！

[![Star History Chart](https://api.star-history.com/svg?repos=youlaitech/youlai-gin&type=Date)](https://star-history.com/#youlaitech/youlai-gin&Date)

---

<div align="center">

**感谢使用 Youlai-Gin！** 

Made with ❤️ by [Youlai Tech](https://github.com/youlaitech)

</div>
