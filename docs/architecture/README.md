# Architecture Docs

这里存放 Convey Engine 的技术设计文档。

当前建议按下面方式扩展：

- `system-overview.md`：系统总览与模块边界
- `backend-stack-design.md`：后端技术栈、服务分层与职责边界
- `worker-runtime-design.md`：Worker 进程模型、Task Queue 与执行运行时设计
- `schema-design.md`：数据库与状态模型
- `api-design.md`：对外与对内接口设计

如果文档属于“为什么做这个决策”，而不是“系统长什么样”，优先写入 `../adr/`。
