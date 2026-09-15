# LexCase（律师案件管理系统）

面向律师事务所的一体化管理平台，支持客户管理、案件全流程跟踪、文档归档和费用结算。

## 快速启动（Docker Compose 一键部署）

```bash
cp .env.example .env
docker compose up -d --build
```

启动完成后访问：

- 前端：http://localhost:28031
- 后端 API：http://localhost:29069
- 后端健康检查：http://localhost:29069/healthz
- PostgreSQL：localhost:38102

预置账号（密码见 database/init.sql 与 backend/internal/service/seed.go）：

| 用户名 | 密码 | 角色 |
| --- | --- | --- |
| admin | Admin@123 | 管理员 |
| lawyer | User@123 | 律师 |
| assistant | User@123 | 助理 |

## 本地开发

后端：

```bash
cd backend && go mod tidy && go run ./cmd/server
```

构建：`cd backend && go build ./...`

前端：

```bash
cd frontend && npm install && npm run dev
```

前端开发服务器通过 Vite 代理将 `/api` 转发到 `http://localhost:29069`。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | React 18 + TypeScript + Ant Design + Vite + Zustand |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 16 |
| 认证 | JWT（github.com/golang-jwt/jwt/v5）+ RBAC |
| 其他依赖 | gin-contrib/cors、golang.org/x/crypto/bcrypt、go-playground/validator/v10 |

## 项目目录结构

```
cy-402/
├── docker-compose.yml
├── .env.example
├── database/init.sql              # PostgreSQL 初始化脚本（建表 + 种子数据）
├── backend/
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/
│       ├── model/                 # user/client/case/document/billing/audit_log
│       ├── repository/            # 按实体分文件
│       ├── service/               # 业务逻辑 + 种子数据 + 单元测试
│       ├── handler/               # HTTP 处理器（含 upload_handler、audit_log_handler）
│       ├── router/                # router.go + 按实体分文件
│       ├── middleware/            # auth/rbac/rate_limiter/error_handler/audit_log/cors/request_logger
│       ├── dto/
│       ├── constants/             # 枚举、错误码、日志模板、文案
│       └── util/                  # jwt/logger/formatters/amount_formatter/app_error/file_upload
└── frontend/
    └── src/
        ├── api/                   # auth/user/client/case/document/billing/auditLog/upload
        ├── stores/                # authStore/userStore/clientStore/caseStore/documentStore/billingStore
        ├── types/
        ├── components/common/     # CaseCard/DocumentList/StatusBadge/TimelineItem/AmountSummary/ClientCard/CaseTable/BillingCard/DocumentCard/FileUploader/FilterBar/AvatarUploader/PermissionGuard
        ├── hooks/                 # useAuth/usePagination/useFileUpload/usePermission
        ├── pages/                 # Cases/CaseDetail/Clients/Billing/Documents/Profile/AuditLogs/Login
        ├── router/                # index.tsx + guards.tsx
        ├── utils/                 # dateFormat/amountFormatter/request
        └── constants/             # case/billing/document/errorCodes
```

## 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | Docker Compose 项目名/容器前缀 | cylawcase |
| DB_NAME | 数据库名 | cylawcase_db |
| DB_USER | 数据库用户 | cylawcase_user |
| DB_PASSWORD | 数据库密码 | cylawcase_pwd |
| DB_ROOT_PASSWORD | 预留 root 密码项 | cylawcase_root |
| JWT_SECRET | JWT 签名密钥 | change_me_to_a_long_random_string |
| JWT_EXPIRE_HOURS | JWT 过期小时数 | 72 |
| APP_CORS_ORIGINS | 允许的跨域来源（逗号分隔） | http://localhost:28031 |
| FRONTEND_PORT | 前端端口 | 28031 |
| BACKEND_PORT | 后端端口 | 29069 |
| DB_PORT | 数据库端口 | 38102 |

## Docker 部署说明

