# gobkd

Go 后端。单体。最小分层。

技术栈：

- `gin`
- `sqlite3`
- `go-pkgz/auth/v2`
- `logrus`

## 架构

请求链路：

`gin route -> middleware -> handler -> service -> repository -> sqlite`

规则：

- `handler` 只处理 HTTP、参数、响应
- `service` 放业务逻辑
- `repository` 只写数据库
- 认证、日志、配置在启动时统一初始化
- 不用 ORM，不做复杂 DI，不拆多服务

## 目录

```text
cmd/server/main.go           # 启动入口
internal/app                 # 组装 router、db、auth、logger
internal/auth                # go-pkgz/auth 封装
internal/config              # 配置读取，自动加载 .env
internal/handler             # HTTP handler
internal/service             # 业务逻辑
internal/repository          # sqlite 访问
internal/middleware          # 请求中间件
internal/requestx            # 参数绑定和校验
internal/apperr             # 业务错误和错误码映射
internal/response            # 统一响应和错误码
internal/model               # 数据结构
migrations                   # SQL migration
```

## 已实现

路由：

- `GET /ping`
- `GET /healthz`
- `ANY /auth/*`
- `GET /api/v1/me`
- `POST /api/v1/echo`

中间件：

- `RequestID`
- `SecurityHeaders`
- `Recovery`
- `RequestLogger`
- `RequestBodyLimit`
- `Auth Trace`
- `RequireUser`

## 认证

认证库：`go-pkgz/auth/v2`

当前接法：

- 使用 `local` direct provider
- 本地账号来自 `AUTH_LOCAL_USER`
- 本地密码来自 `AUTH_LOCAL_PASS`
- `/auth/*` 交给认证库
- `/api/v1/*` 先走 `Trace`，再走 `RequireUser`

拿当前用户：

- `internal/auth.Service.CurrentUser`

## 数据库

数据库访问用 `database/sql + sqlite3`。

初始化流程：

1. 打开 sqlite
2. 执行 `migrations/*.sql`
3. repository 直接写 SQL
4. 多步写操作通过事务封装执行

事务入口：

- `repository.Transactor.WithinTransaction`
- `repository.(*UserRepository).WithTx`

当前表：

- `users`

## 配置

程序启动会自动读取当前目录的 `.env`。

环境变量：

```env
APP_ENV=dev
HTTP_ADDR=:8080
APP_BASE_URL=http://127.0.0.1:8080
SQLITE_PATH=./data/app.db
AUTH_SECRET=
AUTH_LOCAL_USER=
AUTH_LOCAL_PASS=
LOG_LEVEL=info
HTTP_READ_TIMEOUT=10s
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
HTTP_SHUTDOWN_TIMEOUT=10s
HTTP_MAX_HEADER_BYTES=1048576
HTTP_MAX_BODY_BYTES=1048576
HTTP_TRUSTED_PROXIES=
```

## 统一响应

成功响应直接返回 JSON。

错误响应统一格式：

```json
{
  "error": "validation_failed",
  "message": "validation failed",
  "details": [
    {
      "field": "Message",
      "rule": "required"
    }
  ],
  "request_id": "..."
}
```

错误码：

- `invalid_request`
- `validation_failed`
- `unauthorized`
- `forbidden`
- `not_found`
- `method_not_allowed`
- `conflict`
- `payload_too_large`
- `internal_error`
- `service_unavailable`

业务错误建议：

- `service` / `repository` 返回 `internal/apperr`
- `handler` 统一走 `response.FromError`
- 不要在每个 handler 里手写状态码分支

## 参数校验

统一入口：

- `requestx.BindJSON`
- `requestx.BindQuery`
- `requestx.BindURI`

规则：

- 在 handler 里绑定参数
- 校验写在 struct tag
- 校验失败返回 `422 validation_failed`

示例：

```go
type EchoRequest struct {
    Message string `json:"message" binding:"required,max=200"`
}
```

## 给 AI 的开发约定

新增业务时按这个结构写：

1. 在 `internal/model` 定义数据结构
2. 在 `internal/repository` 写 SQL
3. 在 `internal/service` 写业务逻辑
4. 在 `internal/handler` 写接口
5. 在 `internal/app/app.go` 注册路由
6. 需要表时，在 `migrations` 增加 SQL

不要做：

- 不要把 SQL 写进 handler
- 不要把 gin context 传进 repository
- 不要先引入 ORM
- 不要为了抽象而加 interface
- 不要把简单功能拆成很多 package

## 本地运行

```bash
cp .env.example .env
# edit .env and fill AUTH_SECRET / AUTH_LOCAL_USER / AUTH_LOCAL_PASS
go run ./cmd/server
```

检查：

```bash
curl http://127.0.0.1:8080/ping
curl http://127.0.0.1:8080/healthz
```

登录：

```bash
curl -X POST 'http://127.0.0.1:8080/auth/local/login?session=1' \
  -H 'Content-Type: application/json' \
  -d '{"user":"<your-user>","passwd":"<your-pass>"}' \
  -c cookies.txt
```

读取当前用户：

```bash
curl http://127.0.0.1:8080/api/v1/me -b cookies.txt
```

## Linux build 和部署

前提：

- Linux
- `go`
- `gcc` 或 `cc`，这个项目用 `sqlite3`，构建需要 CGO
- `systemd`
- `curl`

准备环境：

```bash
cp .env.example .env
# edit .env
```

部署脚本：

```bash
chmod +x ./scripts/deploy_linux.sh
./scripts/deploy_linux.sh
```

默认行为：

- 构建 `./cmd/server`
- 发布到 `/opt/gobkd/releases/<timestamp>`
- 持久化 `.env` 到 `/opt/gobkd/shared/.env`
- 持久化 sqlite 数据到 `/opt/gobkd/shared/data`
- 切换 `/opt/gobkd/current`
- 如果系统里已有 `gobkd.service`，自动重启并检查 `http://127.0.0.1:8080/healthz`

可改环境变量：

```bash
APP_NAME=gobkd
SERVICE_NAME=gobkd
INSTALL_ROOT=/opt/gobkd
HEALTHCHECK_URL=http://127.0.0.1:8080/healthz
RESTART_SERVICE=auto
```

`systemd` 可参考：

- `deploy/gobkd.service.example`

卸载脚本：

```bash
chmod +x ./scripts/uninstall_linux.sh
CONFIRM_UNINSTALL=YES ./scripts/uninstall_linux.sh
```

默认会：

- 停掉并禁用 `gobkd.service`
- 删除 `/opt/gobkd`

如果连 service 文件也要删：

```bash
REMOVE_SERVICE_FILE=1 CONFIRM_UNINSTALL=YES ./scripts/uninstall_linux.sh
```
