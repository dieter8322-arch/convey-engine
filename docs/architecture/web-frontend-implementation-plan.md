# Pipeline Editor Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付一套“图形化优先、YAML 后置”的流水线前端编辑器，支持阶段泳道、任务节点、依赖关系、右侧属性编辑、YAML 专家模式和保存前校验。

**Architecture:** 前端以 `PipelineDraft` 作为唯一真实数据源，`X6 Graph` 只负责画布渲染与交互，`YAML` 作为 `PipelineDraft` 的序列化视图与专家模式入口。页面层采用 `Naive UI` 承担导航、表单、消息和抽屉，状态层用 `Pinia` 管编辑草稿与画布状态，用 `@tanstack/vue-query` 管远端查询与保存。

**Tech Stack:** Vue 3, TypeScript, Vite, Vue Router, Pinia, @tanstack/vue-query, Naive UI, AntV X6, Monaco Editor, AJV, YAML

---

## 1. 范围与约束

- 首版聚焦单个“新建/编辑流水线”页面，不扩展到运行记录、日志查看、全局后台框架。
- 用户主路径是图形化编排，YAML 仅作为切换视图、导入导出和高级编辑入口。
- 真实执行语义只放在 `PipelineDraft`，坐标、折叠、视口等只放在 `ui` 布局信息里。
- 当前计划遵循用户偏好：只落计划文档，不生成测试脚本，不编译、不运行。
- 文档中的“验证”统一表达为验收点与人工检查项，不附带执行命令。

## 2. 建议文件结构

### 2.1 工作区与应用层

- `package.json`：根工作区脚本与统一开发依赖入口。
- `pnpm-workspace.yaml`：声明 `apps/*`、`packages/*` 工作区。
- `tsconfig.base.json`：前端共享 TypeScript 基础配置与路径别名。
- `apps/web/package.json`：前端应用依赖与脚本。
- `apps/web/index.html`：Vite 入口 HTML。
- `apps/web/vite.config.ts`：Vite 别名、环境变量和构建配置。
- `apps/web/tsconfig.json`：前端应用 TS 配置。
- `apps/web/src/main.ts`：挂载 Vue、Router、Pinia、QueryClient。
- `apps/web/src/App.vue`：应用根组件。

### 2.2 应用骨架

- `apps/web/src/router/index.ts`：前端路由定义。
- `apps/web/src/app/providers/ui-provider.vue`：`NConfigProvider`、`NMessageProvider`、`NDialogProvider`、`NNotificationProvider` 根封装。
- `apps/web/src/app/providers/query-provider.ts`：`vue-query` 实例与默认策略。
- `apps/web/src/app/theme/theme-overrides.ts`：Naive UI 全局主题与 token。
- `apps/web/src/app/layouts/console-layout.vue`：控制台壳层布局。

### 2.3 编排页

- `apps/web/src/views/pipeline-editor/index.vue`：流水线编排页入口。
- `apps/web/src/views/pipeline-editor/components/editor-toolbar.vue`：顶部标题、模式切换、保存/执行按钮。
- `apps/web/src/views/pipeline-editor/components/editor-tabs.vue`：`基本信息 / 任务编排 / 参数设置 / 触发设置 / 权限管理 / 通知订阅` 页签。
- `apps/web/src/views/pipeline-editor/components/pipeline-canvas.vue`：X6 容器与画布事件桥接。
- `apps/web/src/views/pipeline-editor/components/inspector-panel.vue`：右侧属性面板容器。
- `apps/web/src/views/pipeline-editor/components/yaml-workspace.vue`：Monaco 编辑区与 YAML 诊断展示。
- `apps/web/src/views/pipeline-editor/components/stage-column-node.ts`：阶段列节点定义。
- `apps/web/src/views/pipeline-editor/components/job-card-node.ts`：任务节点定义。

### 2.4 状态与领域模型