- 端口映射：前端 28031:80、后端 29069:8080、PostgreSQL 38102:5432。
- 数据卷：`db-data` 持久化 PostgreSQL 数据；`upload-data` 持久化上传文件。
- 服务依赖：backend `depends_on` db（service_healthy），frontend `depends_on` backend（service_healthy）。
- 前端 Nginx 将 `/api/` 反代到 `http://backend:8080/`，支持 SPA 路由 `try_files`。
- 常见问题：
  - 端口冲突：修改 `.env` 中 `FRONTEND_PORT/BACKEND_PORT/DB_PORT` 后重新 `docker compose up -d`。
  - 数据库重置：`docker compose down -v` 后重新启动。

## 枚举出现位置清单

### CaseStatus（filed/investigating/hearing/closed/archived）
- 后端：`backend/internal/constants/case.go`、`backend/internal/model/case.go`、`backend/internal/service/case_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`backend/internal/dto/dto_case.go`、`database/init.sql`
- 前端：`frontend/src/constants/case.ts`、`frontend/src/components/common/StatusBadge.tsx`、`frontend/src/pages/Cases.tsx`、`frontend/src/pages/CaseDetail.tsx`、`frontend/src/components/common/CaseCard.tsx`、`frontend/src/components/common/CaseTable.tsx`

### BillingType（attorney_fee/court_fee/travel_fee/other）
- 后端：`backend/internal/constants/billing.go`、`backend/internal/model/billing.go`、`backend/internal/service/billing_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`backend/internal/dto/dto_billing.go`、`database/init.sql`
- 前端：`frontend/src/constants/billing.ts`、`frontend/src/components/common/BillingCard.tsx`、`frontend/src/pages/Billing.tsx`

