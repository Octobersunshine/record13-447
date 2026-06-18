# Session Management API

一个使用 Go 语言开发的会话管理服务，提供通过用户 ID 冻结/解冻用户当前登录会话的功能。

## 项目结构

```
session-management/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── internal/
│   ├── model/
│   │   └── session.go           # 数据模型定义
│   ├── store/
│   │   └── session_store.go     # 数据存储层（内存存储）
│   ├── service/
│   │   └── session_service.go   # 业务逻辑层
│   ├── handler/
│   │   └── session_handler.go   # HTTP 处理器
│   └── router/
│       └── router.go            # 路由配置
├── go.mod
├── go.sum
└── README.md
```

## 核心功能

1. **创建会话** - 为用户创建新的登录会话
2. **冻结会话** - 根据用户 ID 冻结该用户的所有活跃会话
3. **解冻会话** - 根据用户 ID 解冻该用户的所有冻结会话
4. **查询用户会话** - 查询指定用户的所有会话
5. **验证会话** - 验证会话 token 是否有效
6. **列出所有会话** - 列出系统中的所有会话

## 技术栈

- **Go 1.21+**
- **Gin Web Framework** - HTTP 路由和中间件
- **UUID** - 生成唯一会话 ID 和 Token

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API 接口

所有接口的基础路径为 `/api/v1/sessions`

### 1. 创建会话

**POST** `/api/v1/sessions`

请求体：
```json
{
  "user_id": "user123",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0",
  "device": "Chrome on Windows"
}
```

响应：
```json
{
  "session": {
    "id": "uuid",
    "user_id": "user123",
    "token": "token-uuid",
    "status": "active",
    "ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0",
    "device": "Chrome on Windows",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "expires_at": "0001-01-01T00:00:00Z"
  },
  "message": "session created successfully"
}
```

### 2. 冻结用户会话

**POST** `/api/v1/sessions/admin/freeze`

请求体：
```json
{
  "user_id": "user123",
  "reason": "异常行为检测"
}
```

响应：
```json
{
  "sessions": [
    {
      "id": "uuid",
      "user_id": "user123",
      "token": "token-uuid",
      "status": "frozen",
      ...
    }
  ],
  "total_count": 2,
  "update_count": 2,
  "message": "successfully frozen user sessions"
}
```

### 3. 解冻用户会话

**POST** `/api/v1/sessions/admin/unfreeze`

请求体：
```json
{
  "user_id": "user123",
  "reason": "问题已解决"
}
```

响应：
```json
{
  "sessions": [
    {
      "id": "uuid",
      "user_id": "user123",
      "token": "token-uuid",
      "status": "active",
      ...
    }
  ],
  "total_count": 2,
  "update_count": 2,
  "message": "successfully unfrozen user sessions"
}
```

### 4. 查询用户会话

**GET** `/api/v1/sessions/user/:user_id`

响应：
```json
{
  "sessions": [
    {
      "id": "uuid",
      "user_id": "user123",
      "status": "active",
      ...
    }
  ],
  "total_count": 2,
  "message": "sessions retrieved successfully"
}
```

### 5. 验证会话

**POST** `/api/v1/sessions/validate`

请求头：
```
Authorization: <session-token>
```

响应（成功）：
```json
{
  "session": { ... },
  "message": "session is valid"
}
```

响应（失败/冻结）：
```json
{
  "error": "invalid or frozen session",
  "message": "session not found"
}
```

### 6. 列出所有会话

**GET** `/api/v1/sessions`

## 架构说明

### 分层架构

1. **Model 层** ([session.go](file:///e:/temp/record13/447/internal/model/session.go))
   - 定义会话数据结构、请求/响应模型
   - 定义会话状态枚举（active, frozen, expired）

2. **Store 层** ([session_store.go](file:///e:/temp/record13/447/internal/store/session_store.go))
   - 数据存储接口定义
   - 内存存储实现（使用 sync.RWMutex 保证并发安全）
   - 核心方法：`FreezeByUserID`、`UnfreezeByUserID`

3. **Service 层** ([session_service.go](file:///e:/temp/record13/447/internal/service/session_service.go))
   - 业务逻辑层
   - 封装冻结/解冻操作，返回操作统计信息
   - 会话验证逻辑，检查会话状态是否为 active

4. **Handler 层** ([session_handler.go](file:///e:/temp/record13/447/internal/handler/session_handler.go))
   - HTTP 请求处理
   - 参数绑定和验证
   - 错误处理和响应格式化

5. **Router 层** ([router.go](file:///e:/temp/record13/447/internal/router/router.go))
   - 路由注册和分组
   - API 版本管理

### 并发安全

内存存储使用 `sync.RWMutex` 读写锁保证并发安全：
- 读操作使用 `RLock()` 允许多个 goroutine 同时读取
- 写操作使用 `Lock()` 确保独占访问

## 扩展建议

1. **持久化存储**：当前使用内存存储，重启后数据会丢失。建议集成 Redis 或数据库进行持久化。

2. **认证中间件**：为 `/admin/*` 路由添加管理员认证中间件，确保只有授权用户才能执行冻结/解冻操作。

3. **操作日志**：添加审计日志，记录每次冻结/解冻操作的操作员、时间、原因等信息。

4. **会话过期**：实现自动过期清理机制，定期清理过期会话。

5. **单元测试**：为各层添加单元测试，确保业务逻辑正确性。

## 使用示例

```bash
# 创建会话
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "ip": "192.168.1.1"}'

# 冻结用户会话
curl -X POST http://localhost:8080/api/v1/sessions/admin/freeze \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "reason": "安全原因"}'

# 解冻用户会话
curl -X POST http://localhost:8080/api/v1/sessions/admin/unfreeze \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "reason": "问题解决"}'

# 查询用户会话
curl http://localhost:8080/api/v1/sessions/user/user123
```