- `apps/web/src/stores/pipeline-editor.ts`：编排页草稿、选中态、历史状态。
- `apps/web/src/stores/pipeline-ui.ts`：视口、面板开合、模式切换等纯 UI 状态。
- `apps/web/src/features/pipeline-draft/types.ts`：前端侧 `PipelineDraft` 类型补充。
- `apps/web/src/features/pipeline-draft/commands.ts`：新增阶段、移动节点、删除节点、连线等命令。
- `apps/web/src/features/pipeline-draft/graph-adapter.ts`：`PipelineDraft <-> X6 Graph` 映射。
- `apps/web/src/features/pipeline-draft/serializer.ts`：`PipelineDraft <-> YAML` 转换。
- `apps/web/src/features/pipeline-draft/validation.ts`：AJV 校验与错误归一化。

### 2.5 共享契约

- `packages/shared/package.json`：共享包定义。
- `packages/shared/src/pipeline/schema.ts`：JSON Schema 导出。
- `packages/shared/src/pipeline/types.ts`：前后端共享流水线类型。
- `packages/shared/src/pipeline/defaults.ts`：默认阶段、默认任务、空草稿工厂。
- `packages/shared/src/index.ts`：共享导出入口。

### 2.6 服务访问层

- `apps/web/src/api/http.ts`：HTTP 客户端封装。
- `apps/web/src/api/pipeline.ts`：查询、保存、执行预检 API。
- `apps/web/src/api/types.ts`：接口返回类型。

## 3. 分阶段实施计划

### Task 1: 初始化 Web 工作区与共享包

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `tsconfig.base.json`
- Create: `apps/web/package.json`
- Create: `apps/web/index.html`
- Create: `apps/web/vite.config.ts`
- Create: `apps/web/tsconfig.json`
- Create: `apps/web/src/main.ts`
- Create: `apps/web/src/App.vue`
- Create: `packages/shared/package.json`
- Create: `packages/shared/src/index.ts`

- [ ] 明确工作区使用 `pnpm`，根目录只保留工作区脚本、lint/typecheck 占位脚本和共享依赖版本，不在根目录堆业务代码。
- [ ] 初始化 `apps/web` 为独立 Vite + Vue 3 + TypeScript 应用，预留 `@` 指向 `src`，预留 `@convey/shared` 指向共享包。
- [ ] 初始化 `packages/shared`，先只承载流水线相关类型与 schema，不引入任何运行时框架依赖。
- [ ] 在 `main.ts` 完成 `createApp(App)`、`createPinia()`、`router`、`Vue Query` 挂载顺序约定。
- [ ] 在 `App.vue` 中只保留 Provider 壳层与 `router-view`，避免后续把业务逻辑塞进根组件。
- [ ] 验收点：工作区结构可支持后续 `apps/web` 与 `packages/shared` 同时演进，前端入口与共享包入口边界清晰。

### Task 2: 搭应用骨架与 Naive UI Provider 体系

**Files:**
- Create: `apps/web/src/router/index.ts`
- Create: `apps/web/src/app/providers/ui-provider.vue`
- Create: `apps/web/src/app/providers/query-provider.ts`
- Create: `apps/web/src/app/theme/theme-overrides.ts`
- Create: `apps/web/src/app/layouts/console-layout.vue`
- Modify: `apps/web/src/App.vue`

- [ ] 在 `ui-provider.vue` 中统一包裹 `NConfigProvider`、`NMessageProvider`、`NDialogProvider`、`NNotificationProvider`，并接入 `zhCN`、`dateZhCN` 和全局主题覆盖。
- [ ] 将控制台视觉 token 收敛到 `theme-overrides.ts`，优先定义品牌色、边框色、hover 色、圆角、阴影和字体层级，不把主题散落到页面内联样式。
- [ ] 设计 `console-layout.vue`：顶部导航区、主体内容区、右侧扩展槽位，为编排页提供稳定壳层。
- [ ] 路由首版只保留 `/pipelines/new` 和 `/pipelines/:id/edit` 两个入口，避免过早扩张页面范围。
- [ ] 在 `App.vue` 中将 Provider 壳层与布局壳层解耦，后续其他页面可以复用相同 Provider 栈。
- [ ] 验收点：Naive UI 全局消息、对话框、通知、主题覆盖和中文 locale 的接入路径明确，页面框架可承接编排页。

