# API 与安全规则

- API 变更同步 Swagger 注释与路由契约测试。
- Access Token 与 Refresh Token 不得混用。
- 生产环境必须覆盖默认 JWT 密钥和初始化密码。
- 日志不得记录口令、完整 Token、支付密钥或个人敏感信息。
- 返回结构保持运营后台现有 `/admin-api` 契约兼容。
