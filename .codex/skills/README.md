# Project Skills

项目级 skill 统一放在 `.codex/skills/`。

## 目录结构

```text
.codex/skills/
└── <skill-name>/
    ├── SKILL.md
    ├── references/
    ├── scripts/
    └── assets/
```

## 约定

- 目录名使用英文 `kebab-case`
- 一个 skill 解决一个明确问题
- `SKILL.md` 为必需文件
- `references/`、`scripts/`、`assets/` 按需创建，不强制

## 适合放进项目级 skill 的内容

- 仓库特有流程
- 团队约定的实现套路
- 需要复用的排障、评审、发布动作
- 开源协作约定，例如统一的 REUSE / SPDX 文件头规范

不适合：

- 通用开发知识
- 只服务某一次讨论的临时提示词
