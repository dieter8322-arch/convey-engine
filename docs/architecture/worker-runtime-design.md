# Worker Runtime Design

- 状态：Draft v0.3
- 日期：2026-04-16
- 适用范围：Convey Engine `worker` / `runner agent` 进程、实时控制通道与跨平台执行运行时设计

## 1. 目标

本文用于细化 Worker 侧设计，重点回答下面几个问题：

1. `worker` 是否应该与 API 保持长连接。
2. 如何同时兼顾 Temporal 编排和“每台机器一个 worker”的执行模型。
3. Linux / macOS / Windows / Docker 容器上的运行方式如何统一。
4. Windows 等桌面节点如何尽量实现“打包后即可运行”和执行环境隔离。

本文是 [system-overview.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/system-overview.md) 与 [backend-stack-design.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/backend-stack-design.md) 的补充，聚焦“Worker 怎么接入、怎么领任务、怎么隔离执行”，不重复解释整体 API、数据库与前端边界。

## 2. 结论摘要

首版建议采用下面方案：

1. `worker` 改按“双角色”理解：
   - `control worker`：部署在服务端，消费 Temporal Workflow / Activity。
   - `runner worker`：部署在执行节点，一般一台机器一个实例。
2. `runner worker` 必须与控制面保持长期在线连接，用于注册、心跳、实时接收任务、流式回传日志与接收取消信号。
3. `control worker` 负责把 Temporal 内的执行 Activity 转换成对在线 runner 的任务派发。
4. `runner worker` 默认应原生常驻在宿主机，不把 Docker 作为边缘节点 runner 的默认部署方式。
5. `runner worker` 内置 probe，持续把宿主机状态上报给控制面板。
6. `runner worker` 需要跨平台提供统一二进制分发，至少支持 Linux、macOS、Windows 原生运行，也可在少数服务端场景运行于 Docker 容器中。
7. 执行隔离采用“双执行器”：
   - `docker` executor：优先用于 Linux 和安装了 Docker Desktop / WSL2 的节点。
   - `packaged` executor：用于原生主机，尤其是 Windows，依赖打包 runtime、临时工作目录、受限进程与清理策略实现最小隔离。

这样做的原因：

- 你提出的“worker 和 API 实时通信”本质上是 runner 与 control plane 的实时双向通道需求，而不是让所有边缘节点直接承担 Workflow 编排。
- 一机一 worker 的场景下，不应该让每个边缘节点都直接消费 Workflow Task；否则执行节点和编排节点会耦合得太深。
- 跨平台场景里，Docker 不能覆盖所有机器；Windows 节点必须有非容器路径。
- 如果 runner 还承担宿主机探针职责，把 runner 本身放进 Docker 通常会让宿主机观测、权限控制和平台兼容都更复杂。

## 3. 角色拆分

### 3.1 `control worker`

职责：

- 连接 Temporal Server，消费 Workflow / Activity。
- 选择合适的在线 runner。
- 将 step 下发给 runner，并等待执行结果。
- 汇总日志、制品元数据和执行结果后回写业务数据库。

特点：

- 部署在服务端环境。
- 与 Temporal 的连接是内部基础设施连接。
- 不直接执行用户 step。

### 3.2 `runner worker`

职责：

- 常驻在执行节点。
- 启动后向控制面注册自身信息。
- 保持长连接，持续发送 heartbeat。
- 接收 step，执行后实时回传日志、状态和结果。
- 在本机或容器中管理工作目录、环境变量与清理动作。

特点：

- 面向“每台机器一个 worker”。
- 支持 Linux / macOS / Windows / Docker 容器。
- 默认不直接连接 Temporal。

## 4. 连接模型

### 4.1 为什么需要与 API 长连接

你的判断是对的。
如果 runner 部署在用户机器、办公室电脑或 Windows 节点上，控制面必须能实时知道：

- 哪些 runner 在线
- 该 runner 支持哪些执行能力
- 当前是否空闲 / 正忙 / 离线
- 任务执行中的日志和进度
- 取消、超时、中断是否已被 runner 接收

