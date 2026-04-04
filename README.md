<div align="center">
  <img alt="logo" width="100" height="100" src="https://foruda.gitee.com/images/1733417239320800627/3c5290fe_716974.png">
  <h2>youlai-gin</h2>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.25+-blue.svg"/>
  <img alt="Gin" src="https://img.shields.io/badge/Gin-1.11.0-green.svg"/>
  <a href="https://gitcode.com/youlai/youlai-gin" target="_blank">
    <img alt="GitCode star" src="https://gitcode.com/youlai/youlai-gin/star/badge.svg"/>
  </a>
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

**[youlai-gin](https://gitee.com/youlaiorg/youlai-gin)** 是 **[vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin)** 的 Go/Gin 后端实现，接口路径与返回结构完全对齐，可直接为前端提供后端服务。

- **🚀 技术栈**：Go 1.25+ + Gin + GORM，轻量高性能组合
- **🔐 安全认证**：JWT 无状态认证 + Redis 会话管理，支持会话治理
- **🔑 权限管理**：RBAC 权限模型，菜单/按钮/接口三级权限统一治理
- **🛠️ 模块能力**：用户、角色、菜单、部门、字典、日志等核心模块开箱即用

## 🌈 项目源码

| 项目 | Gitee | GitHub | GitCode |
| --- | --- | --- | --- |
| ✅ Go 后端 | [youlai-gin](https://gitee.com/youlaiorg/youlai-gin) | [youlai-gin](https://github.com/youlaitech/youlai-gin) | [youlai-gin](https://gitcode.com/youlai/youlai-gin) |
| Vue3 管理端 | [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) | [vue3-element-admin](https://github.com/youlaitech/vue3-element-admin) | [vue3-element-admin](https://gitcode.com/youlai/vue3-element-admin) |
| uni-app 移动端 | [youlai-app](https://gitee.com/youlaiorg/youlai-app) | [youlai-app](https://github.com/youlaitech/youlai-app) | [youlai-app](https://gitcode.com/youlai/youlai-app) |

## 📁 目录结构

> 参考 [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 规范

```text
youlai-gin/
├─ api/                       # API 定义 (Swagger/OpenAPI)
├─ build/                     # 构建和部署相关
│  └─ docker/                 # Docker 配置
├─ cmd/                       # 应用入口
│  └─ server/                 # 主服务入口
├─ configs/                   # 配置文件
│  ├─ dev.yaml                # 开发环境配置
│  ├─ prod.yaml               # 生产环境配置
│  └─ test.yaml               # 测试环境配置
├─ internal/                  # 私有应用代码
│  ├─ auth/                   # 认证模块(登录/Token/会话)
│  ├─ codegen/                # 代码生成模块
│  ├─ file/                   # 文件管理模块
│  ├─ health/                 # 健康检查
│  ├─ router/                 # 路由注册
│  └─ system/                 # 系统模块(用户/角色/菜单等)
├─ pkg/                       # 可被外部使用的公共库
│  ├─ middleware/             # 中间件(JWT/CORS/RequestID)
│  ├─ response/               # 统一响应结构
│  ├─ database/               # 数据库连接
│  ├─ redis/                  # Redis 连接
│  ├─ logger/                 # 日志
│  └─ ...                     # 其他通用工具
├─ sql/                       # 数据库脚本
│  └─ mysql/                  # MySQL 脚本
├─ go.mod                     # 依赖管理
├─ go.sum                     # 依赖版本锁定
└─ Dockerfile                 # Docker 镜像构建文件
```

## 🚀 快速启动

### 1. 环境准备

| 技术 | 版本/说明 | 安装文档 |
| --- | --- | --- |
| **Go** | `1.25` 或更高版本 | [官方下载](https://go.dev/dl/) |
| **MySQL** | `5.7` 或 `8.x` | [Windows](https://youlai.blog.csdn.net/article/details/133272887) / [Linux](https://youlai.blog.csdn.net/article/details/130398179) |
| **Redis** | `7.x` | [Windows](https://youlai.blog.csdn.net/article/details/133410293) / [Linux](https://youlai.blog.csdn.net/article/details/130439335) |

> 💡 **贴心小提示**：本地未配置 MySQL、Redis 不影响启动，项目默认会连接 [youlai](https://www.youlai.tech) 线上公共环境运行，方便您快速体验。

### 2. 开发工具

**GoLand（推荐）**：

- 直接使用 JetBrains GoLand 即可，首次打开项目时按提示下载/配置 Go SDK。

**VS Code**：

1. 安装 VS Code 扩展插件（扩展市场搜索安装）：

   | 插件名称             | 作用                                    |
   | -------------------- | --------------------------------------- |
   | **Go**               | Go 语言支持（gopls/调试/格式化/测试）   |
   | **Go Test Explorer** | 测试用例可视化运行（可选）              |
   | **REST Client**      | 直接在 VS Code 内调试 HTTP 接口（可选） |

2. 安装 Go 工具链（首次使用 Go 扩展通常会提示安装）：

   在 VS Code 命令面板（`Ctrl+Shift+P`）中执行 `Go: Install/Update Tools`，建议至少安装：
   - `gopls`（语言服务）
   - `dlv`（Delve 调试器）
   - `goimports`（自动整理 imports）

### 3. 初始化数据库

使用数据库客户端（如 Navicat、DBeaver）执行 `sql/mysql/youlai_admin.sql` 脚本，完成数据库和基础数据的初始化。

### 4. 修改配置

编辑 `configs/dev.yaml` 文件，根据实际情况修改 MySQL 和 Redis 的连接字符串。

### 5. 启动项目

```bash
# 下载依赖
go mod tidy

# 生成 Swagger 文档 (可选)
# swag init

# 启动项目
go run main.go
```

> 💡 **开发技巧：热重载** 推荐使用 `air` 工具实现热重载，提升开发效率。
>
> ```bash
> # 安装 air
> go install github.com/cosmtrek/air@latest
>
> # 启动 (代替 go run)
> air
> ```

启动成功后，访问 [http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html) 验证项目是否成功。

## 🐳 项目部署

### 1. 编译部署

```bash
# 编译适用于当前系统的二进制文件
go build -o youlai-gin main.go

# 运行
./youlai-gin
```

> 💡 **提示**：为了让服务在后台持续运行，你可以使用 `nohup ./youlai-gin &` 命令，或使用 `Systemd` 进行进程守护。

### 2. Docker 部署

```bash
# 构建镜像
docker build -t youlai-gin:latest .

# 运行容器
docker run -d -p 8000:8000 --name youlai-gin youlai-gin:latest
```

## 💖 技术交流

- **问题反馈**：[Gitee Issues](https://gitee.com/youlaiorg/youlai-gin/issues)
- **技术交流**：关注公众号【有来技术】回复“交流群”，或加微信好友【haoxianrui】进微信群
- **官网主页**：[https://www.youlai.tech](https://www.youlai.tech)
