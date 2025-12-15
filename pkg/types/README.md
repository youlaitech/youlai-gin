# Types 包

## LocalTime - 自定义时间类型

### 概述

`LocalTime` 是一个自定义的时间类型，用于统一处理 JSON 序列化和数据库交互中的时间格式。

### 特性

- ✅ **统一格式**: JSON 序列化为 `2006-01-02 15:04:05` 格式
- ✅ **数据库兼容**: 实现 `sql.Scanner` 和 `driver.Valuer` 接口
- ✅ **类型安全**: 基于标准库 `time.Time`，保留所有时间操作能力
- ✅ **零值处理**: 正确处理 `null` 值

### 使用方法

#### 1. 在 VO/DTO 中定义

```go
package model

import "youlai-gin/pkg/types"

type UserVO struct {
    ID         int64            `json:"id"`
    Username   string           `json:"username"`
    CreateTime types.LocalTime  `json:"createTime"`  // ✅ 使用 LocalTime
}
```

#### 2. 数据库查询

GORM 会自动处理 `LocalTime` 类型：

```go
// 查询时自动转换
var user UserVO
db.Table("sys_user").Find(&user)

// JSON 输出格式: {"createTime": "2021-08-03 01:43:26"}
```

#### 3. 手动创建时间

```go
// 获取当前时间
now := types.Now()

// 从标准 time.Time 转换
stdTime := time.Now()
localTime := types.LocalTime(stdTime)

// 转回 time.Time
stdTime = localTime.Time()
```

### 对比

| 类型 | JSON 序列化格式 | 示例 |
|------|---------------|------|
| `time.Time` | ISO 8601 | `2021-08-03T01:43:26+08:00` ❌ |
| `types.LocalTime` | 本地格式 | `2021-08-03 01:43:26` ✅ |
| `string` | 依赖数据库/手动 | 不一致 ⚠️ |

### 最佳实践

1. **VO/DTO 中优先使用 `LocalTime`**
   - 所有需要返回给前端的时间字段
   - 需要统一格式的场景

2. **Entity 中可以使用 `string`**
   - 如果不需要时间计算
   - GORM 的 `autoCreateTime` 标签会自动处理

3. **内部逻辑使用 `time.Time`**
   - 需要时间计算、比较等操作
   - 通过 `.Time()` 方法转换

### 注意事项

- LocalTime 底层是 `time.Time`，可以安全转换
- JSON 序列化和反序列化都支持
- 数据库读写无需额外配置
- 时区使用本地时区 (`time.Local`)

### 示例

```go
// ✅ 推荐用法
type UserPageVO struct {
    ID         int64            `json:"id"`
    Username   string           `json:"username"`
    CreateTime types.LocalTime  `json:"createTime"`  // 自动格式化为友好格式
}

// ❌ 避免用法
type UserPageVO struct {
    ID         int64     `json:"id"`
    Username   string    `json:"username"`
    CreateTime time.Time `json:"createTime"`  // 会输出 ISO 8601 格式
}
```

### 未来扩展

如需要其他时间格式，可以创建类似的类型：

```go
// 仅日期格式: 2006-01-02
type LocalDate time.Time

// 时间戳格式（秒）
type UnixTime time.Time
```