因此，`runner worker` 与控制面应保持长期在线连接。

### 4.2 推荐连接方式

推荐增加专用 `Runner Gateway`，由 API / Control Plane 暴露 gRPC streaming 长连接：

- runner 发起出站连接
- gateway 完成鉴权、注册、心跳维护
- gateway 下发任务、取消指令和配置更新
- runner 流式上报日志、状态、能力信息

为什么不是 API 主动连 runner：

- 大多数 runner 位于 NAT 或办公网络后面
- 出站连接更容易部署
- 安全边界更清晰

### 4.3 Temporal 与长连接的分工

分工建议如下：

- Temporal：
  - 管 Workflow 生命周期
  - 管重试、超时、审批等待、恢复语义
- Runner Gateway 长连接：
  - 管 runner 在线状态
  - 管节点注册与能力上报
  - 管任务下发与日志流
  - 管取消信号快速扇出

结论：

- 需要长连接，但连接重点应是 `runner -> control plane`
- 不建议让所有边缘 runner 直接承担 Temporal Worker 角色

## 5. 进程与部署模型

### 5.1 进程入口建议

建议目录：

```text
apps/server/
  cmd/
    convey-api/
    convey-worker/
```

建议 `convey-worker` 支持两种 mode：

- `control`
- `runner`

这样可以复用一套代码与配置体系，但部署职责分开。

### 5.2 控制面部署

服务端建议至少包含：

- `convey-api`
- `convey-worker --mode=control`
- Temporal Server
- Business DB
- Object Storage

### 5.3 执行节点部署

每个执行节点建议部署：

- `convey-worker --mode=runner`

典型场景：

- 一台 Linux 构建机一个 runner
- 一台 macOS 签名机一个 runner
- 一台 Windows 发布机一个 runner

这比把所有执行任务都塞进统一服务端 Worker 更贴近真实 CI runner 形态。

建议运行形态：

- Linux：`systemd` service
- macOS：`launchd` agent / daemon
- Windows：Windows Service

### 5.4 为什么 runner 默认不放 Docker

不建议把边缘 runner 默认放进 Docker，原因如下：

- Probe 需要观测宿主机 CPU、内存、磁盘、Docker 可用性与网络状态，而不是容器视角状态。
- 若 runner 容器要准确观测宿主机，通常要挂 host 文件系统、提权或绑定 `docker.sock`，隔离收益会明显下降。
- Windows / macOS 原生节点不一定稳定具备 Docker 运行前提，但原生 runner 仍必须存在。
- 长连接、开机自启、系统代理、证书、桌面签名工具等宿主机集成，原生服务通常更直接。

保留例外：

- 开发环境快速体验
- 服务端一体化 demo
- 明确不需要宿主机探针的纯容器执行节点

## 6. runner 注册、探针与能力模型

runner 首次上线时应上报：

- `worker_id`
- `hostname`
- `platform`
- `arch`
- `version`
- `mode=runner`
- `executor` 能力
- `docker` 可用性
- `labels`
- 当前状态
- 默认执行器
- 是否启用 probe

其中能力模型至少覆盖：

- `platform=linux|macos|windows`
- `executor=docker|packaged`
- `docker_available=true|false`
- `shell_available=true|false`
- `powershell_available=true|false`

调度时，控制面根据 job 约束选择 runner，例如：

- 只在 Windows 上执行
- 必须有 Docker
- 只允许指定标签的 runner 执行部署

### 6.1 Heartbeat

Heartbeat 负责表达“在线与否”和“是否空闲”，建议高频上报。

建议字段：

- `worker_id`
- `status`
- `current_run_id`
- `current_job_id`
- `current_step_id`
- `last_activity_at`
- `last_heartbeat_at`

建议频率：

- 5 到 10 秒一次

用途：

- 控制面判断 runner 是否在线
- 调度器判断 runner 是否可接新任务
- 控制面板实时展示忙闲状态

