<div align="center">
   <img alt="logo" width="100" height="100" src="https://foruda.gitee.com/images/1733417239320800627/3c5290fe_716974.png">
   <h2>youlai-gin</h2>
   <img alt="Go" src="https://img.shields.io/badge/Go-1.21+-blue.svg"/>
   <img alt="Gin" src="https://img.shields.io/badge/Gin-1.11.0-green.svg"/>
   <a href="https://gitee.com/youlaiorg/youlai-gin" target="_blank">
     <img alt="Gitee star" src="https://gitee.com/youlaiorg/youlai-gin/badge/star.svg"/>
   </a>     
   <a href="https://github.com/youlaitech/youlai-gin" target="_blank">
     <img alt="Github star" src="https://img.shields.io/github/stars/youlaitech/youlai-gin.svg?style=social&label=Stars"/>
   </a>
</div>

<p align="center">
  <a target="_blank" href="https://vue.youlai.tech/">🖥️ 在线预览</a>
  <span>&nbsp;|&nbsp;</span>
  <a target="_blank" href="https://www.youlai.tech/youlai-gin">📑 阅读文档</a>
  <span>&nbsp;|&nbsp;</span>
  <a target="_blank" href="https://www.youlai.tech">🌐 官网</a>
</p>

## 📢 项目简介

`youlai-gin` 是 `vue3-element-admin` 配套的 Go 语言后端实现，基于 Go 1.21, Gin, GORM, JWT, Redis, MySQL 构建，是 **youlai 全家桶** 的重要组成部分。

- **🚀 快速开发**: 以 Gin 框架为基础，提供高性能的 Web API，代码简洁，易于上手。
- **🔐 安全认证**: 集成 JWT 认证机制，支持与 Redis 结合的会话管理。
- **🔑 权限管理**: 内置基于 RBAC 的权限模型，精确控制接口和按钮权限。
- **🛠️ 功能模块**: 包含用户、角色、菜单、部门、字典等后台管理系统的核心功能。

## 🌈 项目源码

