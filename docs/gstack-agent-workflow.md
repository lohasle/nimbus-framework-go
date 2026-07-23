# Agent 交付流程

```text
Think -> Spec -> Plan -> Build -> Review -> Test/QA -> Ship -> Reflect
```

- Think：核对真实代码、运行状态和用户边界。
- Spec：记录目标、非目标、接口与验收。
- Plan：拆分可独立验证的最小步骤。
- Build：实施局部修改，不覆盖无关变更。
- Review：检查安全、兼容性、模块边界和冗余。
- Test/QA：执行后端、前端、数据库和浏览器验证。
- Ship：审查差异后提交并推送。
- Reflect：更新文档中的完成状态与剩余风险。