### 6.2 Probe / Telemetry

Probe 负责表达“机器现在健康吗、还能跑什么”，建议低频上报。

建议内容：

- CPU 使用率
- 内存总量 / 已用量 / 使用率
- 磁盘总量 / 可用量 / 使用率
- 系统 load 或进程压力摘要
- Docker 可用性与版本
- PowerShell / Shell 可用性
- 网络连通性摘要
- Runner 自身版本与错误摘要

建议频率：

- 15 到 60 秒一次

约束：

- Probe 失败不应立即导致任务失败
- 心跳与 telemetry 应拆分，避免高频上报大 payload
- 控制面板展示优先读最近一次快照，不从日志反推

### 6.3 控制面板建议字段

控制面板至少展示下面字段：

- `worker_name`
- `mode`
- `platform`
- `arch`
- `status`
- `labels`
- `current_run`
- `current_job`
- `current_step`
- `last_heartbeat_at`
- `last_probe_at`
- `cpu_usage_pct`
- `memory_used_pct`
- `disk_used_pct`
- `docker_available`
- `runner_version`

字段来源建议：

- 当前态：`workers`
- 历史趋势与最近探针：`worker_telemetry_snapshots`

## 7. 执行生命周期

推荐执行链路如下：

1. API 创建 `run` 并启动 Temporal Workflow。
2. `control worker` 消费到某个 `ExecuteStepActivity`。
3. `control worker` 从在线 runner 列表中选择满足约束的节点。
4. 目标 runner 通过长连接收到 step。
5. runner 在本机选择 `docker` 或 `packaged` executor 执行。
6. runner 持续回传 stdout / stderr、阶段状态与 heartbeat。
7. runner 结束后回传 exit code、制品清单和摘要元数据。
8. `control worker` 负责最终写入 `run_jobs`、`run_steps`、`artifacts`、`log_objects`、`audit_logs`。
9. Workflow 根据 Activity 结果继续推进或失败恢复。

约束：

- Workflow 不直接做 I/O。
- Runner 不直接决定调度。
- 边缘 runner 默认不直接写业务数据库。

## 8. 执行器与隔离策略

### 8.1 `docker` executor

适用场景：

- Linux 主机构建机
- 装有 Docker Desktop / WSL2 的 Windows 节点
- 需要强隔离的构建和测试任务

要求：

- 镜像、挂载目录、网络权限、用户身份可配置
- 默认只挂最小必要卷
- 退出后清理容器与临时目录

### 8.2 `packaged` executor

这是跨平台 runner 的保底路径，尤其用于 Windows。

设计目标：

- 安装后即可运行，不强依赖用户机器预装一堆工具
- 用打包 runtime 降低“环境不一致”风险
- 用临时工作目录、受限进程、超时和清理动作实现最小隔离

建议方式：

- Worker 分发为自包含安装包或压缩包
- 内置最小运行时与 helper 工具
- 每次任务创建独立 workspace
- 结束后按策略清理 workspace、缓存和敏感环境变量

### 8.3 平台差异建议

Linux：

- 优先 `docker`
- 无 Docker 时可回退到 `packaged`
- runner 本身优先原生 service，不建议容器化

macOS：

- 优先原生 `packaged`
- 如果装了 Docker Desktop，可允许 `docker`
- 适合承担签名、打包等必须在 macOS 执行的任务
- runner 本身优先 `launchd`

Windows：

- 必须支持原生 `packaged`
- 若检测到 Docker Desktop / WSL2，可额外支持 `docker`
- PowerShell / CMD / 受限进程对象应纳入 executor 设计
- runner 本身优先 Windows Service

## 9. 日志、制品与实时反馈

Runner 长连接除了接任务，还应承担实时反馈职责：

- 日志流实时回传
- step 状态切换实时回传
- 心跳持续上报
- 取消指令快速送达

制品上传建议：

- 由控制面签发一次性上传授权
- runner 直接上传对象存储，或经控制面代理上传
- 最终由 `control worker` 回写对象元数据

