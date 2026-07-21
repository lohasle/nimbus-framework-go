# Nimbus Framework Go

Nimbus Framework 的 Go 1.26 模块化单体版。它以 Java Nimbus 底座为功能源进行迁移，保留统一的新 UI 与兼容的前端接口契约，默认使用 MySQL 8.4。

工程按前后端分层：`frontend/` 是 Nimbus Vue 运营后台，`backend/` 是 Go 后端；Go 后端在 `internal/modules/` 下按 System、Infra、Member、Pay 等中心划分模块，公共技术能力位于 `internal/platform/`。

## 边界

- `system`：租户识别、运营后台账号、认证、权限菜单、字典与通知入口。
- `infra`：参数配置、文件存储配置、API 访问日志。
- `member`：会员、等级、分组、标签与积分流水。
- `pay`：支付应用、渠道、订单与退款核心管理闭环。
- `application`、`im`、`app`：扩展模板，当前仅提供健康接口，不预建业务实体。
- 所有公开 HTTP 接口由 Swagger 注释生成 OpenAPI 文档。

## 本地启动

```bash
./scripts/init-local.sh
export NIMBUS_DB_DSN='nimbus:nimbus_dev@tcp(127.0.0.1:23316)/nimbus_platform_go?charset=utf8mb4&parseTime=True&loc=Local'
cd backend
make test build
./bin/nimbus-server

# 另开终端启动前端
cd frontend
pnpm install --frozen-lockfile
pnpm dev
```

- API: `http://localhost:58080`
- Swagger: `http://localhost:58080/swagger/index.html`
- 默认租户：`Nimbus Framework`
- 默认账号：`admin / admin123`（仅本地初始化；生产环境必须通过环境变量覆盖）

前端地址为 `http://localhost:3000`。视觉基线与 Java 两版一致；只有后端语言和部署形态不同。
