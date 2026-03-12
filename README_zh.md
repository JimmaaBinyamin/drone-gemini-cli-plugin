# drone-gemini-cli-plugin

[English](README.md) | 中文

适用于 Drone CI/CD 的 Gemini CLI 插件，基于 Google Gemini CLI 的 Headless 模式，实现 AI 代码分析、审查和自动修复。

## 功能

- **AI 代码审查**：检测安全漏洞、Bug 和代码质量问题
- **文件加载 Prompt**：从仓库中的 Markdown 文件加载审查规范（`prompt_file`）
- **上下文文件**：提供架构文档等额外上下文（`context_file`）
- **YOLO 模式**：AI 自动批准所有操作，支持自动修改代码
- **AI Agent**：自主读取文件、执行 Shell 命令、分析代码
- **Git 集成**：自动获取本次提交的 diff，进行针对性审查
- **详细统计**：Token 用量、成本估算、工具调用明细
- **双重认证**：支持 Gemini API Key 和 Vertex AI 服务账号

## 快速开始

### 方式 A：Vertex AI 服务账号（推荐）

适用于企业环境或非 GCP 基础设施（AWS/Azure/自建机房）。

```bash
# 1. 创建 GCP 服务账号并授权
gcloud iam service-accounts create gemini-drone-sa
gcloud projects add-iam-policy-binding YOUR_PROJECT \
  --member="serviceAccount:gemini-drone-sa@YOUR_PROJECT.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# 2. 下载 JSON 密钥并添加到 Drone
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=gemini-drone-sa@YOUR_PROJECT.iam.gserviceaccount.com

drone secret add --repository your-org/your-repo \
  --name gcp_credentials --data @sa-key.json
```

```yaml
kind: pipeline
type: docker
name: ai-review

steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "审查代码中的安全漏洞和 Bug"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### 方式 B：Gemini API Key（最简单）

从 [Google AI Studio](https://aistudio.google.com/apikey) 获取免费 API Key。

```bash
drone secret add --repository your-org/your-repo \
  --name gemini_api_key --data "你的API密钥"
```

```yaml
steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "审查代码中的 Bug 和安全问题"
      model: gemini-3-flash-preview
      api_key:
        from_secret: gemini_api_key
```

## 配置参数

| 参数 | 环境变量 | 类型 | 默认值 | 说明 |
|-----|---------|------|-------|------|
| `prompt` | `PLUGIN_PROMPT` | string | 必填* | AI 提示词（*设置了 `prompt_file` 时非必填） |
| `prompt_file` | `PLUGIN_PROMPT_FILE` | string | | 从文件加载 prompt（覆盖 `prompt`） |
| `context_file` | `PLUGIN_CONTEXT_FILE` | string | | 从文件加载额外上下文（通过 stdin 传递） |
| `target` | `PLUGIN_TARGET` | string | `.` | 工作目录 |
| `model` | `PLUGIN_MODEL` | string | `gemini-2.5-pro` | 使用的模型（推荐：`gemini-3-flash-preview`） |
| `output_format` | `PLUGIN_OUTPUT_FORMAT` | string | `json` | 输出格式：`text`、`json`、`stream-json` |
| `yolo` | `PLUGIN_YOLO` | bool | `false` | 自动批准所有操作（允许修改文件） |
| `approval_mode` | `PLUGIN_APPROVAL_MODE` | string | | 覆盖审批模式 |
| `include_dirs` | `PLUGIN_INCLUDE_DIRS` | string | | 限定目录（逗号分隔） |
| `stdin_input` | `PLUGIN_STDIN_INPUT` | string | | 通过 stdin 传递的额外内容 |
| `api_key` | `PLUGIN_API_KEY` | string | | Gemini API Key |
| `gcp_project` | `PLUGIN_GCP_PROJECT` | string | | GCP 项目 ID（Vertex AI） |
| `gcp_location` | `PLUGIN_GCP_LOCATION` | string | `us-central1` | GCP 区域（gemini-3-* 模型用 `global`） |
| `gcp_credentials` | `PLUGIN_GCP_CREDENTIALS` | string | | 服务账号 JSON 内容 |
| `git_diff` | `PLUGIN_GIT_DIFF` | bool | `false` | 包含本次提交的 git diff |
| `timeout` | `PLUGIN_TIMEOUT` | int | `300` | 超时时间（秒） |
| `debug` | `PLUGIN_DEBUG` | bool | `false` | 调试模式 |

## 使用示例

### 1. 基础代码审查（内联 Prompt + Git Diff）

```yaml
steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: |
        审查这段代码：
        - 安全漏洞
        - 潜在 Bug
        - 代码质量问题
        给出简要总结。
      git_diff: true
      model: gemini-3-flash-preview
      output_format: json
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      event: pull_request
```

### 2. 文件加载审查规范（prompt_file）

将审查标准写在 Markdown 文件中，插件会自动加载为 prompt：

```yaml
steps:
  - name: review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt_file: ".review-guidelines.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

