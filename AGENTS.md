# Nimbus Framework Go Agent Guide

- Go 版本以 `go.mod` 为准，使用稳定版，不使用预发布工具链。
- 默认数据库必须保持 MySQL 8.4；其他数据库只能作为显式可选适配。
- 未经 SPEC 确认，不新增业务实体。扩展模块默认只保留健康接口。
- 新增或修改 REST API 必须同步 Swagger 注释，并执行 `make swagger`。
- 提交前至少执行 `go test ./...`、`make build` 和前端生产构建。
