# ChatMitm - 流式接口拦截工具

基于 [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy) 实现的通用流式接口数据拦截工具，可以拦截并保存任意流式接口（SSE、流式 JSON 等）的响应数据。

## 功能特性

- ✅ 自动检测并拦截流式响应（Server-Sent Events、流式 JSON 等）
- ✅ 自动保存流式数据到本地文件
- ✅ 支持所有流式接口，无需配置
- ✅ 提供 Web 界面查看请求详情
- ✅ 不修改原始数据流，只进行监控和保存

## 安装

### 方式一：下载预编译二进制（推荐）

#### 从 GitHub Releases 下载

访问 [GitHub Releases](https://github.com/hello--world/chatMitm/releases) 下载适合你系统的二进制文件：

- **Linux AMD64**: `chatMitm-linux-amd64`
- **Linux ARM64**: `chatMitm-linux-arm64`
- **macOS AMD64**: `chatMitm-darwin-amd64`
- **macOS ARM64** (Apple Silicon): `chatMitm-darwin-arm64`
- **Windows AMD64**: `chatMitm-windows-amd64.exe`
- **Windows ARM64**: `chatMitm-windows-arm64.exe`

#### 使用方法

**Linux/macOS:**

```bash
# 下载
wget https://github.com/hello--world/chatMitm/releases/latest/download/chatMitm-linux-amd64

# 设置执行权限
chmod +x chatMitm-linux-amd64

# 运行
./chatMitm-linux-amd64
```

**Windows:**

```cmd
# 下载后直接运行
chatMitm-windows-amd64.exe
```

**使用 curl 下载（Linux/macOS）:**

```bash
# 自动检测最新版本并下载
curl -L https://github.com/hello--world/chatMitm/releases/latest/download/chatMitm-linux-amd64 -o chatMitm
chmod +x chatMitm
./chatMitm
```

### 方式二：使用 Docker

#### 从 GitHub Container Registry 拉取镜像

镜像发布在 [GitHub Container Registry](https://github.com/hello--world/chatMitm/pkgs/container/chatmitm)，可以使用以下命令拉取：

```bash
# 拉取最新版本
docker pull ghcr.io/hello--world/chatmitm:v1.0.0

# 或拉取主版本（推荐）
docker pull ghcr.io/hello--world/chatmitm:v1

# 或拉取完整版本号
docker pull ghcr.io/hello--world/chatmitm:v1.0
```

> **注意**：由于 GitHub Container Registry 的权限设置，首次拉取可能需要登录：
> ```bash
> echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
> ```
> 或者使用个人访问令牌（PAT）进行认证。

#### 运行容器

```bash
docker run -d \
  --name chatmitm \
  -p 9080:9080 \
  -p 9081:9081 \
  -v $(pwd)/stream_data:/app/stream_data \
  -v mitmproxy-certs:/home/appuser/.mitmproxy \
  ghcr.io/hello--world/chatmitm:v1.0.0
```

> **注意**：证书会保存在 Docker volume `mitmproxy-certs` 中。如果需要访问证书文件，可以使用：
> ```bash
> docker run --rm -v mitmproxy-certs:/data alpine cat /data/mitmproxy-ca-cert.pem > mitmproxy-ca-cert.pem
> ```

#### 使用 Docker Compose（推荐）

项目已包含 `docker-compose.yml` 文件，直接运行：

```bash
# 拉取最新镜像并启动
docker-compose pull
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

`docker-compose.yml` 配置：

```yaml
version: '3.8'

services:
  chatmitm:
    image: ghcr.io/hello--world/chatmitm:v1.0.0
    container_name: chatmitm
    ports:
      - "9080:9080"  # 代理端口
      - "9081:9081"  # Web 界面端口
    volumes:
      - ./stream_data:/app/stream_data  # 流式数据保存目录
      - mitmproxy-certs:/home/appuser/.mitmproxy  # 证书目录（持久化）
    restart: unless-stopped
    environment:
      - TZ=Asia/Shanghai

volumes:
  mitmproxy-certs:
    driver: local
```

### 方式三：从源码编译

#### 前置要求

- Go 1.22 或更高版本

#### 安装依赖

```bash
go mod download
```

#### 启动代理服务器

```bash
go run main.go
```

或者编译后运行：

```bash
go build -o chatMitm
./chatMitm
```

启动后，你会看到：

```
=========================================
流式接口拦截代理已启动
代理地址: http://127.0.0.1:9080
Web 界面: http://127.0.0.1:9081
流式数据保存目录: ./stream_data
=========================================
请配置你的客户端使用代理: http://127.0.0.1:9080
首次使用需要安装证书: ~/.mitmproxy/mitmproxy-ca-cert.pem
=========================================
```

### 2. 安装 HTTPS 证书（重要）

#### 使用 Docker 时获取证书

证书保存在 Docker volume 中，可以通过以下方式获取：

```bash
# 从 volume 中复制证书到当前目录
docker run --rm -v mitmproxy-certs:/data alpine cat /data/mitmproxy-ca-cert.pem > mitmproxy-ca-cert.pem
```

或者如果使用 docker-compose，证书在 volume `mitmproxy-certs` 中：

```bash
docker-compose exec chatmitm cat /home/appuser/.mitmproxy/mitmproxy-ca-cert.pem > mitmproxy-ca-cert.pem
```

为了拦截 HTTPS 流量，需要安装根证书：

1. 首次运行后，证书会自动生成在 `~/.mitmproxy/mitmproxy-ca-cert.pem`
2. 安装证书到系统信任列表：
   - **macOS**: 双击证书文件，在钥匙串访问中信任该证书
   - **Linux**: 参考系统文档安装 CA 证书
   - **Windows**: 双击证书文件，安装到"受信任的根证书颁发机构"

### 3. 配置客户端使用代理

#### 浏览器配置

- **Chrome/Edge**: 设置 → 系统 → 打开计算机的代理设置 → 配置 HTTP 代理为 `127.0.0.1:9080`
- **Firefox**: 设置 → 网络设置 → 手动代理配置 → HTTP 代理 `127.0.0.1` 端口 `9080`

#### 命令行工具（curl）

```bash
export http_proxy=http://127.0.0.1:9080
export https_proxy=http://127.0.0.1:9080
```

#### Node.js / fetch API

```javascript
// 使用环境变量
process.env.HTTP_PROXY = 'http://127.0.0.1:9080';
process.env.HTTPS_PROXY = 'http://127.0.0.1:9080';
```

### 4. 查看拦截的数据

#### 控制台输出

当检测到流式响应时，控制台会输出：

```
[流式响应检测] POST https://chat.deepseek.com/api/v0/chat/completion - Content-Type: text/event-stream; charset=utf-8
[开始拦截] 保存流式数据到: ./stream_data/20241120_084215_api_v0_chat_completion.txt
[完成拦截] 已保存 150 行数据到: ./stream_data/20241120_084215_api_v0_chat_completion.txt
[流式响应摘要] POST https://chat.deepseek.com/api/v0/chat/completion - 共 150 行数据
```

#### 文件保存

所有流式数据会自动保存到 `./stream_data/` 目录，文件名格式：
```
时间戳_路径.txt
例如: 20241120_084215_api_v0_chat_completion.txt
```

#### Web 界面

访问 http://localhost:9081/ 可以查看所有请求的详细信息，包括：
- 请求和响应头
- 请求和响应体
- 时间线信息

## 支持的流式格式

工具会自动检测以下 Content-Type 的流式响应：

- `text/event-stream` (Server-Sent Events)
- `application/stream+json`
- `application/x-ndjson` (Newline Delimited JSON)

## 自定义配置

可以修改 `main.go` 中的配置：

```go
opts := &proxy.Options{
    Addr:              ":9080",        // 代理端口
    StreamLargeBodies: 1024 * 1024 * 10, // 流式数据大小限制
    WebAddr:           ":9081",       // Web 界面端口
}

// 自定义输出目录
interceptor := NewStreamInterceptor("./custom_output_dir")
```

## 工作原理

1. **代理服务器**: 在本地 9080 端口启动 HTTP/HTTPS 代理
2. **流量拦截**: 所有经过代理的请求都会被拦截
3. **流式检测**: 自动检测响应头中的 `Content-Type` 是否为流式格式
4. **数据保存**: 使用 `StreamResponseModifier` 在数据流传输过程中同时保存到文件
5. **透明传输**: 原始数据流不受影响，客户端正常接收数据

## 注意事项

1. **证书安装**: 必须安装根证书才能拦截 HTTPS 流量
2. **防火墙**: 确保 9080 和 9081 端口未被占用
3. **性能**: 大流量场景下可能影响性能，建议仅在需要时使用
4. **隐私**: 所有流量数据都会被记录，请注意隐私安全

## 示例：拦截 DeepSeek Chat API

1. 启动代理：`go run main.go`
2. 配置浏览器使用代理 `127.0.0.1:9080`
3. 访问 https://chat.deepseek.com 并发送消息
4. 查看 `./stream_data/` 目录中的保存文件

## 故障排除

### 无法拦截 HTTPS 流量

- 检查证书是否已正确安装
- 确认客户端已配置使用代理
- 查看控制台是否有错误信息

### 没有检测到流式响应

- 确认响应头包含 `text/event-stream` 等流式 Content-Type
- 检查 Web 界面 (http://localhost:9081/) 查看实际响应头

### 数据保存失败

- 检查 `./stream_data/` 目录是否有写入权限
- 查看控制台错误信息

## License

MIT License

## 下载和安装

### GitHub Releases

所有预编译的二进制文件都发布在 [GitHub Releases](https://github.com/hello--world/chatMitm/releases)。

#### 快速下载脚本

**Linux/macOS:**

```bash
# 自动下载最新版本（Linux AMD64）
VERSION=$(curl -s https://api.github.com/repos/hello--world/chatMitm/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L https://github.com/hello--world/chatMitm/releases/download/${VERSION}/chatMitm-linux-amd64 -o chatMitm
chmod +x chatMitm
```

**Windows (PowerShell):**

```powershell
# 下载最新版本
$version = (Invoke-RestMethod https://api.github.com/repos/hello--world/chatMitm/releases/latest).tag_name
Invoke-WebRequest -Uri "https://github.com/hello--world/chatMitm/releases/download/$version/chatMitm-windows-amd64.exe" -OutFile "chatMitm.exe"
```

#### 验证文件完整性

下载后可以使用 checksums.txt 验证文件：

```bash
# 下载 checksums.txt
curl -L https://github.com/hello--world/chatMitm/releases/latest/download/checksums.txt -o checksums.txt

# 验证
sha256sum -c checksums.txt
```

### Docker 镜像

### GitHub Container Registry

Docker 镜像已发布到 [GitHub Container Registry](https://github.com/hello--world/chatMitm/pkgs/container/chatmitm)：

**镜像地址**：`ghcr.io/hello--world/chatmitm`

### 可用版本标签

- `v1.0.0` - 完整版本号（推荐用于生产环境）
- `v1.0` - 主版本.次版本（自动指向最新的 1.0.x 版本）
- `v1` - 主版本（自动指向最新的 1.x.x 版本）

### 查看所有可用版本

访问 [GitHub Packages 页面](https://github.com/hello--world/chatMitm/pkgs/container/chatmitm) 查看所有已发布的版本。

### 认证说明

如果遇到拉取权限问题，需要先登录 GitHub Container Registry：

```bash
# 使用 GitHub Personal Access Token (PAT)
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# 或者交互式登录
docker login ghcr.io
```

创建 PAT 的方法：
1. 访问 GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. 创建新 token，勾选 `read:packages` 权限
3. 使用 token 登录

### 本地构建镜像

如果你想自己构建镜像：

```bash
docker build -t chatmitm:latest .
```

## 参考

- [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy)
- [mitmproxy 文档](https://docs.mitmproxy.org/)

