package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"github.com/lqqyt2423/go-mitmproxy/web"
	logrus "github.com/sirupsen/logrus"
)

// countingWriter 用于统计行数的 Writer
type countingWriter struct {
	w         io.Writer
	lineCount *int
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	// 统计换行符
	for _, b := range p {
		if b == '\n' {
			(*cw.lineCount)++
		}
	}
	return cw.w.Write(p)
}

// StreamInterceptor 流式响应拦截器
type StreamInterceptor struct {
	outputDir string
}

// NewStreamInterceptor 创建新的流式拦截器
func NewStreamInterceptor(outputDir string) *StreamInterceptor {
	if outputDir == "" {
		outputDir = "./stream_data"
	}
	// 确保输出目录存在
	os.MkdirAll(outputDir, 0755)
	return &StreamInterceptor{
		outputDir: outputDir,
	}
}

// ClientConnected 客户端连接
func (s *StreamInterceptor) ClientConnected(*proxy.ClientConn) {}

// ClientDisconnected 客户端断开连接
func (s *StreamInterceptor) ClientDisconnected(*proxy.ClientConn) {}

// ServerConnected 服务器连接
func (s *StreamInterceptor) ServerConnected(*proxy.ConnContext) {}

// ServerDisconnected 服务器断开连接
func (s *StreamInterceptor) ServerDisconnected(*proxy.ConnContext) {}

// TlsEstablishedServer TLS 握手完成
func (s *StreamInterceptor) TlsEstablishedServer(*proxy.ConnContext) {}

// AccessProxyServer 访问代理服务器（Addon 接口要求）
func (s *StreamInterceptor) AccessProxyServer(*http.Request, http.ResponseWriter) {}

// Requestheaders 请求头
func (s *StreamInterceptor) Requestheaders(*proxy.Flow) {}

// Request 完整请求
func (s *StreamInterceptor) Request(*proxy.Flow) {}

// Responseheaders 响应头
func (s *StreamInterceptor) Responseheaders(flow *proxy.Flow) {
	// 检查是否为流式响应
	contentType := flow.Response.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") ||
		strings.Contains(contentType, "application/stream+json") ||
		strings.Contains(contentType, "application/x-ndjson") {
		// 告诉 go-mitmproxy 进入流式模式，不要缓冲响应体
		flow.Stream = true
		flow.Response.Body = nil

		log.Printf("[流式响应检测] %s %s - Content-Type: %s",
			flow.Request.Method,
			flow.Request.URL.String(),
			contentType)
	}
}

// Response 完整响应
func (s *StreamInterceptor) Response(flow *proxy.Flow) {
	// 对于流式响应，数据已经在 StreamResponseModifier 中处理
}

// StreamRequestModifier 流式请求修改器（这里不需要修改，只返回原 reader）
func (s *StreamInterceptor) StreamRequestModifier(flow *proxy.Flow, reader io.Reader) io.Reader {
	return reader
}

// StreamResponseModifier 流式响应修改器 - 核心拦截逻辑
func (s *StreamInterceptor) StreamResponseModifier(flow *proxy.Flow, reader io.Reader) io.Reader {
	// 首先检查 reader 是否为 nil
	if reader == nil {
		log.Printf("[警告] StreamResponseModifier 收到 nil reader，返回空 reader")
		// 返回一个空的 reader 而不是 nil
		return bytes.NewReader(nil)
	}

	// 检查是否为流式响应
	contentType := flow.Response.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") &&
		!strings.Contains(contentType, "application/stream+json") &&
		!strings.Contains(contentType, "application/x-ndjson") {
		// 不是流式响应，直接返回原 reader
		return reader
	}

	// 生成文件名（基于时间戳和 URL）
	timestamp := time.Now().Format("20060102_150405")
	urlPath := strings.ReplaceAll(flow.Request.URL.Path, "/", "_")
	if urlPath == "" || urlPath == "_" {
		urlPath = "root"
	}
	// 限制文件名长度
	if len(urlPath) > 50 {
		urlPath = urlPath[:50]
	}
	filename := filepath.Join(s.outputDir, timestamp+"_"+urlPath+".txt")

	// 创建文件用于保存数据
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("[错误] 无法创建文件 %s: %v", filename, err)
		// 文件创建失败，直接返回原 reader
		return reader
	}

	log.Printf("[开始拦截] %s %s - 保存流式数据到: %s",
		flow.Request.Method,
		flow.Request.URL.String(),
		filename)

	// 创建管道用于传递给下游
	pr, pw := io.Pipe()

	// 启动 goroutine 来复制数据：从 reader 读取，同时写入文件和管道
	go func() {
		defer func() {
			file.Close()
			pw.Close()
			// 捕获 panic，避免程序崩溃
			if r := recover(); r != nil {
				log.Printf("[错误] 流式数据处理时发生 panic: %v", r)
				pw.CloseWithError(fmt.Errorf("panic: %v", r))
			}
		}()

		// 检查 reader 是否为 nil
		if reader == nil {
			log.Printf("[错误] reader 为 nil")
			pw.CloseWithError(io.ErrUnexpectedEOF)
			return
		}

		// 使用 MultiWriter 同时写入文件和管道
		multiWriter := io.MultiWriter(file, pw)

		// 使用 io.Copy 来复制数据，这样更安全可靠
		// 同时创建一个自定义 Writer 来统计行数
		var totalBytes int64
		lineCount := 0

		// 创建一个带统计功能的 Writer
		countingWriter := &countingWriter{
			w:         multiWriter,
			lineCount: &lineCount,
		}

		totalBytes, err := io.Copy(countingWriter, reader)
		if err != nil && err != io.EOF {
			log.Printf("[错误] 复制流式数据时出错: %v", err)
			pw.CloseWithError(err)
			return
		}

		log.Printf("[完成拦截] %s %s - 已保存 %d 字节 (约 %d 行) 到: %s",
			flow.Request.Method,
			flow.Request.URL.String(),
			totalBytes,
			lineCount,
			filename)
	}()

	return pr
}

func main() {
	// 配置日志级别，减少噪音
	// 设置为 InfoLevel 可以过滤掉大部分正常的连接关闭、EOF、context canceled 等错误
	// 如果需要更详细的日志，可以改为 logrus.DebugLevel
	// 如果只想看严重错误，可以改为 logrus.WarnLevel
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})

	// 配置代理选项
	opts := &proxy.Options{
		Addr:              ":9080",
		StreamLargeBodies: 1024 * 1024 * 10, // 10MB，足够处理流式数据
	}

	// 创建代理实例
	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal("创建代理失败:", err)
	}

	// 创建流式拦截器
	interceptor := NewStreamInterceptor("./stream_data")

	// 添加拦截器插件
	p.AddAddon(interceptor)

	// 添加 Web 界面插件（启用 Web UI）
	p.AddAddon(web.NewWebAddon(":9081"))

	log.Println("=========================================")
	log.Println("流式接口拦截代理已启动")
	log.Println("代理地址: http://127.0.0.1:9080")
	log.Println("Web 界面: http://127.0.0.1:9081")
	log.Println("流式数据保存目录: ./stream_data")
	log.Println("=========================================")
	log.Println("请配置你的客户端使用代理: http://127.0.0.1:9080")
	log.Println("首次使用需要安装证书: ~/.mitmproxy/mitmproxy-ca-cert.pem")
	log.Println("=========================================")

	// 启动代理服务器
	if err := p.Start(); err != nil {
		log.Fatal("启动代理失败:", err)
	}
}
