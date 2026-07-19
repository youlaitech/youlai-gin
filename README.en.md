<div align="center">



# <img alt="youlai-gin" width="28" src="./docs/images/logo/logo.png" align="center"> youlai-gin



**Enterprise-grade permission management backend based on Go/Gin**



[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)

[![Gin](https://img.shields.io/badge/Gin-1.11-green?logo=gin)](https://gin-gonic.com/)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue?logo=apache)](LICENSE)

[![Gitee Star](https://gitee.com/youlaiorg/youlai-gin/badge/star.svg)](https://gitee.com/youlaiorg/youlai-gin/stargazers)

[![GitHub Star](https://img.shields.io/github/stars/youlaitech/youlai-gin?style=social)](https://github.com/youlaitech/youlai-gin)

[![GitCode Star](https://gitcode.com/youlai/youlai-gin/star/badge.svg)](https://gitcode.com/youlai/youlai-gin/stargazers)



</div>



![](https://foruda.gitee.com/images/1708618984641188532/a7cca095_716974.png "rainbow.png")



<div align="center">



[![Live Preview](https://img.shields.io/badge/Live%20Preview-2D8CF0?style=for-the-badge&logo=google-chrome&logoColor=white)](https://vue.youlai.tech)

[![Mobile Preview](https://img.shields.io/badge/Mobile%20Preview-19BE6B?style=for-the-badge&logo=android&logoColor=white)](https://app.youlai.tech)

[![Documentation](https://img.shields.io/badge/Documentation-8B5CF6?style=for-the-badge&logo=gitbook&logoColor=white)](https://www.youlai.tech/docs/server/gin/)

[![简体中文](https://img.shields.io/badge/简体中文-00B4D8?style=for-the-badge&logo=google-translate&logoColor=white)](./README.md)



</div>



## Introduction



**youlai-gin** is an enterprise-grade permission management backend built on Go/Gin. It ships with the frontend [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) and the mobile app [youlai-app](https://gitee.com/youlaiorg/youlai-app), and is one of **7 language implementations** (Java / Node.js / Go / Python / PHP / C# / Rust) that share the same API specification and database schema. It is suitable for learning, reference, and secondary development of enterprise admin systems.



## Core Features



- 🔐 **Security** — JWT + Redis dual-session model, token renewal, multi-device mutual exclusion

- 🛡️ **Fine-grained permissions** — RBAC model governing menus, buttons, and APIs in one place

- ⚡ **Code generator** — one-click generation of full-stack CRUD code

- 📦 **Complete modules** — users, roles, menus, departments, dictionaries, files, message center, operation logs

- 🔌 **Real-time communication** — SSE push: online user count, dictionary sync, notification broadcast



## System Preview



**PC**



<table align="center">

  <tr>

    <td><img alt="PC Preview 1" width="400" src="./docs/images/preview/pc-01.png"></td>

    <td><img alt="PC Preview 2" width="400" src="./docs/images/preview/pc-02.png"></td>

  </tr>

  <tr>

    <td><img alt="PC Preview 3" width="400" src="./docs/images/preview/pc-03.png"></td>

    <td><img alt="PC Preview 4" width="400" src="./docs/images/preview/pc-04.png"></td>

  </tr>

  <tr>

    <td><img alt="PC Preview 5" width="400" src="./docs/images/preview/pc-05.png"></td>

    <td><img alt="PC Preview 6" width="400" src="./docs/images/preview/pc-06.png"></td>

  </tr>

</table>



**Mobile**



<table align="center">

  <tr>

    <td><img alt="App Preview 1" width="200" src="./docs/images/preview/app-01.png"></td>

    <td><img alt="App Preview 2" width="200" src="./docs/images/preview/app-02.png"></td>

    <td><img alt="App Preview 3" width="200" src="./docs/images/preview/app-03.png"></td>

    <td><img alt="App Preview 4" width="200" src="./docs/images/preview/app-04.png"></td>

  </tr>

</table>



## Quick Start



**Requirements**: Go 1.25+ · MySQL 8.0+ · Redis 7.x+



1. Clone: `git clone https://gitee.com/youlaiorg/youlai-gin.git`

2. Import database: `sql/mysql/youlai_admin.sql`

3. Adjust config (optional, a read-only online data source is configured by default): `configs/dev.yaml`

4. Install dependencies: `go mod tidy`

5. Start: `go run main.go`, then visit http://localhost:8000/swagger/index.html



Default credentials: `admin` / `123456`



> 💡 **Hot reload**: install `air` with `go install github.com/cosmtrek/air@latest`, then run `air`



Detailed guide: [Deployment Docs](https://www.youlai.tech/docs/server/gin/deploy)



## Tech Stack



| Tech | Version | Description |

|:-----|:--------|:------------|

| Go | 1.25+ | Core language |

| Gin | 1.11 | Web framework |

| GORM | — | ORM |

| MySQL | 5.7+ / 8.x | Database |

| Redis | 7.x+ | Cache · Session |

| Swagger | — | API docs |



## Directory Structure



```

youlai-gin/

├── internal/                       # Private application code

│   ├── auth/                       # Auth (login/token/session)

│   ├── codegen/                    # Code generation

│   ├── common/                     # Common module (db/Redis/permission/util)

│   ├── file/                       # File management

│   ├── message/                    # SSE push

│   ├── middleware/                 # Middleware (JWT/CORS/rate limit)

│   ├── router/                     # Route registration

│   └── system/                     # System module (user/role/menu/dept)

├── pkg/                            # Public libraries

│   ├── constant/                   # Constants

│   ├── enums/                      # Enums

│   ├── errs/                       # Unified error types

│   ├── model/                      # Common models (pagination/options/entities)

│   └── types/                      # Custom types (BigInt/LocalTime)

├── configs/                        # Config files (dev/prod/test)

├── sql/                            # Database init scripts

├── main.go                         # Application entry

└── Dockerfile                      # Docker image build

```



## Ecosystem



**Frontend**



| Project | Stack | Description |

|:-----|:------|:------------|

| [vue3-element-admin](https://gitee.com/youlaiorg/vue3-element-admin) | Vue 3 + Element Plus | PC admin frontend (recommended) |

| [youlai-app](https://gitee.com/youlaiorg/youlai-app) | Vue 3 + UniApp | Mobile App |



**Backend**



| Project | Stack | Description |
| [youlai-boot](https://gitee.com/youlaiorg/youlai-boot) | Spring Boot + MyBatis-Plus | Java (recommended) |
| [youlai-nest](https://gitee.com/youlaiorg/youlai-nest) | NestJS + TypeORM | Node.js |
| [youlai-gin](https://gitee.com/youlaiorg/youlai-gin) | Go + Gorm | Go |
| [youlai-django](https://gitee.com/youlaiorg/youlai-django) | Django + DRF | Python |
| [youlai-fastapi](https://gitee.com/youlaiorg/youlai-fastapi) | FastAPI + SQLAlchemy | Python |
| [youlai-think](https://gitee.com/youlaiorg/youlai-think) | ThinkPHP + ThinkORM | PHP |
| [youlai-aspnet](https://gitee.com/youlaiorg/youlai-aspnet) | ASP.NET Core + EF Core | C# |
| [youlai-axum](https://gitee.com/youlaiorg/youlai-axum) | Axum + SeaORM | Rust |
> **youlai-boot** also provides the following variants and branches: [Multi-Tenant](https://gitee.com/youlaiorg/youlai-boot-tenant) · [MyBatis-Flex](https://gitee.com/youlaiorg/youlai-boot-flex) · [Spring Boot 3](https://gitee.com/youlaiorg/youlai-boot/tree/spring-boot-3) · [PostgreSQL](https://gitee.com/youlaiorg/youlai-boot/tree/db-pg) · [Multi-Module](https://gitee.com/youlaiorg/youlai-boot/tree/multi-module)

>

> The eight backends share the same **RESTful API specification** and **database schema**, so the frontend can switch seamlessly.



## Documentation



| Resource | Link |

|:-----|:-----|

| 📖 Full docs site | [www.youlai.tech](https://www.youlai.tech/) |

| 🖥️ PC live preview | [vue.youlai.tech](https://vue.youlai.tech) |

| 📱 Mobile live preview | [app.youlai.tech](https://app.youlai.tech) |

| 🔗 Apifox API docs | [apifox.com](https://www.apifox.cn/apidoc/shared-195e783f-4d85-4235-a038-eec696de4ea5) |

| 🔗 Local API docs | [localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html) |



## Contributing



Issues and Pull Requests are welcome! See the [Contribution Guide](https://www.youlai.tech/faq/help).



## License



Released under the [Apache License 2.0](LICENSE); free for commercial use.



---



<table align="center">

  <tr>

    <td align="center">

      <img src="./docs/images/qrcode/wechat-official.jpg" height="180" alt="Official WeChat Account"><br>

      <sub>Official WeChat Account</sub>

    </td>

    <td>&nbsp;&nbsp;&nbsp;&nbsp;</td>

    <td align="center">

      <img src="./docs/images/qrcode/wechat-mp.jpg" height="180" alt="Mini Program"><br>

      <sub>Mini Program</sub>

    </td>

    <td>&nbsp;&nbsp;&nbsp;&nbsp;</td>

    <td align="center">

      <img src="./docs/images/qrcode/wechat-personal.png" height="180" alt="Add author on WeChat"><br>

      <sub>Add author on WeChat</sub>

    </td>

  </tr>

</table>



<p align="center"><em>Technical discussion · Feedback · Business cooperation</em></p>

