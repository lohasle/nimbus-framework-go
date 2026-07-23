# Nimbus Framework Go Agent Guide

本仓库采用 Spec 驱动的最小变更流程。开始非简单修改前按顺序阅读：

1. `docs/README.md`
2. `docs/agent-context.md`
3. `docs/development-standards.md`
4. `docs/specs/README.md`
5. 当前生效的 `docs/specs/SPEC-*.md`

稳定规则位于 `.rule/`。工作流程为：

```text
Think -> Spec -> Plan -> Build -> Review -> Test/QA -> Ship -> Reflect
```

- Go 版本以 `backend/go.mod` 为准，使用稳定版，不使用预发布工具链。
- 数据库固定使用 MySQL 8.4；未经明确需求不增加其他数据库驱动。
- 未经 SPEC 确认，不新增或删除模块、业务实体、菜单与服务进程。
- `application`、`im`、`app` 默认只保留 Health，不预建业务实现。
- 新增或修改 REST API 必须同步 Swagger 注释和路由契约测试，并执行 `make swagger`。
- 后端代码位于 `backend/`；模块代码位于 `internal/modules/<module>`，公共能力位于 `internal/platform`。
- 不覆盖无关工作区修改。提交前至少执行 `go test ./...`、`make build`、前端类型检查和生产构建。
