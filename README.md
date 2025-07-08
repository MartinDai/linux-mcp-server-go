# Linux MCP Server Go

基于 Go 语言实现的模型上下文协议（MCP）服务器，通过 SSH 提供远程 Shell 执行功能。

## 功能特性

- **远程 Shell 执行**: 通过 SSH 在远程 Linux 机器上执行 Shell 命令
- **配置驱动**: 通过 JSON 配置文件管理多个远程主机
- **跨平台支持**: 支持多种操作系统和架构
- **MCP 兼容**: 完全兼容 MCP 客户端和工具
- **双模式支持**: 支持 stdio 和 HTTP 传输模式

## 提供的工具

### execute_shell

通过 SSH 连接在远程机器上执行 Shell 命令。

**参数:**
- `machine_ip` (string): 目标机器的 IP 地址
- `path` (string): 远程机器上的工作目录路径
- `shell` (string): 要执行的 Shell 命令

**使用示例:**
```json
{
  "tool": "execute_shell",
  "arguments": {
    "machine_ip": "192.168.1.100",
    "path": "/home/user",
    "shell": "ls -la"
  }
}
```

## 配置

### hosts.json

在服务器二进制文件同目录下创建 `hosts.json` 文件来配置 SSH 连接：

```json
[
  {
    "ip": "192.168.1.100",
    "user": "root",
    "password": "your_password",
    "port": 22
  },
  {
    "ip": "192.168.1.101",
    "user": "admin",
    "password": "admin_password",
    "port": 2222
  }
]
```

**配置字段:**
- `ip`: 远程机器的 IP 地址或主机名
- `user`: SSH 用户名
- `password`: SSH 密码（明文）
- `port`: SSH 端口（默认：22）

## 安装

### 前置要求

- Go 1.24.4 或更高版本
- Git

### 从源码构建

1. 克隆仓库：
```bash
git clone <repository-url>
cd linux-mcp-server-go
```

2. 安装依赖：
```bash
make deps
```

3. 构建服务器：
```bash
make build
```

### 跨平台构建

构建所有支持的平台：
```bash
make build-all
```

构建特定平台：
```bash
make build-linux-amd64      # Linux 64位
make build-darwin-arm64     # macOS Apple Silicon
make build-windows-amd64    # Windows 64位
```

## 使用方法

### Stdio 模式（默认）

以 stdio 模式运行服务器，用于直接与 MCP 客户端通信：
```bash
./build/linux-mcp-server
```

### HTTP 模式

在指定端口以 HTTP 模式运行服务器：
```bash
./build/linux-mcp-server -http :8080
```

或使用 Makefile：
```bash
make run-http
```

### 与 MCP 客户端集成

配置你的 MCP 客户端使用此服务器。Claude Desktop 配置示例：

```json
{
  "mcpServers": {
    "linux-mcp-server": {
      "command": "/path/to/linux-mcp-server",
      "args": []
    }
  }
}
```

## 开发

### 可用的 Make 目标

```bash
make help           # 显示所有可用目标
make build          # 为当前平台构建
make test           # 运行测试
make fmt            # 格式化代码
make vet            # 运行 go vet
make lint           # 运行 golangci-lint（需要安装 golangci-lint）
make clean          # 清理构建目录
```

### 项目结构

```
.
├── main.go         # 主服务器实现
├── hosts.json      # SSH 主机配置
├── go.mod          # Go 模块定义
├── go.sum          # Go 模块校验和
├── Makefile        # 构建自动化
└── README.md       # 本文件
```

## 安全考虑

⚠️ **重要安全注意事项：**

1. **密码存储**: 此实现将 SSH 密码以明文形式存储在 `hosts.json` 文件中。生产环境建议：
   - 使用基于 SSH 密钥的认证
   - 加密配置文件
   - 使用环境变量存储敏感数据

2. **网络安全**: 确保 SSH 连接在安全网络上进行

3. **访问控制**: 使用适当的文件权限限制对 `hosts.json` 文件的访问：
   ```bash
   chmod 600 hosts.json
   ```

## 依赖项

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - MCP 实现
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH 客户端实现

## 许可证

此项目使用 MIT 许可证 - 详情请参阅 LICENSE 文件。

## 贡献

1. Fork 此仓库
2. 创建功能分支
3. 进行修改
4. 如适用，添加测试
5. 运行 `make fmt` 和 `make vet`
6. 提交 Pull Request

## 故障排除

### 常见问题

1. **"hosts.json not found"**: 确保 `hosts.json` 文件存在于二进制文件同目录下
2. **SSH 连接失败**: 验证网络连接和 `hosts.json` 中的凭据
3. **权限被拒绝**: 检查目标机器上的 SSH 用户权限
4. **命令未找到**: 确保 Shell 命令在目标系统上存在

### 调试模式

通过设置环境变量启用调试日志：
```bash
MCP_DEBUG=1 ./build/linux-mcp-server
```