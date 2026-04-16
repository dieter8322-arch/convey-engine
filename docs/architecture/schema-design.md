# Schema Design

- 状态：Draft v0.2
- 日期：2026-04-16
- 适用范围：Convey Engine 业务数据库 schema 与迁移约束

## 1. 范围

本文只描述业务数据库 schema，不描述 Temporal 内部持久化 schema。

约束：

- 业务 schema 使用 GORM 建模，但正式迁移由 Atlas 管理。
- Temporal 内部表由 Temporal 官方 schema 管理，不纳入本文件。
- 业务数据库默认 PostgreSQL，同时保持 MySQL 兼容。

## 2. 数据库边界

### 2.1 业务数据库

职责：

- 项目与流水线配置
- 运行记录与投影状态
- 制品与日志对象元数据
- 审计日志
- 密钥引用

建议逻辑库名：

- PostgreSQL：`convey_app`
- MySQL：`convey_app`

### 2.2 Temporal 持久化数据库

职责：

- Temporal Workflow Execution
- 历史事件
- 可见性索引

建议逻辑库名：

- PostgreSQL：`temporal`
- MySQL：`temporal`

约束：

- 业务应用不直接访问该库
- Atlas 不管理该库
- GORM 不映射该库

## 3. 通用建模约束

### 3.1 主键与时间

- 所有主表主键统一使用字符串 UUID
- 逻辑类型记为 `uuid`
- 物理层建议：
  - PostgreSQL：优先 `uuid`
  - MySQL：优先 `char(36)`
- 所有时间字段统一使用 UTC
- 每张高频表至少包含：
  - `created_at`
  - `updated_at`

### 3.2 状态字段

- 状态字段统一使用 `varchar`
- 不依赖数据库原生 enum
- 状态迁移由应用层和 Workflow 保证，不由数据库触发器驱动

### 3.3 JSON 字段

- 逻辑上允许配置与扩展字段使用 JSON
- 物理层建议：
  - PostgreSQL：`jsonb`
  - MySQL：`json`
- 查询路径尽量落在标量列，不依赖复杂 JSON 方言查询

## 4. 核心表设计

### 4.1 `projects`

用途：项目定义

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `name` | varchar(128) | 项目名，唯一 |
| `repo_url` | varchar(512) | 仓库地址 |
| `provider` | varchar(32) | `github` / `gitlab` 等 |
| `default_branch` | varchar(128) | 默认分支 |
| `status` | varchar(32) | `active` / `archived` |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`name`)
- index(`provider`)

### 4.2 `pipeline_defs`

用途：流水线逻辑定义

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `project_id` | uuid | 所属项目 |
| `name` | varchar(128) | 流水线名称 |
| `status` | varchar(32) | `active` / `disabled` |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`project_id`, `name`)
- index(`project_id`, `status`)

### 4.3 `pipeline_versions`

用途：流水线配置版本

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `pipeline_def_id` | uuid | 所属流水线 |
| `version` | integer | 递增版本 |
| `config_raw` | text | 原始 `convey.yaml` |
| `config_hash` | varchar(128) | 配置 hash |
| `parsed_summary_json` | json | 解析摘要 |
| `created_at` | timestamp | 创建时间 |

索引建议：

- unique(`pipeline_def_id`, `version`)
- index(`config_hash`)

### 4.4 `triggers`

用途：触发器定义

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `project_id` | uuid | 所属项目 |
| `pipeline_def_id` | uuid | 目标流水线 |
| `type` | varchar(32) | `push` / `pull_request` / `manual` |
| `filter_json` | json | 分支、路径等规则 |
| `enabled` | boolean | 是否启用 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- index(`project_id`, `enabled`)
- index(`pipeline_def_id`, `type`)

### 4.5 `runs`

用途：一次流水线运行

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `project_id` | uuid | 所属项目 |
| `pipeline_version_id` | uuid | 对应配置版本 |
| `status` | varchar(32) | `pending` / `running` / `awaiting_approval` / `succeeded` / `failed` / `canceled` |
| `ref` | varchar(255) | 分支或 tag |
| `commit_sha` | varchar(64) | 提交哈希 |
| `trigger_type` | varchar(32) | webhook / manual / retry |
| `triggered_by` | varchar(128) | 触发者 |
| `temporal_workflow_id` | varchar(255) | Temporal workflow id |
| `temporal_run_id` | varchar(255) | Temporal run id |
| `started_at` | timestamp | 开始时间 |
| `finished_at` | timestamp | 结束时间 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`temporal_workflow_id`)
- index(`project_id`, `status`, `created_at`)
- index(`pipeline_version_id`, `created_at`)

### 4.6 `run_jobs`

用途：job 级读模型投影

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `run_id` | uuid | 所属运行 |
| `worker_id` | uuid | 实际执行该 job 的 runner |
| `name` | varchar(128) | job 名称 |
| `stage` | varchar(64) | 所属阶段 |
| `status` | varchar(32) | `pending` / `running` / `succeeded` / `failed` / `skipped` / `canceled` |
| `executor_kind` | varchar(32) | `docker` / `packaged` |
| `started_at` | timestamp | 开始时间 |
| `finished_at` | timestamp | 结束时间 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`run_id`, `name`)
- index(`run_id`, `stage`, `status`)
- index(`worker_id`, `status`, `updated_at`)

