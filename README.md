# DirAI-Scan

一个基于AI增强的目录扫描工具，采用现代化的Go架构设计。

## 项目特性

- 🚀 **高性能扫描**: 基于协程的并发扫描，支持自定义并发数
- 🤖 **AI智能分析**: 集成AI模型对扫描结果进行智能分析和风险评估
- 📊 **多格式输出**: 支持JSON、YAML、CSV等多种输出格式
- ⚙️ **灵活配置**: 基于YAML的配置文件，支持环境变量覆盖
- 🔧 **模块化设计**: 采用DDD架构，代码结构清晰，易于扩展
- 📝 **完善日志**: 基于zap的结构化日志，支持日志轮转
- 🔌 **插件系统**: 支持插件扩展，可自定义扫描逻辑
- 🌐 **分布式支持**: 支持分布式扫描，可横向扩展

## 项目结构

```
dirAi-Scan/
├── cmd/                    # 命令行入口
│   └── dirmap/
│       ├── main.go        # 主程序入口
│       └── commands/      # Cobra命令定义
│           ├── root.go    # 根命令
│           ├── scan.go    # 扫描命令
│           └── analyze.go # 分析命令
├── internal/              # 内部包
│   ├── domain/           # 领域层
│   │   ├── model/        # 领域模型
│   │   └── service/      # 领域服务
│   ├── infra/           # 基础设施层
│   │   ├── ai/          # AI服务实现
│   │   ├── logging/     # 日志服务
│   │   └── messaging/   # 消息队列
│   └── pkg/             # 内部工具包
│       ├── config/      # 配置管理
│       ├── di/          # 依赖注入
│       ├── errors/      # 错误处理
│       └── utils/       # 工具函数
├── configs/             # 配置文件
│   └── dirmap.yaml     # 主配置文件
├── data/               # 数据文件
│   ├── common.txt     # 通用字典
│   ├── paths.txt      # 路径字典
│   └── dicc.txt       # 目录字典
├── logs/              # 日志目录
├── results/           # 扫描结果
├── plugins/           # 插件目录
├── go.mod            # Go模块文件
├── go.sum            # 依赖校验文件
└── README.md         # 项目说明
```

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build -o dirmap ./cmd/dirmap
```

### 基本使用

#### 1. 扫描目标

```bash
# 基本扫描
./dirmap scan -u http://example.com

# 指定字典文件
./dirmap scan -u http://example.com -w data/common.txt

# 启用AI分析
./dirmap scan -u http://example.com --ai-enable --ai-model deepseek-r1

# 自定义并发数和超时
./dirmap scan -u http://example.com -t 100 --timeout 10s
```

#### 2. 分析结果

```bash
# 分析扫描结果
./dirmap analyze -f results/scan_results_example.com_20240101_120000.json

# 指定AI模型进行分析
./dirmap analyze -f results/scan_results.json --ai-model deepseek-r1
```

### 配置文件

项目使用YAML格式的配置文件，位于 `configs/dirmap.yaml`。主要配置项包括：

```yaml
# 应用配置
app:
  name: "DirAI-Scan"
  version: "1.0.0"
  debug: false

# 扫描配置
scan:
  concurrency: 50
  timeout: "5s"
  retries: 3
  delay: "0ms"

# AI配置
ai:
  enabled: false
  endpoint: "http://localhost:11434"
  model: "deepseek-r1"
  timeout: "30s"

# 日志配置
logging:
  level: "info"
  dir: "./logs"
  max_size: 100
  max_backups: 3
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
export DIRAISCAN_AI_ENABLED=true
export DIRAISCAN_AI_ENDPOINT=http://localhost:11434
export DIRAISCAN_AI_MODEL=deepseek-r1
export DIRAISCAN_SCAN_CONCURRENCY=100
```

## 架构设计

### 领域驱动设计 (DDD)

项目采用DDD架构，分为以下几层：

1. **领域层 (Domain Layer)**
   - `model/`: 核心业务模型和实体
   - `service/`: 领域服务接口

2. **基础设施层 (Infrastructure Layer)**
   - `ai/`: AI服务的具体实现
   - `logging/`: 日志服务实现
   - `messaging/`: 消息队列实现

3. **应用层 (Application Layer)**
   - `commands/`: 命令行接口实现

4. **工具层 (Utility Layer)**
   - `config/`: 配置管理
   - `di/`: 依赖注入容器
   - `errors/`: 统一错误处理

### 核心组件

#### 1. 配置管理

基于Viper的配置管理系统，支持：
- YAML配置文件
- 环境变量覆盖
- 配置热重载
- 配置验证

#### 2. 日志系统

基于Zap的高性能日志系统：
- 结构化日志
- 日志轮转
- 多级别日志
- 文件和控制台输出

#### 3. 依赖注入

轻量级DI容器：
- 服务注册和解析
- 工厂模式支持
- 生命周期管理

#### 4. 错误处理

统一的错误处理机制：
- 错误码定义
- 错误包装和链式调用
- 调用栈追踪

## 开发指南

### 添加新功能

1. **定义领域模型**: 在 `internal/domain/model/` 中定义新的实体和值对象
2. **定义服务接口**: 在 `internal/domain/service/` 中定义服务接口
3. **实现基础设施**: 在 `internal/infra/` 中实现具体的服务
4. **添加命令**: 在 `cmd/dirmap/commands/` 中添加新的命令
5. **更新配置**: 在配置文件中添加相关配置项

### 代码规范

- 遵循Go官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释和文档
- 编写单元测试
- 使用依赖注入而非全局变量

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/pkg/config

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 部署

### 单机部署

```bash
# 构建二进制文件
go build -o dirmap ./cmd/dirmap

# 创建配置文件
cp configs/dirmap.yaml /etc/diraiscan/

# 运行
./dirmap scan -u http://example.com
```

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o dirmap ./cmd/dirmap

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dirmap .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/data ./data
CMD ["./dirmap"]
```

### 分布式部署

1. **控制器节点**:
```bash
./dirmap --mode controller --bind 0.0.0.0:8080
```

2. **工作节点**:
```bash
./dirmap --mode worker --controller http://controller:8080
```

## 贡献

欢迎提交Issue和Pull Request！

### 贡献流程

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 更新日志

### v1.0.0
- 初始版本发布
- 基础扫描功能
- AI分析集成
- 配置管理系统
- 日志系统
- 依赖注入容器

## 联系方式

- 项目主页: [GitHub](https://github.com/diraiscan/dirAi-Scan)
- 问题反馈: [Issues](https://github.com/diraiscan/dirAi-Scan/issues)