`.review-guidelines.md` 示例：

```markdown
# 代码审查指南

你是一名资深安全工程师。检查以下内容：
- SQL 注入：是否使用参数化查询？
- XSS：用户输入是否正确转义？
- 敏感数据：密码/Token 是否暴露在 API 响应中？

输出格式：CRITICAL / WARNING / INFO
```

### 3. 架构合规审查（context_file）

提供架构文档作为额外上下文，AI 会对照文档审查代码：

```yaml
steps:
  - name: review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "对照架构文档审查代码合规性"
      context_file: "docs/architecture.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### 4. 完整审查（prompt_file + context_file）

同时加载审查规范和架构文档，实现最全面的代码审查：

```yaml
steps:
  - name: full-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt_file: ".review-guidelines.md"
      context_file: "docs/architecture.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### 5. YOLO 模式 - AI 自动修复

> ⚠️ **警告**：YOLO 模式允许 AI 自动修改文件和执行命令。请仅在非保护分支上使用。

```yaml
steps:
  - name: auto-fix
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: |
        修复代码中的 SQL 注入漏洞，改用参数化查询。
        修复 XSS 漏洞，对用户输入进行转义。
        只修改源文件，不要创建新文件。
        修复后展示你的修改 diff。
      yolo: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      branch: fix/*
```

### 6. 生成 Release Notes

```yaml
steps:
  - name: release-notes
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "根据最近的提交生成 CHANGELOG 格式的 Release Notes"
      git_diff: true
      output_format: text
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      event: tag
```

## 本地测试

```bash
# 构建插件
go build -o drone-gemini-cli-plugin .

# 使用 Vertex AI 测试
PLUGIN_PROMPT="描述这个项目" \
PLUGIN_GCP_PROJECT="your-project-id" \
PLUGIN_GCP_LOCATION="global" \
PLUGIN_GCP_CREDENTIALS="$(cat service-account.json)" \
PLUGIN_MODEL="gemini-3-flash-preview" \
./drone-gemini-cli-plugin

# 使用 API Key 测试
PLUGIN_PROMPT="描述这个项目" \
PLUGIN_API_KEY="your-api-key" \
PLUGIN_MODEL="gemini-3-flash-preview" \
./drone-gemini-cli-plugin

# YOLO 模式测试（允许修改文件）
PLUGIN_PROMPT="修复 main.go 中的安全漏洞" \
PLUGIN_GCP_PROJECT="your-project-id" \
PLUGIN_GCP_LOCATION="global" \
PLUGIN_GCP_CREDENTIALS="$(cat service-account.json)" \
PLUGIN_YOLO="true" \
./drone-gemini-cli-plugin
```

## 构建 Docker 镜像

```bash
docker build -t ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin .
docker push ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
```

## Drone 环境变量

插件自动接收以下 Drone 环境变量：

| 变量 | 说明 |
|------|------|
| `DRONE_COMMIT_SHA` | 当前提交 SHA（用于 git_diff） |
| `DRONE_REPO_NAME` | 仓库名称 |
| `DRONE_BUILD_EVENT` | 构建事件类型（push、pull_request、tag） |

## 开源协议

Apache License 2.0