### 4.7 `run_steps`

用途：step 级读模型投影

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `run_job_id` | uuid | 所属 job |
| `seq` | integer | 步骤序号 |
| `kind` | varchar(32) | `run` / `uses` / `upload` |
| `name` | varchar(128) | 展示名称 |
| `status` | varchar(32) | `pending` / `running` / `succeeded` / `failed` / `canceled` |
| `exit_code` | integer | 退出码 |
| `started_at` | timestamp | 开始时间 |
| `finished_at` | timestamp | 结束时间 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`run_job_id`, `seq`)
- index(`run_job_id`, `status`)

### 4.8 `artifacts`

用途：制品对象元数据

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `run_id` | uuid | 所属运行 |
| `run_job_id` | uuid | 所属 job |
| `kind` | varchar(64) | `archive` / `image-metadata` / `report` 等 |
| `storage_backend` | varchar(16) | `local` / `s3` |
| `bucket` | varchar(255) | S3 bucket 或 local logical bucket |
| `object_key` | varchar(1024) | 对象 key |
| `filename` | varchar(255) | 原始文件名 |
| `content_type` | varchar(255) | MIME 类型 |
| `size_bytes` | bigint | 文件大小 |
| `checksum` | varchar(128) | 内容摘要 |
| `created_at` | timestamp | 创建时间 |

索引建议：

- index(`run_id`, `kind`)
- index(`run_job_id`)
- unique(`storage_backend`, `bucket`, `object_key`)

### 4.9 `log_objects`

用途：日志对象元数据

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `run_id` | uuid | 所属运行 |
| `run_job_id` | uuid | 所属 job |
| `run_step_id` | uuid | 所属 step |
| `storage_backend` | varchar(16) | `local` / `s3` |
| `bucket` | varchar(255) | bucket 或 logical bucket |
| `object_key` | varchar(1024) | 对象 key |
| `seq` | integer | 分段序号 |
| `size_bytes` | bigint | 分段大小 |
| `created_at` | timestamp | 创建时间 |

索引建议：

- unique(`run_step_id`, `seq`)
- index(`run_id`)
- unique(`storage_backend`, `bucket`, `object_key`)

### 4.10 `deployments`

用途：部署记录

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `run_id` | uuid | 所属运行 |
| `environment` | varchar(64) | 目标环境 |
| `status` | varchar(32) | `pending` / `awaiting_approval` / `running` / `succeeded` / `failed` / `canceled` |
| `revision` | varchar(128) | 部署版本 |
| `approver` | varchar(128) | 批准人 |
| `approved_at` | timestamp | 批准时间 |
| `started_at` | timestamp | 开始时间 |
| `finished_at` | timestamp | 结束时间 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- index(`run_id`)
- index(`environment`, `status`, `created_at`)

### 4.11 `secret_refs`

用途：密钥引用

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `project_id` | uuid | 所属项目 |
| `scope` | varchar(32) | `project` / `environment` / `global` |
| `name` | varchar(128) | 密钥名 |
| `ciphertext` | text | 加密密文 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`project_id`, `scope`, `name`)

### 4.12 `audit_logs`

用途：审计日志

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `actor` | varchar(128) | 操作者 |
| `action` | varchar(64) | 动作 |
| `target_type` | varchar(64) | 目标类型 |
| `target_id` | uuid | 目标标识 |
| `payload_json` | json | 补充上下文 |
| `created_at` | timestamp | 创建时间 |

索引建议：

- index(`target_type`, `target_id`, `created_at`)
- index(`actor`, `created_at`)

### 4.13 `workers`

用途：执行节点与控制节点的注册信息

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `name` | varchar(128) | worker 名称 |
| `mode` | varchar(32) | `control` / `runner` |
| `status` | varchar(32) | `online` / `busy` / `offline` / `draining` |
| `platform` | varchar(32) | `linux` / `macos` / `windows` |
| `arch` | varchar(32) | `amd64` / `arm64` 等 |
| `hostname` | varchar(255) | 宿主机名 |
| `version` | varchar(64) | worker 版本 |
| `token_hash` | varchar(255) | 注册令牌摘要 |
| `labels_json` | json | 标签与资源池信息 |
| `capabilities_json` | json | docker、powershell、shell 等能力 |
| `connected_at` | timestamp | 当前连接建立时间 |
| `last_seen_at` | timestamp | 最后活跃时间 |
| `last_heartbeat_at` | timestamp | 最后心跳时间 |
| `last_probe_at` | timestamp | 最后探针时间 |
| `current_run_id` | uuid | 当前执行中的 run |
| `current_job_id` | uuid | 当前执行中的 job |
| `current_step_id` | uuid | 当前执行中的 step |
| `status_reason` | varchar(255) | 当前状态说明 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |

