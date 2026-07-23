# 开发规范

## 变更流程

非简单变更先更新 SPEC，写明目标、非目标、接口和验收标准。实现应优先复用现有平台层，不创建重复抽象。

## 后端

- 包路径固定为 `github.com/lohasle/nimbus-framework-go`。
- 模块不得直接修改其他模块的数据；跨模块事务必须在装配层明确表达。
- 数据写入必须检查错误，金额、余额和积分变更必须使用事务。
- API 变更同步 Swagger 和路由契约测试。
- 日志使用 `slog` 结构化字段，不打印口令、Token 或支付密钥。

## 前端

- 保留原有表格和业务页面组件体系。
- 品牌修改集中在 Logo、文案、登录页和 Loading。
- 当前菜单未开放的页面不作为后端迁移清单。

## 验证

```bash
cd backend
go test ./...
make build

cd ../frontend
pnpm ts:check
pnpm lint
pnpm build:local
```