### Task 3: 定义 `PipelineDraft` 与共享 Schema

**Files:**
- Create: `packages/shared/src/pipeline/types.ts`
- Create: `packages/shared/src/pipeline/schema.ts`
- Create: `packages/shared/src/pipeline/defaults.ts`
- Modify: `packages/shared/src/index.ts`
- Create: `apps/web/src/features/pipeline-draft/types.ts`

- [ ] 在共享包中定义最小可用领域对象：`PipelineDraft`、`StageDraft`、`JobDraft`、`EdgeDraft`、`TriggerDraft`、`VariableDraft`、`NotificationDraft`、`UiLayoutDraft`。
- [ ] 明确 `ui` 只承载坐标、宽度、折叠、视口、选中等展示态，绝不承载 `needs`、`stage`、`environment` 等执行语义。
- [ ] 在 `schema.ts` 中输出与后端一致的 JSON Schema 草案，首版先覆盖阶段、任务、依赖、触发器和通知配置。
- [ ] 在 `defaults.ts` 中提供 `createEmptyPipelineDraft()`、`createDefaultStage()`、`createDefaultJob()` 等工厂，供可视化新增动作复用。
- [ ] 在前端侧 `features/pipeline-draft/types.ts` 中补充仅供 UI 使用的派生类型，如 `SelectionTarget`、`InspectorMode`、`ValidationIssueViewModel`。
- [ ] 验收点：草稿模型足以表达截图中的“阶段列 + 任务卡片 + 并行任务 + 新建阶段”，且后续 YAML、表单、画布都围绕同一份模型工作。

### Task 4: 设计编排页壳层与页签组织

**Files:**
- Create: `apps/web/src/views/pipeline-editor/index.vue`
- Create: `apps/web/src/views/pipeline-editor/components/editor-toolbar.vue`
- Create: `apps/web/src/views/pipeline-editor/components/editor-tabs.vue`
- Create: `apps/web/src/stores/pipeline-ui.ts`

- [ ] 在编排页入口中组织顶部标题、状态徽标、模式切换、取消/保存/执行操作，不把这些操作分散到子组件。
- [ ] 用 `editor-tabs.vue` 统一管理六个业务页签，但让“任务编排”页签成为默认主入口，其余页签共享同一份 `PipelineDraft`。
- [ ] 在 `pipeline-ui` store 中集中管理 `activeTab`、`editorMode`、`inspectorVisible`、`yamlPanelVisible`、`fullscreen` 等纯 UI 状态。
- [ ] 顶部工具栏中的“可视化 / YAML”切换必须作用于同一页面上下文，避免切换时丢失选中对象与脏状态。
- [ ] 预留执行前校验和未保存提醒的视觉位置，但首版不在工具栏里塞入过多次级动作。
- [ ] 验收点：用户进入页面后先看到图形化编排区，所有非图形配置都能通过页签和侧板进入，不需要跳页。

### Task 5: 接入 X6 画布与图形化交互

**Files:**
- Create: `apps/web/src/views/pipeline-editor/components/pipeline-canvas.vue`
- Create: `apps/web/src/views/pipeline-editor/components/stage-column-node.ts`
- Create: `apps/web/src/views/pipeline-editor/components/job-card-node.ts`
- Create: `apps/web/src/features/pipeline-draft/graph-adapter.ts`
- Create: `apps/web/src/stores/pipeline-editor.ts`

- [ ] 在 `pipeline-canvas.vue` 中创建 X6 `Graph` 实例，并打开首版必须能力：网格、缩放、平移、选中、键盘、历史、小地图、剪贴板。
- [ ] 为阶段列定义自定义节点，让阶段标题、任务数、增删入口和阶段操作图标能独立渲染。
- [ ] 为任务卡片定义自定义节点，让任务标题、告警状态、加号入口、连接锚点和并行任务占位统一可控。
- [ ] 在 `graph-adapter.ts` 中只做映射与事件桥接：`draft -> graph cells`、`graph event -> draft command`，不在画布组件里直接改业务数据。
- [ ] 在 `pipeline-editor` store 中落地命令式更新：新增阶段、删除阶段、新增任务、删除任务、更新依赖、更新坐标、撤销重做。
- [ ] 首版只支持横向阶段列、纵向任务栈和基础依赖线，不提前引入复杂自动布局算法。
- [ ] 验收点：用户可以从空白画布开始，完成“新建阶段 -> 新建任务 -> 连线 -> 移动节点 -> 撤销重做”的完整主路径。

