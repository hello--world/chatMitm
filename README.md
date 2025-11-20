# ChatMitm - 流式接口拦截工具

基于 [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy) 实现的通用流式接口数据拦截工具，可以拦截并保存任意流式接口（SSE、流式 JSON 等）的响应数据。

## 功能特性

- ✅ 自动检测并拦截流式响应（Server-Sent Events、流式 JSON 等）
- ✅ 自动保存流式数据到本地文件
- ✅ 支持所有流式接口，无需配置
- ✅ 提供 Web 界面查看请求详情
- ✅ 不修改原始数据流，只进行监控和保存

## 安装

### 前置要求

- Go 1.21 或更高版本

### 安装依赖

```bash
go mod download
```

## 使用方法

### 1. 启动代理服务器

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

## 参考

- [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy)
- [mitmproxy 文档](https://docs.mitmproxy.org/)

