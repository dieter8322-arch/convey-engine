# Convey Engine

Convey Engine 是一个面向开源方向演进的轻量流水线执行引擎，目标是提供比重型 CI 平台更易部署、比零散脚本更易管理的构建、测试、部署编排能力。

当前仓库仍处于早期设计阶段，代码尚未开始落地，但目录已经按“前后端同仓、一个项目”的方式预留好单仓结构，避免后续演进时反复搬目录。

## 仓库结构

```text
.
├── apps/                   # 前后端应用
│   ├── web/                # 前端应用
│   └── server/             # 后端服务
├── packages/               # 共享代码与契约
│   └── shared/
├── .codex/                 # 项目共享 AI 资产
│   ├── README.md
│   └── skills/             # 项目级 skills
├── docs/                   # 项目文档
│   ├── README.md
│   ├── architecture/       # 架构与技术设计
│   ├── product/            # 产品范围、路线图、需求说明
│   ├── adr/                # Architecture Decision Records
│   └── operations/         # 运维、发布、部署与运行手册
├── CONTRIBUTING.md
└── LICENSE
```

## 文档入口

- 架构总览：`docs/architecture/system-overview.md`
- 文档规范：`docs/README.md`
- 协作规范：`CONTRIBUTING.md`
- AI 资产规范：`.codex/README.md`
- 前后端目录说明：`apps/README.md`

## 当前原则

- 设计文档统一放在 `docs/`
- 前端与后端代码统一放在同一仓库
- 应用层放在 `apps/`，共享层放在 `packages/`
- 项目共享 AI skills 统一放在 `.codex/skills/`
- 文件级版权与许可证元数据遵循 REUSE / SPDX 约定
- 新增目录前优先复用现有分类，避免顶层目录膨胀
- 文件命名优先使用英文 kebab-case，正文可使用中文
