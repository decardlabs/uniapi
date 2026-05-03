# 项目 API 接口文档（自动生成）

---

## 1. API 端点总览

### 用户模块
- `POST /api/user/register`  用户注册
- `POST /api/user/login`  用户登录
- `GET /api/user/logout`  用户登出
- `GET /api/user/self`  获取当前用户信息
- `PUT /api/user/`  更新用户信息（需认证）
- `DELETE /api/user/:id`  删除用户（需管理员）
- `GET /api/user`  用户列表（需管理员）
- `GET /api/user/search`  用户搜索（需管理员）
- `POST /api/user/manage`  用户权限/状态变更（需管理员）

### 通道模块
- `GET /api/channel`  通道列表（需管理员）
- `GET /api/channel/:id`  通道详情（需管理员）
- `POST /api/channel/`  新增通道（需管理员）
- `PUT /api/channel/`  更新通道（需管理员）
- `DELETE /api/channel/:id`  删除通道（需管理员）
- `GET /api/channel/types`  通道类型（公开）

### 日志模块
- `GET /api/log`  日志列表（需认证）
- `GET /api/log/search`  日志搜索（需认证）
- `DELETE /api/log/`  删除日志（需认证）

### Token（API Key）模块
- `GET /api/token`  Token 列表（需认证）
- `POST /api/token/`  新建 Token（需认证）
- `PUT /api/token/`  更新 Token（需认证）
- `DELETE /api/token/:id`  删除 Token（需认证）

### 充值模块
- `GET /api/topup`  充值请求列表（需管理员）
- `POST /api/topup/`  新建充值请求（需认证）
- `PUT /api/topup/`  审核充值（需管理员）

### 其它
- `GET /api/status`  服务状态
- `GET /api/models`  支持模型列表（需认证）
- `GET /api/models/display`  公共模型展示
- `GET /api/notice`  公告
- `GET /api/about`  关于

---

## 2. 接口详情（示例：用户注册）

### POST /api/user/register
- **描述**：新用户注册
- **请求参数**：
  - body: `{ username: string, password1: string, password2: string, email?: string }`
- **请求示例**：
  - curl:
    ```bash
    curl -X POST https://yourdomain/api/user/register \
      -H 'Content-Type: application/json' \
      -d '{"username":"test","password1":"12345678","password2":"12345678","email":"test@example.com"}'
    ```
  - axios:
    ```js
    axios.post('/api/user/register', { username: 'test', password1: '12345678', password2: '12345678', email: 'test@example.com' })
    ```
- **响应结构**：
  - 成功：
    ```json
    { "success": true, "data": { "id": 1, "username": "test", ... }, "message": "" }
    ```
  - 失败：
    ```json
    { "success": false, "message": "用户名已存在" }
    ```
- **错误码说明**：
  - 用户名已存在、参数不合法等
- **认证要求**：无需认证

---

## 3. 前端调用对照表

| 页面/组件 | 调用接口 |
|---|---|
| 登录页 | /api/user/login |
| 注册页 | /api/user/register |
| 用户管理 | /api/user、/api/user/search、/api/user/manage |
| 通道管理 | /api/channel、/api/channel/:id、/api/channel/types |
| 日志管理 | /api/log、/api/log/search |
| Token 管理 | /api/token、/api/token/:id |
| 充值管理 | /api/topup、/api/topup/:id |

---

## 4. 核心数据模型

### User
- id: int
- username: string
- password: string
- display_name: string
- role: int
- status: int
- email: string
- quota: int64
- group: string
- access_token: string
- totp_secret: string
- mcp_tool_blacklist: string[]
- created_at: int64
- updated_at: int64

### Channel
- id: int
- type: int
- key: string
- status: int
- name: string
- models: string
- group: string
- balance: float64
- config: string
- ...

### Log
- id: int
- user_id: int
- created_at: int64
- type: int
- content: string
- token_name: string
- model_name: string
- quota: int
- ...

### Token
- id: int
- user_id: int
- key: string
- status: int
- name: string
- remain_quota: int64
- unlimited_quota: bool
- expired_time: int64
- models: string
- subnet: string
- ...

---

## 5. 错误码与响应结构

- 统一响应结构：
  - 成功：`{ "success": true, "data": ..., "message": "" }`
  - 失败：`{ "success": false, "message": "错误信息" }`
- 典型错误码：
  - 认证失败、参数错误、权限不足、配额不足、服务异常等
- 代理/AI接口错误结构：
  ```json
  {
    "error": { "message": "详细错误信息", "type": "upstream", "code": "bad_response", ... },
    "message": "..."
  }
  ```

---

> 本文档由自动化工具生成，详细参数、响应字段、错误码可参考源码 controller/model/dto 目录。