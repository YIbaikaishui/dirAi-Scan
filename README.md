🚀 DirAI-Scan

**下一代高并发智能 Web 扫描器 | 基于 Go 语言与 AI 增强**
**版本**: v0.1.0

---

## 🔍 项目简介

DirAI-Scan 是一个高性能、模块化的 Web 资源扫描工具，结合 **Go 语言的高并发能力** 和 **AI 智能分析**，支持快速识别 Web 资源、技术栈及潜在漏洞。

### 核心特性

- **高并发扫描引擎**：支持大规模目标的快速探测。
- **AI 智能决策**：基于 Ollama 的 AI 模型（如 `qwen3:4b`）辅助分析结果。
- **多模式扫描**：支持 `fast`（快速）、`deep`（深度）、`stealth`（隐蔽）模式。
- **现代技术栈识别**：精准识别前端框架、后端语言、CMS 等。
- **分布式支持**：支持控制器（`controller`）与工作节点（`worker`）架构。

---

## 📦 安装指南

### 1. **从 Go 模块安装**

```bash
go install github.com/yourusername/diraiscan@latest
```

### 2. **从源码构建**

```bash
git clone https://github.com/yourusername/diraiscan.git
cd diraiscan
go build -o diraiscan
```

---

## 🛠️ 使用说明

### 1. **基础命令**

```bash
diraiscan [flags]
diraiscan [command]
```

### 2. **可用命令**


| 命令         | 描述                    |
| ------------ | ----------------------- |
| `scan`       | 开始扫描目标 URL        |
| `version`    | 显示版本信息            |
| `help`       | 显示帮助信息            |
| `completion` | 生成 Shell 自动补全脚本 |

### 3. **常用标志**


| 标志               | 描述                               | 默认值        |
| ------------------ | ---------------------------------- | ------------- |
| `-u, --url`        | 目标 URL（必填）                   | N/A           |
| `-t, --threads`    | 并发线程数                         | `50`          |
| `-m, --mode`       | 扫描模式：`fast`/`deep`/`stealth`  | `fast`        |
| `-o, --output`     | 输出文件路径（支持 JSON/CSV/HTML） | N/A           |
| `--ai-enable`      | 启用 AI 分析（需 Ollama 服务）     | `false`       |
| `--proxy`          | 代理服务器（支持 HTTP/SOCKS5）     | N/A           |
| `--exclude-status` | 排除指定状态码的结果               | `404`         |
| `--status-codes`   | 仅显示指定状态码的结果             | `200,403,500` |

### 4. **完整参数列表**

运行 `diraiscan --help` 查看所有参数详情。

---

## 🧪 示例用法

### 1. **基础扫描**

```bash
diraiscan -u https://example.com
```

### 2. **启用 AI 分析**

```bash
diraiscan -u https://example.com --ai-enable --ai-model qwen3:4b
```

### 3. **自定义配置与输出**

```bash
diraiscan -u https://example.com -c config.json -o results.json --verbose
```

### 4. **分布式模式（控制器 + 工作节点）**

- 启动控制器：
  ```bash
  diraiscan --distribute controller
  ```
- 启动工作节点：
  ```bash
  diraiscan --distribute worker
  ```

---

## ⚙️ 配置管理

### 1. **配置文件**

- 默认配置文件：`config.json`
- 支持字段：
  - 扫描线程数、超时时间、用户代理
  - 日志目录与级别
  - 输出格式（JSON/CSV/HTML）
  - AI 模型参数（模型名、API 密钥、端点）

### 2. **环境变量（.env）**

- 敏感信息（如 API 密钥）应存储在 `.env` 文件中。
- 示例 `.env`：
  ```env
  AI_KEY=your_api_key_here
  AI_ENDPOINT=http://localhost:11434
  AI_ENABLED=true
  ```

---

## 📁 项目结构

```
diraiscan/
├── config/              # 配置管理模块
│   └── config.go        # 配置加载与默认值
├── main.go              # 程序入口
├── .env                 # 敏感配置（需 .gitignore）
├── config.json          # 主配置文件
└── README.md            # 项目文档
```

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

- 提交前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)（如未创建，请补充）。
- 确保代码符合 Go 编码规范，并通过单元测试。

---

## 📄 许可证

本项目采用 MIT License。详情见 [LICENSE](LICENSE) 文件。

---

## 📣 联系方式

- 作者：你的名字
- 邮箱：your@email.com
- GitHub：[https://github.com/yourusername/diraiscan](https://github.com/yourusername/diraiscan)

---

将以上内容保存为 `README.md` 文件，并根据实际项目信息调整 GitHub 仓库地址、作者信息等。此文档结构清晰，适合开发者快速上手和贡献代码。
