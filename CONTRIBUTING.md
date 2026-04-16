# Contributing

欢迎为 Convey Engine 提交设计、文档、代码和项目级 AI 资产。

## 目录约定

- 前端与后端统一维护在同一仓库
- 前端应用放在 `apps/web/`
- 后端服务放在 `apps/server/`
- 共享代码与契约放在 `packages/`
- 架构设计放在 `docs/architecture/`
- 产品范围、需求与路线图放在 `docs/product/`
- 重要架构决策记录放在 `docs/adr/`
- 运维、部署、发布与排障文档放在 `docs/operations/`
- 项目共享 AI skills 放在 `.codex/skills/<skill-name>/`

## 文档约定

- 文件名使用英文 `kebab-case`
- 文档正文默认可使用简体中文
- 同一主题优先增量更新已有文档，而不是重复创建相似文档
- ADR 建议使用 `YYYY-MM-DD-topic.md` 命名
- 跨前后端共享内容优先沉淀到 `packages/`，避免重复定义

## AI Skills 约定

- 一个 skill 一个目录
- 每个 skill 必须包含 `SKILL.md`
- 可按需增加 `references/`、`scripts/`、`assets/`
- skill 只解决一个明确问题，避免把多类职责塞进同一 skill

## 开源文件头规范

仓库默认采用 REUSE / SPDX 风格管理文件版权与许可证信息，并以 MIT 作为当前项目许可证。

基本规则：

- 新增一方源码文件优先使用标准 SPDX 头，而不是自定义 `Created by`
- 一方源码文件默认包含：
  - `SPDX-FileCopyrightText`
  - `SPDX-License-Identifier`
- 当前项目一方文件默认使用统一版权主体：`2026 dieter8322-arch`
- 许可证标识默认使用：`MIT`
- 完整许可证文本以根目录 `LICENSE` 与 `LICENSES/MIT.txt` 为准

源码示例：

```go
// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package main
```

```ts
// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

export function example() {}
```

补充规则：

- 若文件存在多个版权主体，可写多行 `SPDX-FileCopyrightText`
- 若复制或改写第三方代码，必须保留并补充上游版权与许可证信息
- 生成文件应保留生成器声明；版权标签通常仍按项目规则标注
- 不再使用 `Author:`、`Created by:`、系统用户名或本地隐藏配置作为默认头信息
- 不便直接加头的文件，可通过 `REUSE.toml` 做批量标注

当前项目级 skill：

- `.codex/skills/open-source-file-header/`

## 提交建议

- 目录调整应同步更新对应 `README.md`
- 新增设计文档时，优先在 `docs/README.md` 中补入口
- 新增项目级 skill 时，优先参考 `.codex/skills/_template/`
