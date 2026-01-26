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

## 🚀 快速启动

### 1. 环境准备

| 要求 | 说明 | 安装指引 |
| --- | --- | --- |
| **Go** | 1.21+ | [官方下载](https://go.dev/dl/) |
| **MySQL** | 5.7+ 或 8.x | 业务数据存储，必需安装：[Windows](https://youlai.blog.csdn.net/article/details/133272887) / [Linux](https://youlai.blog.csdn.net/article/details/130398179) |
| **Redis** | 7.x 稳定版 | 会话缓存，必需安装：[Windows](https://youlai.blog.csdn.net/article/details/133410293) / [Linux](https://youlai.blog.csdn.net/article/details/130439335) |

> ⚠️ **重要提示**：MySQL 与 Redis 为项目启动必需依赖，请确保服务已启动。

### 2. 数据库初始化

推荐使用 **Navicat**、**DBeaver** 或 **MySQL Workbench** 执行 `scripts/mysql/youlai_admin.sql` 脚本，完成数据库和基础数据的初始化。

### 3. 修改配置

编辑开发环境配置文件 `configs/dev.yaml`，根据实际情况修改 MySQL 和 Redis 的连接信息。

### 4. 启动项目

```bash
# 下载依赖
go mod tidy

# 启动服务
go run main.go
```

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
