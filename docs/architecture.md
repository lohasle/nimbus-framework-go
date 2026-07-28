# 架构说明

```text
Vue 运营后台
      |
  /admin-api
      |
 nimbus-server
      |
 +-- system
 +-- infra
 +-- member
 +-- pay
 +-- application / im / app (Health)
      |
 +----+----+
 |         |
MySQL 8.4 Redis 7.4
主数据存储   监控/缓存基础设施
```

后端采用 Gin、GORM 与模块化单体结构：

- `backend/cmd/server`：进程装配、启动与优雅停机。
- `backend/internal/platform`：配置、数据库、HTTP、中间件和路由。
- `backend/internal/modules`：按业务中心隔离模型、处理器和路由。
- `backend/docs`：Swagger 生成物。

认证使用独立的 Access Token 与 Refresh Token。Access Token 用于业务请求；Refresh Token 仅用于轮换新的令牌对。

前端在 [`frontend/src/config/axios/service.ts`](../frontend/src/config/axios/service.ts) 统一处理认证失效：

1. 普通请求收到 HTTP 401 或业务码 401 后，共享同一个刷新任务；多标签页通过 Web Locks 串行刷新。
2. 刷新成功后，每个原请求只携带新令牌重试一次。
3. Refresh Token 缺失、刷新失败或重试仍为 401 时，立即清理令牌、用户缓存与动态路由。
4. 退登过程具备幂等性，只提示一次，并直接跳转 `/login?redirect=<原页面>`，不再重载受保护页面。

System 与 Infra 复用 `muse-app-go` 已验证的通用实现，Nimbus 保留自己的包名、品牌、菜单和 Member/Pay 模块。Go 后端仍为单进程模块化单体，Redis 不构成新的业务微服务。
