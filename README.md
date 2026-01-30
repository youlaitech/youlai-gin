<div align="center">
  <img alt="logo" width="100" height="100" src="https://foruda.gitee.com/images/1733417239320800627/3c5290fe_716974.png">
  <h2>youlai-gin</h2>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.21+-blue.svg"/>
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

`youlai-gin` 是 `vue3-element-admin` 的 Go/Gin 后端实现，接口路径与返回结构完全对齐，可直接为前端提供后端服务。

- **🚀 技术栈**：Go 1.21+ + Gin + GORM，轻量高性能组合
- **🔐 安全认证**：JWT 无状态认证 + Redis 会话管理，支持会话治理
- **🔑 权限管理**：RBAC 权限模型，菜单/按钮/接口三级权限统一治理
- **🛠️ 模块能力**：用户、角色、菜单、部门、字典、日志等核心模块开箱即用

## 🌈 项目源码

| 项目 | Gitee | GitHub | GitCode |
| --- | --- | --- | --- |
| ✅ Go 后端 | [youlai-gin](https://gitee.com/youlaiorg/youlai-gin) | [youlai-gin](https://github.com/youlaitech/youlai-gin) | [youlai-gin](https://gitcode.com/youlai/youlai-gin) |
| Vue3 管理端 | [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) | [vue3-element-admin](https://github.com/youlaitech/vue3-element-admin) | [vue3-element-admin](https://gitcode.com/youlai/vue3-element-admin) |
| uni-app 移动端 | [vue-uniapp-template](https://gitee.com/youlaiorg/vue-uniapp-template) | [vue-uniapp-template](https://github.com/youlaitech/vue-uniapp-template) | [vue-uniapp-template](https://gitcode.com/youlai/vue-uniapp-template) |

## 📚 项目文档

| 文档名称           | 访问地址                                                                 |
| ------------------ | ------------------------------------------------------------------------ |
| 项目介绍与使用指南 | [https://www.youlai.tech/youlai-gin](https://www.youlai.tech/youlai-gin) |

## 📁 项目目录

<details>
<summary>目录结构</summary>

```text
youlai-gin/
├─ configs/                   # 配置文件 (dev/prod)
├─ docs/                      # 项目文档
├─ examples/                  # 示例代码
├─ internal/                  # 核心业务源码
│  ├─ auth/                   # 认证模块(登录/Token/会话)
│  ├─ health/                 # 健康检查
│  ├─ platform/               # 平台模块(文件/扩展能力)
│  ├─ router/                 # 路由注册
│  └─ system/                 # 系统模块(用户/角色/菜单等)
├─ pkg/                       # 通用包 (中间件/响应等)
├─ scripts/                   # 数据库脚本
├─ Dockerfile                 # Docker 镜像构建文件
├─ go.mod                     # 依赖管理
├─ go.sum                     # 依赖版本锁定
└─ main.go                    # 应用入口
```

</details>

## 🚀 快速启动

### 1. 环境准备

| 技术 | 版本/说明 | 安装文档 |
| --- | --- | --- |
| **Go** | `1.25` 或更高版本 | [官方下载](https://go.dev/dl/) |
| **MySQL** | `5.7` 或 `8.x` | [Windows](https://youlai.blog.csdn.net/article/details/133272887) / [Linux](https://youlai.blog.csdn.net/article/details/130398179) |
| **Redis** | `7.x` | [Windows](https://youlai.blog.csdn.net/article/details/133410293) / [Linux](https://youlai.blog.csdn.net/article/details/130439335) |

> ⚠️ **重要提示**：MySQL 与 Redis 为项目启动必需依赖，请确保服务已启动。

### 2. 初始化数据库

使用数据库客户端（如 Navicat、DBeaver）执行 `scripts/mysql/youlai_admin.sql` 脚本，完成数据库和基础数据的初始化。

### 3. 修改配置

编辑 `configs/dev.yaml` 文件，根据实际情况修改 MySQL 和 Redis 的连接字符串。

### 4. 启动项目

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

## 🤝 前端整合

`youlai-gin` 与 `vue3-element-admin` 前后端协议完全兼容，可无缝对接。

```bash
# 1. 获取前端项目
git clone https://gitee.com/youlaiorg/vue3-element-admin.git
cd vue3-element-admin

# 2. 安装依赖
pnpm install

# 3. 配置后端地址 (编辑 .env.development)
VITE_APP_API_URL=http://localhost:8000

# 4. 启动前端
pnpm run dev
```

- **访问地址**: [http://localhost:3000](http://localhost:3000)
- **登录账号**: `admin` / `123456`

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
