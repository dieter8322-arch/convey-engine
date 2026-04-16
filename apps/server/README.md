# Server App

这里预留给 Convey Engine 的后端服务。

建议后续承载：

- API Server
- Scheduler
- Worker 相关进程入口
- 后端业务逻辑与持久化实现

当前已经开始按 Go 单仓应用骨架落地：

- `cmd/api`：Gin API 入口
- `cmd/worker`：Temporal worker 入口
- `internal/api`：HTTP 路由与 handler
- `internal/app`：应用服务与 `convey.yaml` 解析
- `internal/repository`：GORM 持久化实现
- `internal/workflow`：Temporal workflow / activity / starter
- `internal/storage`：对象存储抽象与 local 实现
- `internal/config`：环境变量配置

当前阶段只覆盖后端最小闭环：

- 手动触发 run
- 解析并摘要 `convey.yaml`
- 写入 `projects / pipeline_defs / pipeline_versions / runs`
- 启动最小 `PipelineRunWorkflow`
- 查询 run 状态