这样可以兼顾：

- 实时性
- 跨平台一致性
- 控制面对元数据和审计链路的掌控

## 10. 幂等性与状态回写

所有副作用都应围绕“可重试、可追踪”设计。

### 10.1 回写原则

- `runs` 记录由 API / 应用服务在启动 Workflow 时先创建
- `control worker` 负责最终业务投影回写
- runner 的长连接消息只作为执行事实输入，不直接成为业务事实源

### 10.2 幂等建议

- 对象 key 使用稳定命名：

```text
{project_id}/runs/{run_id}/jobs/{job_id}/steps/{step_id}/logs/{seq}.log
{project_id}/runs/{run_id}/jobs/{job_id}/artifacts/{artifact_id}/{filename}
```

- `run_jobs` 应记录 `worker_id`
- `PersistProjectionActivity` 应使用稳定主键或唯一键更新投影，避免重试重复写入
- Runner reconnect 后应能根据 `run_id/job_id/step_id` 恢复本地执行上下文或明确标记失败

## 11. 配置建议

除通用后端配置外，建议增加下面参数。

控制面：

- `RUNNER_GATEWAY_ADDR`
- `RUNNER_AUTH_ISSUER`
- `RUNNER_HEARTBEAT_TIMEOUT`
- `RUNNER_DISPATCH_TIMEOUT`
- `RUNNER_PROBE_STALE_TIMEOUT`

Runner：

- `WORKER_MODE=runner`
- `WORKER_ID`
- `WORKER_NAME`
- `WORKER_LABELS`
- `WORKER_WORKSPACE_ROOT`
- `WORKER_LOG_CHUNK_BYTES`
- `WORKER_ARTIFACT_TMP_ROOT`
- `WORKER_CLEANUP_POLICY`
- `WORKER_DOCKER_ENABLED`
- `WORKER_PACKAGED_EXECUTOR_ENABLED`
- `WORKER_HEARTBEAT_INTERVAL`
- `WORKER_TELEMETRY_INTERVAL`
- `WORKER_PROBE_ENABLED`
- `WORKER_NATIVE_SERVICE_MODE`

## 12. 可观测性与运维

至少应输出下面观测信息：

- 结构化日志字段：`worker_id`、`worker_mode`、`platform`、`run_id`、`job_id`、`step_id`
- 指标：在线 runner 数、空闲 runner 数、分发耗时、心跳超时数、probe 过期数、执行超时数、日志上传失败数
- 健康信号：gateway 连通性、runner 最后心跳、runner 最近 probe、workspace 可写性、Docker 可用性

重点排障链路：

1. Workflow 是否已经推进到需要派发 runner 的节点
2. 目标 runner 是否在线且能力匹配
3. 长连接是否存活
4. Runner 是否真正启动了本地 executor
5. 最终结果是否已被 `control worker` 回写

## 13. MVP 边界与延后项

首版纳入：

1. `control worker` 与 `runner worker` 双角色模型
2. `runner -> control plane` 长连接
3. Runner 自注册、心跳与最小能力模型
4. 内置 probe 与控制面板当前态展示
5. Linux / macOS / Windows 原生 runner 支持
6. `docker` 与 `packaged` 两类执行器
7. `run_jobs.worker_id` 级别的执行节点追踪

明确延后：

1. 多租户资源池与复杂调度权重
2. 自动弹性扩容
3. 远端 VM / Kubernetes / Nomad 执行器
4. 更强的 OS 级沙箱与快照回滚
5. Runner 灰度升级与自动更新平台

## 14. 与现有文档的关系

- 系统级角色分工与整体链路：见 [system-overview.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/system-overview.md)
- 后端分层、Temporal / GORM / Atlas 边界：见 [backend-stack-design.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/backend-stack-design.md)
- 业务数据库表结构与索引：见 [schema-design.md](/Users/yandi/Workspace/dieter/convey-engine/docs/architecture/schema-design.md)
