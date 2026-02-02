# drone-gemini-cli-plugin

[English](README.md) | 中文

适用于 Drone CI/CD 的 Gemini CLI 插件，基于 Google Gemini CLI 的 Headless 模式，用于代码分析和审查。

## 功能

- 代码审查：检测安全漏洞、Bug 和代码质量问题
- YOLO 模式：自动批准 AI 操作，允许自动修改文件
- 工具调用：支持 Bash、文件读写、Web 搜索
- Git 集成：分析指定 commit 或 PR diff
- 双重认证：支持 Gemini API Key 和 Vertex AI 服务账号

## 快速开始

### 方式 A：Gemini API Key

从 [Google AI Studio](https://aistudio.google.com/apikey) 获取免费 API Key。

```bash
drone secret add --repository your-org/your-repo \
  --name gemini_api_key --data "你的API密钥"
```

```yaml
kind: pipeline
type: docker
name: ai-review

steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "检查这段代码的 Bug 和安全问题"
      model: gemini-2.5-flash
      api_key:
        from_secret: gemini_api_key
```

### 方式 B：Vertex AI 服务账号

适用于企业环境或非 GCP 基础设施。

```bash
# 创建服务账号
gcloud iam service-accounts create gemini-cli-sa
gcloud projects add-iam-policy-binding YOUR_PROJECT \
  --member="serviceAccount:gemini-cli-sa@YOUR_PROJECT.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# 下载密钥并添加到 Drone
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=gemini-cli-sa@YOUR_PROJECT.iam.gserviceaccount.com

drone secret add --repository your-org/your-repo \
  --name gcp_credentials --data @sa-key.json
```

```yaml
steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "审查这段代码的安全问题"
      model: gemini-3-pro-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

## 配置参数

| 参数 | 环境变量 | 类型 | 默认值 | 说明 |
|-----|---------|------|-------|------|
| `prompt` | `PLUGIN_PROMPT` | string | 必填 | AI 提示词 |
| `target` | `PLUGIN_TARGET` | string | `.` | 工作目录 |
| `model` | `PLUGIN_MODEL` | string | `gemini-2.5-pro` | 使用的模型 |
| `output_format` | `PLUGIN_OUTPUT_FORMAT` | string | `json` | 输出格式 |
| `yolo` | `PLUGIN_YOLO` | bool | `false` | 自动批准所有操作 |
| `api_key` | `PLUGIN_API_KEY` | string | | Gemini API Key |
| `gcp_project` | `PLUGIN_GCP_PROJECT` | string | | GCP 项目 ID |
| `gcp_location` | `PLUGIN_GCP_LOCATION` | string | `us-central1` | GCP 区域 |
| `gcp_credentials` | `PLUGIN_GCP_CREDENTIALS` | string | | 服务账号 JSON |
| `git_diff` | `PLUGIN_GIT_DIFF` | bool | `false` | 包含 git diff |
| `timeout` | `PLUGIN_TIMEOUT` | int | `300` | 超时时间(秒) |
| `debug` | `PLUGIN_DEBUG` | bool | `false` | 调试模式 |

## 使用示例

### PR 代码审查

```yaml
steps:
  - name: ai-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: |
        审查此 PR:
        - 安全漏洞
        - 性能问题
        - 代码质量
      git_diff: true
      model: gemini-2.5-flash
      api_key:
        from_secret: gemini_api_key
    when:
      event: pull_request
```

### YOLO 模式自动修复

注意：YOLO 模式允许 AI 自动修改文件，请谨慎使用。

```yaml
steps:
  - name: auto-fix
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "修复所有 lint 错误并格式化代码"
      yolo: true
      api_key:
        from_secret: gemini_api_key
```

## 本地测试

```bash
# 构建插件
go build -o drone-gemini-cli-plugin .

# 使用 API Key 测试
PLUGIN_PROMPT="描述这个项目" \
PLUGIN_API_KEY="your-api-key" \
PLUGIN_MODEL="gemini-2.5-flash" \
./drone-gemini-cli-plugin
```

## 开源协议

Apache License 2.0