### BillingStatus（pending/paid/invoiced/void）
- 后端：`backend/internal/constants/billing.go`、`backend/internal/model/billing.go`、`backend/internal/service/billing_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`database/init.sql`
- 前端：`frontend/src/constants/billing.ts`、`frontend/src/components/common/StatusBadge.tsx`、`frontend/src/components/common/AmountSummary.tsx`、`frontend/src/components/common/BillingCard.tsx`、`frontend/src/pages/Billing.tsx`

## API 接口清单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /healthz | 服务健康检查 |
| GET | /api/healthz | Nginx 反代健康检查 |
| GET | /api/v1/healthz | API 版本健康检查 |
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录，返回 JWT |
| GET | /api/v1/users/me | 当前登录用户信息 |
| PUT | /api/v1/users/me | 修改个人资料 |
| GET | /api/v1/users | 用户列表（仅管理员） |
| GET | /api/v1/clients | 客户分页列表 |
| POST | /api/v1/clients | 新建客户 |
| GET | /api/v1/clients/:id | 客户详情与历史案件 |
| PUT | /api/v1/clients/:id | 编辑客户 |
| DELETE | /api/v1/clients/:id | 删除客户 |
| GET | /api/v1/cases | 案件分页列表 |
| POST | /api/v1/cases | 创建案件 |
| GET | /api/v1/cases/:id | 案件详情 |
| PUT | /api/v1/cases/:id | 更新案件 |
| POST | /api/v1/cases/:id/status | 案件状态流转 |
| POST | /api/v1/cases/:id/assign | 分配主办律师 |
| GET | /api/v1/documents | 文档中心分页列表 |
| POST | /api/v1/documents | 上传文档记录 |
| GET | /api/v1/documents/by-case/:id | 按案件查询文档 |
| DELETE | /api/v1/documents/:id | 删除文档 |
| GET | /api/v1/billings | 账单分页列表 |
| POST | /api/v1/billings | 创建账单 |
| GET | /api/v1/billings/summary | 本月应收/已收/待收汇总 |
| GET | /api/v1/billings/by-case/:id | 按案件查询账单 |
| POST | /api/v1/billings/:id/paid | 标记支付 |
| POST | /api/v1/billings/:id/invoiced | 标记开票 |
| POST | /api/v1/billings/:id/void | 作废账单 |
| GET | /api/v1/audit-logs | 审计日志（仅管理员） |
| POST | /api/v1/upload/file | 文件上传 |

归档费用收口相关错误码（均为 HTTP 409 Conflict）：

| code | 常量 | 含义 |
| --- | --- | --- |
| 40903 | CodeCaseArchivePendingBilling | 案件存在待支付账单，拒绝归档，状态与结案日期不变 |
| 40904 | CodeBillingCreateArchivedCase | 案件已归档，禁止再新增待支付账单 |

并发语义：归档（`POST /cases/:id/status` 传 `archived`）与新建账单（`POST /billings`）在数据库事务内通过案件行排他锁串行化；两个请求同时到达时只会得到“已结案+待支付账单”或“已归档+无待支付账单”之一，对同一终态重复请求结果不变（归档幂等、归档后建账恒为 40904）。

业务编号发号（案件 `case_no` / 账单 `bill_no`）：

- 编号由 PostgreSQL 序列 `case_no_seq`、`bill_no_seq` 通过事务内 `nextval` 分配，服务启动时 `CREATE SEQUENCE IF NOT EXISTS` 幂等创建，并自动对齐历史编号的最大数字后缀。
- 唯一性在多次启动、多个实例（水平扩容）、同一数据库连续运行下均成立，不依赖进程内计数器，无需清库或人工改号；历史时间戳编号与新序列编号格式互不冲突。
- 编号撞唯一索引时：整事务回滚（本次 `nextval` 作废、绝不复用旧号，回滚同时释放案件行锁），用全新序列值重新取号、重新加锁并复核归档终态后重试；成功严格只落一条记录。业务错误（如 40904）不重试、原样返回。
- 取号失败或多次冲突仍未成功时，稳定返回可重试错误（HTTP 503，错误码 50301）且不落任何记录；待序列恢复后重试同一请求即可。

| code | 常量 | HTTP | 含义 |
| --- | --- | --- | --- |
| 50301 | CodeNumberAllocateFailed | 503 | 编号生成临时失败（可重试），请求未产生任何数据 |

后端并发集成测试位于 `backend/internal/service/archive_closure_test.go`（归档/建账竞态）与 `backend/internal/service/numbering_test.go`（多实例编号唯一、冲突重试、缺序列 50301、升级对齐），通过 `CY_TEST_DB_DSN` 指向 PostgreSQL 后运行（如 `go test ./internal/service/ -race`）；未配置 DSN 时自动跳过，纯逻辑表驱动测试始终运行。

## 主要功能

- 客户管理：新建/编辑/检索客户，查看历史案件。
- 案件管理：创建案件、状态流转（立案→调查→庭审→结案→归档）、律师分配、筛选查询。
- 归档前费用收口：案件由“已结案”进入“已归档”时，只要仍存在 `pending`（待支付）账单即拒绝归档（HTTP 409，错误码 40903），案件状态与结案日期保持不变；待支付账单补齐支付（paid）或作废（void）后才能再次归档。案件归档后不得再新增待支付账单（HTTP 409，错误码 40904）。归档与新建账单并发到达时，后端在同一事务内对案件行加排他锁串行处理，冲突只会收敛为“已结案且挂待支付账单”或“已归档且无待支付账单”两种稳定终态，重复重试/重复归档结果不变，不会出现“已归档案件仍挂待支付账单”。
- 文档归档：按案件上传/查看/删除文档（起诉状/答辩状/证据/判决书/合同等）。
- 费用结算：创建账单、标记支付、开票、作废，本月应收/已收/待收汇总。
- 审计日志：写操作自动记录（管理员查看）。
- 角色权限：JWT + RBAC（admin/lawyer/assistant）。

## License

MIT License