| 项目类型 | Gitee | Github | GitCode |
| --- | --- | --- | --- |
| ✅ Go 后端 | [youlai-gin](https://gitee.com/youlaiorg/youlai-gin) | [youlai-gin](https://github.com/youlaitech/youlai-gin) | [youlai-gin](https://gitcode.com/youlai/youlai-gin) |
| vue3 前端 | [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) | [vue3-element-admin](https://github.com/youlaitech/vue3-element-admin) | [vue3-element-admin](https://gitcode.com/youlai/vue3-element-admin) |
| uni-app 移动端 | [vue-uniapp-template](https://gitee.com/youlaiorg/vue-uniapp-template) | [vue-uniapp-template](https://github.com/youlaitech/vue-uniapp-template) | [vue-uniapp-template](https://gitcode.com/youlai/vue-uniapp-template) |

## 📚 项目文档

| 文档名称           | 访问地址                                                                 |
| ------------------ | ------------------------------------------------------------------------ |
| 项目介绍与使用指南 | [https://www.youlai.tech/youlai-gin](https://www.youlai.tech/youlai-gin) |

## 📁 项目目录

<details>
<summary> 目录结构 </summary>

```text
youlai-gin/
├─ internal/                  # 核心业务源码
│  ├─ auth/                   # 认证模块（登录/Token/会话）
│  ├─ health/                 # 健康检查
│  ├─ platform/               # 平台模块（文件/扩展能力）
│  ├─ router/                 # 路由注册
│  └─ system/                 # 系统模块（用户/角色/菜单等）
│
├─ pkg/                       # 通用包（中间件/响应/工具等）
├─ configs/                   # 配置文件
│  ├─ dev.yaml
│  ├─ prod.yaml
│  └─ test.yaml
│
├─ scripts/                   # 数据库脚本
│  └─ mysql/
│     └─ youlai_admin.sql     # 建库 / 建表 / 初始化数据
│
├─ main.go                    # 应用入口
├─ Dockerfile                 # Docker 镜像构建
├─ go.mod                     # Go Module 定义
└─ go.sum                     # 依赖锁定
```

</details>

## 环境准备

### 1. 准备基础环境

| 要求      | 说明              |
| --------- | ----------------- |
| **Go**    | `1.25` 或更高版本 |
| **MySQL** | `5.7` 或 `8.x`    |
| **Redis** | `7.x`             |

> ⚠️ **重要提示**：MySQL 与 Redis 为项目启动必需依赖，请确保服务已启动。

### 2. 安装开发工具

**GoLand**（推荐）：

- 直接使用 JetBrains GoLand 即可，首次打开项目时按提示下载/配置 Go SDK

**VS Code**：

1. **安装 Go**: 建议安装 `1.25` 或更高版本 ([官方下载](https://go.dev/dl/))，安装后请在终端执行 `go version` 验证。

2. 安装 VS Code 扩展插件（VS Code 扩展市场搜索安装）：

   | 插件名称             | 作用                                    |
   | -------------------- | --------------------------------------- |
   | **Go**               | Go 语言支持（gopls/调试/格式化/测试）   |
   | **Go Test Explorer** | 测试用例可视化运行（可选）              |
   | **REST Client**      | 直接在 VS Code 内调试 HTTP 接口（可选） |

### 3. 初始化数据库

使用数据库客户端（如 Navicat、DBeaver）执行项目根目录下的 `scripts/mysql/youlai_admin.sql` 脚本，完成数据库及基础数据的初始化。

## 项目启动

### 1. 配置应用程序

开发环境配置文件：`configs/dev.yaml`

```yaml
database:
  host: localhost
  port: 3306
  username: youlai
  password: 123456
  dbname: youlai_admin

redis:
  host: localhost
  port: 6379
  password: ""
  database: 0

security:
  sessionType: jwt # jwt / redis-token
  jwt:
    secretKey: "请改为生产安全密钥" # 生产环境建议使用至少 32 字节的随机字符串
    accessTokenTTL: 7200
    refreshTokenTTL: 2592000
  redisToken:
    accessTokenTTL: 7200
    refreshTokenTTL: 2592000
```

**配置项说明：**

- `database.*`：MySQL 连接信息，启动前请确保库表已初始化。
- `redis.*`：Redis 连接配置，用于会话与缓存。
- `security.sessionType`：会话模式，`jwt` 为无状态，`redis-token` 为服务端会话。
- `security.jwt.secretKey`：JWT 签名密钥，生产务必修改。
- `security.redisToken.*`：选择 `redis-token` 模式时的会话 TTL。

其他环境可参考 `configs/prod.yaml`、`configs/test.yaml`，所有字段均可用环境变量 `APP_<模块>_<字段>` 形式覆盖（如 `APP_DATABASE_PASSWORD`）。

### 2. 启动后端服务

```bash
# 1. 克隆项目
git clone https://gitee.com/youlaiorg/youlai-gin.git
cd youlai-gin

# 2. 下载依赖
go mod tidy

# 3. 生成 Swagger 文档
swag init

# 4. 启动项目
go run main.go
```

> 💡 **开发技巧：热重载**
>
> 为了提升开发效率，避免每次修改代码后都手动重启服务，推荐使用 `air` 工具实现热重载。
>
> ```bash
> # 1. 安装 air
> go install github.com/cosmtrek/air@latest
>
> # 2. 在项目根目录启动 (代替 go run)
> air
> ```
>
> `air` 会自动监听文件变动并重新编译启动项目。首次使用时，它会在项目根目录生成一个 `.air.toml` 配置文件，通常无需修改。

启动成功后，访问 [http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html) 验证项目是否成功。

### 3. 整合并启动前端

`youlai-gin` 与 `vue3-element-admin` 完全兼容。

```bash
# 1. 获取前端项目
git clone https://gitee.com/youlaiorg/vue3-element-admin.git
cd vue3-element-admin

# 2. 安装依赖 (推荐使用 pnpm)
pnpm install

# 3. 配置后端接口地址 (编辑 .env.development)
VITE_APP_API_URL=http://localhost:8000

# 4. 启动前端
pnpm run dev
```

## 🐳 项目部署

### 1. 传统部署

```bash
# 编译
go build -o youlai-gin

# 运行
./youlai-gin
```

### 2. Docker 部署

```bash
# 构建镜像
docker build -t youlai-gin:latest .

# 运行容器
docker run -d -p 8000:8000 --name youlai-gin youlai-gin:latest
```

## 💖 技术交流

- **问题反馈**：[Gitee Issues](https://gitee.com/youlaiorg/youlai-gin/issues)
- **技术交流群**：[QQ 群：950387562](https://qm.qq.com/cgi-bin/qm/qr?k=U57IDw7ufwuzMA4qQ7BomwZ44hpHGkLg)
- **博客教程**：[https://www.youlai.tech](https://www.youlai.tech)
