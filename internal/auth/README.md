# Auth 认证模块

参考 `youlai-boot` 的 `AuthController` 实现 Go 版本的认证接口。

## 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| DELETE | `/api/v1/auth/logout` | 退出 |
| POST | `/api/v1/auth/refresh-token` | 刷新令牌 |

## 快速测试

```bash
# 登录
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

## 技术栈

- **密码加密**: bcrypt
- **会话管理**: JWT / Redis Token
- **用户查询**: `internal/user/repository`

详细说明见：[认证登录接口文档](../../../vue3-element-admin-docs/backend/go/getting-started.md#10-认证登录接口)
