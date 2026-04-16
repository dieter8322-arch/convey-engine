# Backend Stack Design

- 状态：Draft v0.1
- 日期：2026-04-16
- 适用范围：Convey Engine 后端基础设施与服务边界

## 1. 目标

本文用于细化 Convey Engine 后端基础技术栈与职责边界，统一下面几个关键决策：

1. 工作流编排使用 Temporal。
2. 业务数据访问使用 GORM。
3. 业务 schema 迁移使用 Atlas。
4. 业务数据库默认 PostgreSQL，同时保持 MySQL 兼容。
5. 对象存储同时支持本地文件系统与 S3 兼容存储。

本文关注“后端怎么分层、各组件负责什么”，不展开到字段级 DDL；字段级设计见 [schema-design.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/schema-design.md)。

## 2. 选型结论

### 2.1 Go + Gin

- Go 作为主语言，满足单二进制、自托管、执行器适配和并发 I/O 场景。
- Gin 作为 HTTP 框架，负责 REST API、middleware、参数绑定、错误处理与路由组织。
- API 层保持薄，避免将长流程逻辑堆进 handler。

### 2.2 Temporal

- Temporal 负责长流程编排、超时、重试、人工批准等待与失败恢复。
- Workflow 负责“流程决策”和“编排状态”，Activity 负责“外部副作用”。
- API Server 通过 Temporal Client 启动 Workflow，并通过 Signal / Update 控制运行状态。
- 这意味着 Convey Engine 不再自研核心 scheduler / claim / retry 内核。

### 2.3 GORM + Atlas

- GORM 负责业务实体映射、查询封装和 repository 层实现。
- Atlas 负责业务 schema 的版本化迁移。
- 生产环境不依赖 GORM `AutoMigrate` 做正式迁移。
- GORM 只操作业务数据库，不接管 Temporal 内部表。

### 2.4 PostgreSQL / MySQL

- 业务数据库默认 PostgreSQL。
- 首版 schema 按“PostgreSQL 优先、MySQL 兼容”约束设计。
- MySQL 兼容的目标是保证主业务路径、常规 CRUD、常规索引与分页查询可用，而不是追求所有高级 SQL 特性等价。

### 2.5 Local / S3 Object Storage

- `local` 适用于开发环境、单机环境和最小部署。
- `s3` 适用于生产环境与多节点部署。
- 两者通过统一对象存储接口屏蔽差异，业务只关心对象 key 与元数据。

## 3. 总体边界

```text
Git Webhook / Manual Trigger
  -> API Server (Gin)
  -> Application Service
  -> Temporal Client
  -> Workflow / Activity
  -> Executor / Object Storage / Integrations
  -> Business DB Projection
```

### 3.1 业务数据库 vs Temporal 持久化

必须严格区分两套存储：

- 业务数据库：
  - 存项目、流水线定义、运行记录、投影状态、制品元数据、审计日志。
  - 服务于 API 查询、后台管理、统计与审计。
- Temporal 持久化：
  - 存 Temporal 自己的 Workflow 执行状态、历史与内部索引。
  - 由 Temporal 官方 schema 管理。

约束：

- 不用 GORM 管理 Temporal 表。
- 不用 Atlas 管理 Temporal 表。
- 业务查询不直接依赖 Temporal 内部表。

### 3.2 读模型策略

Temporal 是编排事实源，业务数据库是查询事实源。
因此需要“工作流状态 -> 业务投影”的同步策略：

- Workflow 启动时创建 `runs` 记录。
- Job / step 进入关键状态时回写 `run_jobs`、`run_steps`。
- 制品、日志对象上传后写入 `artifacts`、`log_objects`。
- 审批、取消、重试等控制动作写入 `audit_logs`。

首版建议采用“同一 Activity 末尾显式回写”的简单模型，而不是事件总线式异步投影。

## 4. 服务分层建议

建议 `apps/server` 至少分成下面几层：

### 4.1 `api`

职责：

- HTTP handler
- 请求校验
- 认证鉴权
- 错误映射
- 返回统一响应结构

不负责：

- 业务规则编排
- 直接写复杂 SQL
- 直接操作对象存储 SDK

### 4.2 `app`

职责：

- 应用服务编排
- API 请求到领域动作的转换
- 调用 repository、Temporal Client、Object Storage Service

典型服务：

- `RunService`
- `ProjectService`
- `DeploymentApprovalService`

### 4.3 `workflow`

职责：

- 定义 Workflow 输入输出
- 定义主流程、阶段推进、审批等待、取消语义
- 只保留确定性逻辑

建议首版工作流：

- `PipelineRunWorkflow`
- `DeploymentApprovalWorkflow` 可先不拆，合并在 `PipelineRunWorkflow`

### 4.4 `activity`

职责：

- 调用 executor 执行步骤
- 上传日志与制品
- 回写业务投影
- 调用外部集成，如通知器、部署器

建议按职责拆 Activity：

- `PrepareRunActivity`
- `ExecuteStepActivity`
- `PersistProjectionActivity`
- `UploadArtifactActivity`
- `NotifyRunResultActivity`

### 4.5 `executor`

职责：

- 执行 `docker` / `packaged` 两类 step
- 收集 stdout / stderr / exit code
- 管理工作目录、缓存目录、环境变量注入

约束：

