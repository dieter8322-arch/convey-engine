# Docs Guide

`docs/` 用于存放所有面向项目本身的长期文档，不放临时讨论记录。

## 目录说明

### `architecture/`

存放系统架构、模块边界、数据模型、接口设计等技术文档。

适合放：

- 系统总览
- 模块拆分说明
- 数据库与 schema 设计
- API 设计草案

### `product/`

存放产品目标、MVP 范围、需求说明、路线图等文档。

适合放：

- MVP 计划
- 版本路线图
- 需求说明

### `adr/`

存放 Architecture Decision Records，用于记录关键技术决策与取舍。

命名建议：

- `YYYY-MM-DD-topic.md`

### `operations/`

存放部署、运行、发布、排障、值班等操作性文档。

适合放：

- 本地开发环境启动说明
- 部署流程
- 故障排查手册

## 文档命名建议

- 目录名统一使用英文小写
- 文件名统一使用英文 `kebab-case`
- 一个文件聚焦一个主题，避免“大而全”文档无限增长

## 当前文档

- `architecture/system-overview.md`：当前系统总体设计与 MVP 范围草案
- `architecture/backend-stack-design.md`：后端技术栈、分层与存储边界设计
- `architecture/worker-runtime-design.md`：Worker 进程、Task Queue 与执行运行时设计
- `architecture/schema-design.md`：业务数据库 schema、索引与迁移约束
- `architecture/web-frontend-implementation-plan.md`：图形化优先的流水线前端实施计划
