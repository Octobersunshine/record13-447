# Session Management API

一个使用 Go 语言开发的会话管理服务，提供通过用户 ID 冻结/解冻用户当前登录会话的功能。

## 问题修复记录

### Bug 修复：冻结/解冻逻辑持久化同步问题

**问题描述**：
原冻结逻辑仅更新内存会话缓存，持久化会话库状态未同步更新，导致服务重启后冻结状态丢失。

**修复方案**：
1. 新增持久化存储层（JSON 文件存储）
2. 新增缓存层，采用 **Write-Through** 写策略：先写持久化存储，成功后再更新缓存
3. 实现缓存失效机制，确保数据一致性
4. 启动时自动从持久化存储加载数据到缓存

## 项目结构

```
session-management/
├── cmd/
│   └── server/
│       └── main.go                  # 主程序入口
├── internal/
│   ├── model/
│   │   └── session.go               # 数据模型定义
│   ├── store/
│   │   ├── session_store.go         # 内存存储（缓存层）
│   │   ├── persistent_store.go      # 持久化存储（JSON 文件）
│   │   └── cached_store.go          # 带缓存的组合存储层
│   ├── service/
│   │   └── session_service.go       # 业务逻辑层
│   ├── handler/
│   │   └── session_handler.go       # HTTP 处理器
│   └── router/
│       └── router.go                # 路由配置
├── data/
│   └── sessions.json                # 持久化数据文件（自动创建）
├── go.mod
├── go.sum
└── README.md
```

## 核心功能

1. **创建会话** - 为用户创建新的登录会话
2. **冻结会话** - 根据用户 ID 冻结该用户的所有活跃会话（双写：持久化+缓存）
3. **解冻会话** - 根据用户 ID 解冻该用户的所有冻结会话（双写：持久化+缓存）
4. **查询用户会话** - 查询指定用户的所有会话
5. **验证会话** - 验证会话 token 是否有效
6. **列出所有会话** - 列出系统中的所有会话
7. **刷新缓存** - 手动从持久化存储刷新缓存

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

### 7. 刷新缓存

**POST** `/api/v1/sessions/admin/cache/refresh`

响应：
```json
{
  "message": "cache refreshed successfully from persistent storage"
}
```

## 架构说明

### 存储架构（修复后）

```
┌─────────────────────────────────────────────────────┐
│                  CachedSessionStore                 │
│  ┌──────────────────┐    ┌──────────────────────┐  │
│  │  In-Memory Cache │◄──►│  Persistent Storage  │  │
│  │  (session_store) │    │  (persistent_store)  │  │
│  └──────────────────┘    └──────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

#### 写策略（Write-Through）：
1. **冻结/解冻操作** → 先更新持久化存储 → 成功后更新内存缓存 → 失效用户级缓存
2. **保证**：持久化存储失败时，整个操作失败，不会出现数据不一致

#### 读策略（Cache-Aside）：
1. 先查内存缓存
2. 缓存未命中时查持久化存储
3. 回填缓存并返回

### 分层架构

1. **Model 层** ([session.go](file:///e:/temp/record13/447/internal/model/session.go))
   - 定义会话数据结构、请求/响应模型
   - 定义会话状态枚举（active, frozen, expired）

2. **Store 层**：
   - **[session_store.go](file:///e:/temp/record13/447/internal/store/session_store.go)** - 内存存储实现（缓存层）
   - **[persistent_store.go](file:///e:/temp/record13/447/internal/store/persistent_store.go)** - 持久化存储（JSON 文件）
   - **[cached_store.go](file:///e:/temp/record13/447/internal/store/cached_store.go)** - 组合存储层，协调双写和缓存
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

存储层使用 `sync.RWMutex` 读写锁保证并发安全：
- 读操作使用 `RLock()` 允许多个 goroutine 同时读取
- 写操作使用 `Lock()` 确保独占访问

### 数据一致性保证

1. **原子写操作**：持久化存储使用临时文件+原子重命名，确保写入不损坏
2. **双写顺序**：先写持久化，成功后再更新缓存
3. **缓存失效**：写操作后主动失效相关缓存，避免脏读
4. **启动加载**：服务启动时从持久化存储全量加载数据到缓存
5. **手动刷新**：提供 API 接口支持手动刷新缓存

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DATA_DIR` | 持久化数据存储目录 | `./data` |

## 扩展建议

1. **Redis 持久化**：当前使用 JSON 文件存储，建议替换为 Redis 以支持分布式部署

2. **认证中间件**：为 `/admin/*` 路由添加管理员认证中间件，确保只有授权用户才能执行冻结/解冻操作

3. **操作日志**：添加审计日志，记录每次冻结/解冻操作的操作员、时间、原因等信息

4. **会话过期**：实现自动过期清理机制，定期清理过期会话

5. **单元测试**：为各层添加单元测试，确保业务逻辑正确性

6. **分布式锁**：如果部署多个实例，需要添加分布式锁防止并发冻结/解冻冲突

## 使用示例

```bash
# 创建会话
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "ip": "192.168.1.1"}'

# 冻结用户会话（双写：持久化+缓存）
curl -X POST http://localhost:8080/api/v1/sessions/admin/freeze \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "reason": "安全原因"}'

# 解冻用户会话（双写：持久化+缓存）
curl -X POST http://localhost:8080/api/v1/sessions/admin/unfreeze \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "reason": "问题解决"}'

# 查询用户会话
curl http://localhost:8080/api/v1/sessions/user/user123

# 手动刷新缓存
curl -X POST http://localhost:8080/api/v1/sessions/admin/cache/refresh

# 查看持久化数据文件
cat ./data/sessions.json
```

## 冻结/解冻操作流程

### 冻结操作流程
```
1. 接收冻结请求（user_id, reason）
2. 调用持久化存储 FreezeByUserID
   ├─ 遍历所有会话，找到该用户的 active 会话
   ├─ 更新状态为 frozen
   └─ 原子写入 JSON 文件
3. 持久化成功后，更新内存缓存中的会话状态
4. 失效该用户的查询缓存
5. 全量清空缓存防止脏数据
6. 返回冻结结果
```

### 解冻操作流程
```
1. 接收解冻请求（user_id, reason）
2. 调用持久化存储 UnfreezeByUserID
   ├─ 遍历所有会话，找到该用户的 frozen 会话
   ├─ 更新状态为 active
   └─ 原子写入 JSON 文件
3. 持久化成功后，更新内存缓存中的会话状态
4. 失效该用户的查询缓存
5. 全量清空缓存防止脏数据
6. 返回解冻结果
```

