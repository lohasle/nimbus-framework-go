# Nimbus Framework Go 文档

## 阅读顺序

1. [项目 README](../README.md)
2. [Agent 上下文](agent-context.md)
3. [架构说明](architecture.md)
4. [开发规范](development-standards.md)
5. [测试策略](testing-strategy.md)
6. [SPEC 索引](specs/README.md)

## 目录约定

- `docs/`：人工维护的架构、开发、测试与 SPEC 文档。
- `backend/docs/`：Swagger 自动生成物，不放人工文档。
- `.rule/`：跨迭代稳定规则。
- `.agents/skills/`：仅存放 Nimbus 特有的可复用工作指引。
- `scripts/`：本地初始化、启动和验收脚本。

文档必须描述仓库当前真实能力。未开放菜单或未实现的 Java 扩展能力不得标记为已完成。
