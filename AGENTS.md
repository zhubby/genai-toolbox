# Agent 指南（本仓库）

## 语言与沟通

- **始终使用中文（简体）回复**；专业术语可保留英文（如 MCP、OpenTelemetry、Toolset）。
- 优先给出可执行结论：如何复现/运行、改动点、验证方式。

## 项目概述

本仓库是 **MCP Toolbox for Databases**（原 “Gen AI Toolbox for Databases”）的源码：一个面向数据库与数据源的 **MCP Server**。它位于应用/编排框架与数据源之间，通过一个统一的控制面（control plane）来**配置、发布与调用工具（tools）**，并内置数据库侧的通用复杂度处理能力：

- 以 `tools.yaml` 为核心的声明式配置：`sources`（数据源）、`tools`（能力）、`toolsets`（工具集合）、`prompts`（提示词资源）。
- 多数据源适配：Postgres/MySQL/SQL Server/Spanner/BigQuery/Redis/Neo4j 等（见 `internal/sources/`、`internal/tools/`）。
- 安全与鉴权：支持基于请求头的 auth service/claims 机制（见 `internal/auth/` 与 `internal/server/api.go`）。
- 可观测性：内置 OpenTelemetry metrics/tracing（见 `internal/telemetry/`）。

服务形态上，它是一个 Go 实现的 HTTP 服务（默认监听 `:5000`），同时提供：

- `/mcp`：MCP 协议入口（含 SSE/POST 通道，见 `internal/server/mcp.go`）
- `/api`：工具/工具集查询与调用接口（见 `internal/server/api.go`）

## 代码结构速览

- `main.go`：入口，调用 `cmd.Execute()`。
- `cmd/`：CLI 与启动流程（Cobra），负责加载/校验 `tools.yaml`、初始化 Server、信号处理等。
- `internal/server/`：HTTP Server、路由、MCP 协议实现、Web/API。
- `internal/sources/`：数据源连接与方言/能力抽象（按产品分目录）。
- `internal/tools/`：工具实现与注册（大量包通过 side-effect import 进行注册）。
- `internal/auth/`：鉴权服务抽象与实现。
- `internal/telemetry/`：OpenTelemetry 初始化与指标/链路埋点。
- `internal/prebuiltconfigs/`：预置工具/配置样例（YAML）。
- `tests/`：集成测试（部分依赖云资源/Build Tags）。
- `docs/`：文档（Hugo）。
- `ui/`：前端（Next.js）；当前 `ui/README.md` 为默认模板，具体用途以实际路由/产品规划为准。

## 核心概念（与 `tools.yaml` 对齐）

- `sources`：声明数据源连接信息与种类（`kind`），工具通常引用某个 source。
- `tools`：声明或引用某类工具（`kind`），并绑定 source、参数与语句/行为。
- `toolsets`：将多个 tool 按场景分组，便于按 agent/应用加载。
- `prompts`：用于与 LLM 交互的提示词资源定义。

## 开发与验证（常用命令）

- 查看 CLI 参数：`go run . --help`
- 本地运行（示例）：`go run . --tools-file tools.yaml`（默认端口 `5000`）
- 健康检查：`curl http://127.0.0.1:5000`
- 单元测试：`go test -race -v ./cmd/... ./internal/...`
- Lint：`golangci-lint run --fix`
- 集成测试：`go test -race -v ./tests/<DIR>`（具体环境变量与资源见 `.ci/` 与 `DEVELOPER.md`）

## 贡献与实现规范（面向代码改动）

- Go 风格：遵循 Effective Go 与常见工程约定（清晰的错误处理、避免隐式全局副作用、并发安全）。
- 变更边界：**只改动与需求直接相关的内容**；不要顺手重构大范围代码或改格式。
- 可观测性优先：新增/调整对外路径（API/MCP）时，尽量保持埋点一致性（trace span、指标计数、状态标签）。
- 安全默认：涉及鉴权/授权逻辑时，默认拒绝（fail closed），避免扩大权限面。

## 命名约定（Tools）

参见 `DEVELOPER.md`：

- **Tool name**：建议使用下划线（`list_tables`），避免产品前缀；名称微调通常不视为破坏性变更。
- **Tool kind**：建议使用连字符并带产品名（如 `firestore-list-collections`）；变更通常视为破坏性变更，应谨慎。