### Task 6: 构建右侧属性面板与表单编辑

**Files:**
- Create: `apps/web/src/views/pipeline-editor/components/inspector-panel.vue`
- Create: `apps/web/src/views/pipeline-editor/components/inspector-stage-form.vue`
- Create: `apps/web/src/views/pipeline-editor/components/inspector-job-form.vue`
- Create: `apps/web/src/views/pipeline-editor/components/inspector-edge-form.vue`
- Modify: `apps/web/src/stores/pipeline-editor.ts`

- [ ] 将属性编辑统一放在右侧固定侧栏或 `NDrawer`，不要把编辑表单嵌进节点内部，保持图面干净。
- [ ] 依据选中对象类型切换面板内容：阶段表单编辑名称、顺序、阶段说明；任务表单编辑名称、执行器、环境、条件；依赖表单编辑连线条件。
- [ ] 使用 `NForm`、`NInput`、`NSelect`、`NSwitch`、`NDynamicInput` 组织配置项，所有字段直接绑定到 `PipelineDraft`。
- [ ] 对高风险变更如“删除阶段”“删除任务”“清空依赖”统一走 `useDialog()` 二次确认，不在按钮点击后直接破坏性删除。
- [ ] 让属性面板对校验错误有可视反馈，例如字段错误、阶段缺少任务、依赖回环等。
- [ ] 验收点：所有画布主对象都能在右侧修改核心属性，图形节点只保留摘要，不承载长表单。

### Task 7: 补 YAML 工作区与双向转换

**Files:**
- Create: `apps/web/src/views/pipeline-editor/components/yaml-workspace.vue`
- Create: `apps/web/src/features/pipeline-draft/serializer.ts`
- Modify: `apps/web/src/views/pipeline-editor/index.vue`
- Modify: `apps/web/src/stores/pipeline-ui.ts`

- [ ] 将 YAML 模式定义为“专家模式”，入口仍在顶部工具栏，与图形化模式共享同一份 `PipelineDraft`。
- [ ] 在 `serializer.ts` 中先完成 `draft -> yaml` 的稳定序列化，保证阶段顺序、任务顺序、依赖关系和触发器能完整表达。
- [ ] 再补 `yaml -> draft` 解析，遇到不支持的高级字段时给出错误说明，而不是静默丢弃。
- [ ] `yaml-workspace.vue` 负责 Monaco 挂载、只读/可编辑切换、schema diagnostics 展示和错误列表，不直接承担业务规则。
- [ ] 切回图形化模式时，如果 YAML 有未解析错误，应阻止覆盖当前草稿，并明确提示冲突来源。
- [ ] 验收点：用户可以从图形化生成 YAML，也可以在 YAML 中微调后回到图形化，且不会因为视图切换导致语义漂移。

### Task 8: 接入 AJV 校验、保存流与执行前预检

**Files:**
- Create: `apps/web/src/features/pipeline-draft/validation.ts`
- Create: `apps/web/src/api/http.ts`
- Create: `apps/web/src/api/pipeline.ts`
- Create: `apps/web/src/api/types.ts`
- Modify: `apps/web/src/stores/pipeline-editor.ts`
- Modify: `apps/web/src/views/pipeline-editor/components/editor-toolbar.vue`