- 不直接操作 GORM
- 不直接修改业务状态
- 通过返回结构把执行结果交给 Activity 再处理
- Runner 默认作为宿主机原生服务存在，不以 Docker 容器作为默认部署方式

### 4.6 `repository`

职责：

- 封装 GORM 查询
- 按聚合边界组织持久化逻辑
- 控制事务边界

约束：

- 禁止将 Workflow / Activity 语义泄漏进 repository
- 尽量避免在业务层拼接方言差异 SQL

### 4.7 `storage/object`

职责：

- 抽象对象上传、下载、删除、查询与预签名
- 提供 local / s3 两个实现

建议接口：

```go
type ObjectStorage interface {
    Put(ctx context.Context, req PutObjectRequest) (ObjectDescriptor, error)
    Get(ctx context.Context, req GetObjectRequest) (io.ReadCloser, ObjectMetadata, error)
    Delete(ctx context.Context, req DeleteObjectRequest) error
    Stat(ctx context.Context, req StatObjectRequest) (ObjectMetadata, error)
    PresignGet(ctx context.Context, req PresignGetRequest) (string, error)
}
```

## 5. Temporal 设计原则

### 5.1 Workflow 只做编排

Workflow 中只保留：

- 阶段推进
- 依赖判断
- 审批等待
- 重试/超时策略
- 状态分支决策

不要在 Workflow 里做：

- 网络调用
- 数据库 I/O
- 对象存储上传
- docker / packaged 执行

这些都放入 Activity。

### 5.2 Activity 承载副作用

Activity 适合承载：

- 拉取源码
- 执行命令
- 上传日志和制品
- 更新业务投影
- 发送通知

每个 Activity 尽量短、小、可重试，并显式声明 timeout / retry policy。

### 5.3 人工批准

人工批准推荐建模为：

- Workflow 进入 `awaiting_approval`
- API 调用 `approve` 接口
- 服务层向 Workflow 发送 Signal / Update
- Workflow 收到批准后继续

不要用数据库轮询去模拟等待批准。

## 6. GORM 与多数据库兼容策略

### 6.1 基本原则

- 默认开发、测试、生产模板优先 PostgreSQL。
- MySQL 兼容作为正式约束纳入 schema 设计和 CI。
- 只使用两边共有的主路径特性。

### 6.2 建模约束

- 主键默认使用字符串 UUID，避免依赖数据库自增语义。
- 枚举值用 `varchar` 存，不依赖数据库原生 enum。
- JSON 字段作为逻辑 JSON 使用：
  - PostgreSQL 可映射 `jsonb`
  - MySQL 可映射 `json`
- 不依赖数组字段、部分索引、方言特有函数。
- 时间统一使用 UTC。

### 6.3 迁移约束

- 迁移脚本必须由 Atlas 管理并落库到仓库。
- 生产环境迁移通过显式命令执行，不在应用启动时偷偷变更 schema。
- 每次 schema 变更都需要同时验证 PostgreSQL 与 MySQL。

## 7. 对象存储设计

### 7.1 存储对象分类

首版建议至少区分两类对象：

1. 日志对象
2. 制品对象

建议对象 key：

```text
{project_id}/runs/{run_id}/jobs/{job_id}/steps/{step_id}/logs/{seq}.log
{project_id}/runs/{run_id}/jobs/{job_id}/artifacts/{artifact_id}/{filename}
```

### 7.2 Local 后端

建议：

- 使用根目录配置 `OBJECT_STORAGE_LOCAL_ROOT`
- 所有对象 key 映射为相对路径
- API 下载时校验路径逃逸，禁止访问根目录外文件

适合：

- 本地开发
- 单机部署
- 测试环境

### 7.3 S3 后端

建议：

- 使用 bucket + prefix 模式隔离环境
- 支持 AWS S3 和 S3 兼容存储
- 对下载接口预留预签名 URL 能力

适合：

- 多节点部署
- 跨实例共享日志与制品
- 生命周期管理与归档策略

## 8. 配置建议

建议按下面分类组织配置：

- `HTTP_*`：API 服务监听、超时、CORS
- `DB_*`：业务数据库连接
- `TEMPORAL_*`：Temporal namespace、task queue、server address
- `OBJECT_STORAGE_*`：存储后端类型、本地根目录、S3 endpoint/bucket/region
- `EXECUTOR_*`：docker / packaged 执行参数
- `SECRET_*`：主密钥、加密参数

建议通过环境变量驱动，开发环境可配 `.env` 或 `docker compose`。

## 9. 首版落地顺序

### Phase 1

- Temporal 跑通最小 `PipelineRunWorkflow`
- 业务数据库落 `projects / pipeline_defs / pipeline_versions / runs`
- Object Storage 先接 `local`

### Phase 2

- 补 `run_jobs / run_steps / workers / worker_telemetry_snapshots / artifacts / log_objects / audit_logs`
- 补 `docker` 与 `packaged` 执行器
- 补 runner heartbeat / probe 通道
- 补审批信号流转

### Phase 3

- 对象存储扩到 `s3`
- 数据库兼容验证补 MySQL
- API 补下载与预签名能力

## 10. 待确认问题

1. run / job / step 投影是否允许短暂最终一致。
2. 日志读取首版是 API 聚合读取，还是优先支持对象直读。
3. MySQL 兼容是否属于 MVP 必做，还是在 PostgreSQL 跑稳后补齐。
4. 是否需要把通知器和部署器从 Activity 中继续拆成独立适配层。
