# Project AI Assets

`.codex/` 用于存放仓库共享的 AI 协作资产，避免把这类内容散落到 `docs/` 或顶层目录。

## 当前约定

- `skills/`：项目共享 skills

后续如果确实需要，再按主题增加：

- `prompts/`：长期复用的提示模板
- `context/`：供 agent 复用的领域背景材料

## 原则

- AI 资产服务于仓库协作，不替代正式产品文档
- 能写进 `docs/` 的项目知识，优先写进 `docs/`
- 只有对 agent 工作流有直接价值的内容，才进入 `.codex/`
