<div align="center">

# <img alt="youlai-gin" width="28" src="./docs/images/logo/logo.png" valign="middle"> youlai-gin



**Go/Gin 企业级权限管理系统后端**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.11-green?logo=gin)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue?logo=apache)](LICENSE)
[![Gitee Star](https://gitee.com/youlaiorg/youlai-gin/badge/star.svg)](https://gitee.com/youlaiorg/youlai-gin/stargazers)
[![GitHub Star](https://img.shields.io/github/stars/youlaitech/youlai-gin?style=social)](https://github.com/youlaitech/youlai-gin)
[![GitCode Star](https://gitcode.com/youlai/youlai-gin/star/badge.svg)](https://gitcode.com/youlai/youlai-gin/stargazers)

</div>

![](./docs/images/rainbow.png "rainbow.png")

<div align="center">

[![在线预览](https://img.shields.io/badge/在线预览-2D8CF0?style=for-the-badge&logo=google-chrome&logoColor=white)](https://vue.youlai.tech)
[![移动端预览](https://img.shields.io/badge/移动端预览-19BE6B?style=for-the-badge&logo=android&logoColor=white)](https://app.youlai.tech)
[![阅读文档](https://img.shields.io/badge/阅读文档-8B5CF6?style=for-the-badge&logo=gitbook&logoColor=white)](https://www.youlai.tech/docs/server/gin/)
[![English](https://img.shields.io/badge/English-00B4D8?style=for-the-badge&logo=google-translate&logoColor=white)](./README.en.md)

</div>

## 项目简介

**youlai-gin** 是一套基于 Go/Gin 的企业级权限管理系统后端，配套前端 [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) 和移动端 [youlai-app](https://gitee.com/youlaiorg/youlai-app)，并提供 **7 种语言实现**（Java / Node.js / Go / Python / PHP / C# / Rust），共享同一套 API 规范与数据库结构。适用于企业中后台管理系统的学习参考与二次开发。

## 核心特性

- 🔐 **安全体系** — JWT + Redis Token 双会话模式、令牌续期、多端互斥
- 🛡️ **细粒度权限** — RBAC 权限模型，菜单/按钮/接口统一治理
- ⚡ **代码生成器** — 一键生成前后端 CRUD 代码
- 📦 **模块齐全** — 用户、角色、菜单、部门、字典、文件、消息中心、操作日志
- 🔌 **实时通信** — SSE 推送：在线用户数、字典同步、通知广播

## 系统预览

**PC 端**

<table align="center">
  <tr>
    <td><img alt="PC预览1" width="400" src="./docs/images/preview/pc-01.png"></td>
    <td><img alt="PC预览2" width="400" src="./docs/images/preview/pc-02.png"></td>
  </tr>
  <tr>
    <td><img alt="PC预览3" width="400" src="./docs/images/preview/pc-03.png"></td>
    <td><img alt="PC预览4" width="400" src="./docs/images/preview/pc-04.png"></td>
  </tr>
  <tr>
    <td><img alt="PC预览5" width="400" src="./docs/images/preview/pc-05.png"></td>
    <td><img alt="PC预览6" width="400" src="./docs/images/preview/pc-06.png"></td>
  </tr>
</table>

**移动端**

<table align="center">
  <tr>
    <td><img alt="APP预览1" width="200" src="./docs/images/preview/app-01.png"></td>
    <td><img alt="APP预览2" width="200" src="./docs/images/preview/app-02.png"></td>
    <td><img alt="APP预览3" width="200" src="./docs/images/preview/app-03.png"></td>
    <td><img alt="APP预览4" width="200" src="./docs/images/preview/app-04.png"></td>
  </tr>
</table>

## 快速开始

**环境要求**：Go 1.25+ · MySQL 8.0+ · Redis 7.x+

1. 克隆项目：
   ```bash
   git clone https://gitee.com/youlaiorg/youlai-gin.git
   ```

2. 导入数据库：
   ```bash
   mysql -u root -p youlai_admin < sql/mysql/youlai_admin.sql
   ```

3. 修改配置（可选，默认已配置线上只读数据源）：
   ```bash
   # 编辑 configs/dev.yaml 调整数据库连接等
   ```

4. 安装依赖：
   ```bash
   go mod tidy
   ```

5. 启动服务：

   **方式一：GoLand 启动（推荐）**
   用 GoLand 打开项目，等待依赖索引完成，点击 `main.go` 旁的运行按钮或 `go run` 配置即可。

   **方式二：命令行启动**
   ```bash
   go run main.go
   ```
   启动后访问 [http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)，能打开接口文档即说明后端已正常运行。

6. 启动前端（可选）：
   如需可视化操作界面，启动配套前端 [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin)，访问 [http://localhost:3000](http://localhost:3000)，使用 `admin` / `123456` 登录。

> 更多内容详见官方文档：[快速开始](https://www.youlai.tech/docs/server/gin/quick-start.html) · [部署指南](https://www.youlai.tech/docs/server/gin/deploy.html)

## 目录结构

```
youlai-gin/
├── internal/                       # 私有应用代码
│   ├── auth/                       # 认证模块（登录/Token/会话）
│   ├── codegen/                    # 代码生成模块
│   ├── common/                     # 公共模块（数据库/Redis/权限/工具）
│   ├── file/                       # 文件管理模块
│   ├── message/                    # SSE 消息推送
│   ├── middleware/                 # 中间件（JWT/CORS/限流）
│   ├── router/                     # 路由注册
│   └── system/                     # 系统模块（用户/角色/菜单/部门）
├── pkg/                            # 可被外部使用的公共库
│   ├── constant/                   # 常量定义
│   ├── enums/                      # 枚举类型
│   ├── errs/                       # 统一错误类型
│   ├── model/                      # 通用模型（分页/选项/实体）
│   └── types/                      # 自定义类型（BigInt/LocalTime）
├── configs/                        # 配置文件（dev/prod/test）
├── sql/                            # 数据库初始化脚本
├── main.go                         # 应用入口
└── Dockerfile                      # Docker 镜像构建文件
```

## 生态矩阵

**前端**

| 项目 | 技术栈 | 说明 | 更新状态 |
|:-----|:-------|:-----|:---------|
| [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) | Vue 3 + Element Plus | PC 管理前端（主推） | ✅️ |
| [youlai-app](https://gitee.com/youlaiorg/youlai-app) | Vue 3 + UniApp | 移动端 App | ✅️ |

**后端**

| 项目 | 技术栈 | 说明 | 更新状态 |
|:-----|:-------|:-----|:---------|
| [youlai-boot](https://gitee.com/youlaiorg/youlai-boot) | Spring Boot + MyBatis-Plus | Java（主推） | ✅️ |
| [youlai-nest](https://gitee.com/youlaiorg/youlai-nest) | NestJS + TypeORM | Node.js | ✅️ |
| [youlai-gin](https://gitee.com/youlaiorg/youlai-gin) | Go + Gorm | Go | ✅️ |
| [youlai-django](https://gitee.com/youlaiorg/youlai-django) | Django + DRF | Python | ✅️ |
| [youlai-fastapi](https://gitee.com/youlaiorg/youlai-fastapi) | FastAPI + SQLAlchemy | Python | ✅️ |
| [youlai-think](https://gitee.com/youlaiorg/youlai-think) | ThinkPHP + ThinkORM | PHP | ✅️ |
| [youlai-aspnet](https://gitee.com/youlaiorg/youlai-aspnet) | ASP.NET Core + EF Core | C# | ✅️ |
| [youlai-axum](https://gitee.com/youlaiorg/youlai-axum) | Axum + SeaORM | Rust | ✅️ |
| [youlai-laravel](https://gitee.com/youlaiorg/youlai-laravel) | Laravel + Eloquent | PHP | ❌️ |

> 九种后端共享同一套 **RESTful API 规范** 和 **数据库结构**，前端可无缝切换。

**衍生版本**

| 项目 | 基于 | 类型 | 说明 | 更新状态 |
|:-----|:-----|:-----|:-----|:---------|
| [youlai-boot-tenant](https://gitee.com/youlaiorg/youlai-boot-tenant) | youlai-boot | 独立仓库 | 多租户 SaaS，租户隔离与租户配置 | ✅️ |
| [youlai-boot-flex](https://gitee.com/youlaiorg/youlai-boot-flex) | youlai-boot | 独立仓库 | 改用 MyBatis-Flex | ✅️ |
| [youlai-boot (db-pg)](https://gitee.com/youlaiorg/youlai-boot/tree/db-pg) | youlai-boot | 分支 | PostgreSQL 数据库分支 | ✅️ |
| [youlai-boot (multi-module)](https://gitee.com/youlaiorg/youlai-boot/tree/multi-module) | youlai-boot | 分支 | 多模块工程拆分 | ✅️ |
| [youlai-nest (multi-tenant)](https://gitee.com/youlaiorg/youlai-nest/tree/multi-tenant) | youlai-nest | 分支 | 多租户 SaaS，租户隔离与租户配置 | ✅️ |
| [youlai-boot (spring-boot-3)](https://gitee.com/youlaiorg/youlai-boot/tree/spring-boot-3) | youlai-boot | 分支 | Spring Boot 3 兼容分支 | ❌️ |

## 技术合作

本项目采用 [Apache License 2.0](LICENSE) 开源，可免费商用。欢迎在 [Issue](https://gitee.com/youlaiorg/youlai-gin/issues) 提交问题或反馈，也欢迎提交 [Pull Request](https://gitee.com/youlaiorg/youlai-gin/pulls) 共建。

如需技术支持、商务合作、二次开发、项目定制或私有化部署，可联系作者微信（见下方二维码）。

<table align="center">
  <tr>
    <td align="center">
      <img src="./docs/images/qrcode/wechat-official.jpg" height="180" alt="公众号「有来技术」"><br>
      <sub>公众号「有来技术」</sub>
    </td>
    <td>&nbsp;&nbsp;&nbsp;&nbsp;</td>
    <td align="center">
      <img src="./docs/images/qrcode/wechat-mp.jpg" height="180" alt="小程序「有来技术」"><br>
      <sub>小程序「有来技术」</sub>
    </td>
    <td>&nbsp;&nbsp;&nbsp;&nbsp;</td>
    <td align="center">
      <img src="./docs/images/qrcode/wechat-personal.png" height="180" alt="添加作者微信"><br>
      <sub>添加作者微信</sub>
    </td>
  </tr>
</table>

<p align="center"><em>技术交流 · 问题反馈 · 商务合作</em></p>