索引建议：

- unique(`name`)
- index(`mode`, `status`, `last_seen_at`)
- index(`platform`, `arch`)
- index(`last_heartbeat_at`)
- index(`last_probe_at`)

### 4.14 `worker_telemetry_snapshots`

用途：runner 上报的宿主机状态快照

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uuid | 主键 |
| `worker_id` | uuid | 所属 worker |
| `cpu_usage_pct` | decimal(5,2) | CPU 使用率 |
| `memory_total_bytes` | bigint | 内存总量 |
| `memory_used_bytes` | bigint | 已用内存 |
| `memory_used_pct` | decimal(5,2) | 内存使用率 |
| `disk_total_bytes` | bigint | 工作盘总量 |
| `disk_free_bytes` | bigint | 工作盘剩余量 |
| `disk_used_pct` | decimal(5,2) | 工作盘使用率 |
| `docker_available` | boolean | Docker 是否可用 |
| `docker_version` | varchar(128) | Docker 版本 |
| `network_ok` | boolean | 控制面连通性摘要 |
| `error_summary` | varchar(255) | 最近探针异常摘要 |
| `collected_at` | timestamp | 快照采集时间 |
| `created_at` | timestamp | 入库时间 |

索引建议：

- index(`worker_id`, `collected_at`)
- index(`collected_at`)

## 5. 关系约束建议

- `projects 1:n pipeline_defs`
- `pipeline_defs 1:n pipeline_versions`
- `workers 1:n run_jobs`
- `workers 1:n worker_telemetry_snapshots`
- `projects 1:n runs`
- `runs 1:n run_jobs`
- `run_jobs 1:n run_steps`
- `runs 1:n artifacts`
- `runs 1:n log_objects`
- `runs 1:n deployments`

外键策略：

- 首版建议保留数据库外键，优先保证数据完整性。
- 删除策略尽量保守，默认禁止级联删除运行数据。

## 6. 状态模型建议

### 6.1 `runs.status`

```text
pending -> running -> awaiting_approval -> running -> succeeded
pending -> running -> failed
pending -> running -> canceled
```

说明：

- `awaiting_approval` 为可选中间态。
- 重试时建议创建新 run，而不是原 run 回滚到早期状态。

### 6.2 `run_jobs.status`

```text
pending -> running -> succeeded
pending -> running -> failed
pending -> skipped
pending -> canceled
```

### 6.3 `workers.status`

```text
online -> busy -> online
online -> draining -> offline
online -> offline
busy -> offline
```

说明：

- `last_seen_at` 超过心跳超时阈值后可判定为 `offline`
- `draining` 表示不再接新任务，但允许当前任务自然完成
- `last_probe_at` 过旧不一定等于离线，但控制面板应标记 telemetry 过期

### 6.4 `deployments.status`

```text
pending -> awaiting_approval -> running -> succeeded
pending -> awaiting_approval -> canceled
pending -> running -> failed
```

## 7. 迁移策略

### 7.1 业务 schema

- 业务 schema 由 Atlas 管理版本化迁移文件。
- 每次 schema 变更都需要：
  - 更新模型定义
  - 生成 Atlas migration
  - 在 PostgreSQL 与 MySQL 上验证

### 7.2 Temporal schema

- Temporal schema 由 Temporal 官方工具和部署清单维护。
- 不把 Temporal migration 混入业务迁移目录。

### 7.3 本地开发

建议：

- `docker compose` 启动 PostgreSQL / MySQL、Temporal、对象存储依赖
- 业务库初始化由 Atlas 执行
- Temporal 库初始化由官方镜像 / 脚本执行

## 8. PostgreSQL / MySQL 兼容约束

为了保持双数据库兼容，schema 设计应遵守下面规则：

- 不依赖 PostgreSQL 数组类型
- 不依赖 PostgreSQL 部分索引
- 不依赖 MySQL 专属全文索引行为
- JSON 只做扩展载荷，不做高复杂度查询主路径
- 索引尽量落在：
  - `status`
  - `created_at`
  - 外键列
  - 唯一业务键

## 9. 对象元数据与清理策略

### 9.1 元数据写入

- 对象上传成功后再写 `artifacts` / `log_objects`
- 写元数据失败时，需要补偿删除孤儿对象，或记录待清理事件

### 9.2 清理策略

- 删除 run 默认不自动物理删除对象
- 通过后台清理任务按保留策略清理对象
- Local 与 S3 共用同一套“元数据驱动清理”逻辑

## 10. 待补细节

1. 是否需要 `environments` 独立表承载环境配置。
2. 是否需要 `run_variables` 表持久化运行时变量快照。
3. 是否需要 `notifications` 表承载通知订阅配置。
4. 是否需要将日志索引从 `log_objects` 继续拆成 chunk 级别更细的表。
5. 是否需要把 `workers` 的在线会话继续拆成独立 `worker_sessions` 表。
6. 是否需要对 `worker_telemetry_snapshots` 做按天分区或保留策略。
