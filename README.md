# Nimbus Framework Go

Nimbus Framework 的 Go 1.26 模块化单体版。它保留统一的新 UI 与清晰的业务边界，默认使用 MySQL 8.4，面向创业团队的后台和 APP API 脚手架。

## 边界

- `system`：租户、运营后台账号、认证与权限入口。
- `infra`、`member`、`pay`：预留中心边界，当前仅提供健康接口。
- `application`、`im`、`app`：扩展模板，当前仅提供健康接口，不预建业务实体。
- 所有公开 HTTP 接口由 Swagger 注释生成 OpenAPI 文档。

## 本地启动

```bash
./scripts/init-local.sh
export NIMBUS_DB_DSN='nimbus:nimbus_dev@tcp(127.0.0.1:23316)/nimbus_platform_go?charset=utf8mb4&parseTime=True&loc=Local'
make build
./bin/nimbus-server
```

- API: `http://localhost:58080`
- Swagger: `http://localhost:58080/swagger/index.html`
- 默认租户：`Nimbus Framework`
- 默认账号：`admin / admin123`（仅本地初始化；生产环境必须通过环境变量覆盖）

前端复用 `frontend/`，视觉基线与 Java 两版一致。