- [ ] 在 `validation.ts` 中接入 AJV，对共享 schema 做本地即时校验，并产出统一 `ValidationIssue` 结构。
- [ ] `editor-toolbar.vue` 在保存前先触发本地校验，错误时给出计数、摘要和定位入口，不直接发请求。
- [ ] 在 API 层区分“读取详情”“保存草稿”“执行前预检”“立即执行”四类接口，避免页面直接拼 URL。
- [ ] 保存成功后刷新 `draftVersion`、`lastSavedAt`、`dirty` 标识；保存失败时保留本地草稿，不回滚用户输入。
- [ ] 执行前预检只负责展示风险、缺失字段和阻断原因，不在首版承担完整模拟运行。
- [ ] 验收点：用户保存时能先看到本地规则错误，保存成功后状态明确，执行前能看到阻断项而不是盲点执行。

### Task 9: 补强交互细节与交付前检查项

**Files:**
- Modify: `apps/web/src/views/pipeline-editor/components/pipeline-canvas.vue`
- Modify: `apps/web/src/views/pipeline-editor/components/editor-toolbar.vue`
- Modify: `apps/web/src/views/pipeline-editor/components/inspector-panel.vue`
- Modify: `apps/web/src/app/theme/theme-overrides.ts`

- [ ] 完成基础快捷键映射：删除、复制、粘贴、撤销、重做、适配画布。
- [ ] 为空状态、新建入口、只读状态、未保存状态、校验失败状态补视觉反馈，避免只有“能用”没有“能懂”。
- [ ] 收敛颜色、阴影、边框、hover、选中态，让 Naive UI 表单层和 X6 画布层视觉语言统一。
- [ ] 检查键盘焦点、最小点击热区、左右栏滚动独立性、错误提示可达性，避免首版就在可用性上掉坑。
- [ ] 输出人工验收清单：图形化新建、属性编辑、YAML 切换、保存、预检、未保存拦截、错误定位。
- [ ] 验收点：页面在复杂度增加后仍保持“主路径清晰、错误可定位、状态可解释”。

## 4. 关键技术决策

### 4.1 为什么改用 Naive UI

- `NConfigProvider` 的全局主题覆盖更适合早期把设计 token 一次性收住。
- Provider 体系天然适合控制台里的全局消息、确认框、通知。
- `NForm`、`NDrawer`、`NTabs`、`NDataTable` 足够覆盖控制台型表单与侧板场景。

### 4.2 为什么仍保留 `PipelineDraft` 为唯一真相

- 避免把 X6 的 cell 结构误当成业务领域模型。
- 避免图形化模式和 YAML 模式出现双数据源漂移。
- 后续即使更换图引擎或补 CLI 导入，仍可复用同一份领域模型。

### 4.3 为什么 YAML 要后置

- 当前用户主路径是图形化编排，不应让 YAML 结构成为第一次使用门槛。
- 但 YAML 仍是对外配置交换格式，所以必须在首版保留双向转换边界。

## 5. 外部依赖与官方文档

- Vue 3 官方文档：https://vuejs.org/
- Naive UI 官方与仓库：https://www.naiveui.com/ 、https://github.com/tusen-ai/naive-ui
- Naive UI `NConfigProvider` / `NDialogProvider` / `NMessageProvider` / `NForm` / `NDrawer` / `NTabs` 用法：Context7 `/tusen-ai/naive-ui`
- AntV X6 官方文档：https://x6.antv.vision/en/docs/tutorial/getting-started/
- AJV 官方文档：https://ajv.js.org/

## 6. 风险与前置依赖

- 根工作区文件尚未初始化，真正开工时会触及根级 `package.json`、工作区配置和前端依赖，需要单独确认实施窗口。
- 后端 API 与共享 schema 还未落地，前端保存流需要允许临时 mock 或占位协议。
- YAML 双向转换是首版最容易返工的部分，必须先冻结共享类型，再接序列化。
- 如果后续引入更复杂审批流或条件分支，`EdgeDraft` 结构需要预留扩展字段。

## 7. 完成定义

- 用户能在图形化模式下完成一条基础流水线的创建与编辑。
- 图形化内容可以稳定生成 YAML，并支持从 YAML 回写到图形化。
- 本地校验、保存流、执行前预检形成闭环。
- 页面样式、交互和状态反馈达到“内部可演示、可继续扩展”的质量线。
