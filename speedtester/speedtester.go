package speedtester

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"reporter"

	"github.com/faceair/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/provider"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigPaths      string
	FilterRegex      string
	BlockRegex       string
	ServerURL        string
	DownloadSize     int
	UploadSize       int
	Timeout          time.Duration
	Concurrent       int
	EnableUnlock     bool
	UnlockConcurrent int
	DebugMode        bool
	EnableRisk       bool
	HTMLReport       string
	OutputPath       string
	FastMode         bool
}

type SpeedTester struct {
	config           *Config
	debugMode        bool
	blockedNodes     []string
	blockedNodeCount int
}

func New(config *Config, debugMode bool) *SpeedTester {
	if config.Concurrent <= 0 {
		config.Concurrent = 1
	}
	if config.DownloadSize <= 0 {
		config.DownloadSize = 100 * 1024 * 1024
	}
	if config.UploadSize <= 0 {
		config.UploadSize = 10 * 1024 * 1024
	}
	return &SpeedTester{
		config:    config,
		debugMode: debugMode,
	}
}

type CProxy struct {
	constant.Proxy
	Config map[string]any
}

type RawConfig struct {
	Providers map[string]map[string]any `yaml:"proxy-providers"`
	Proxies   []map[string]any          `yaml:"proxies"`
}

func (st *SpeedTester) LoadProxies() (map[string]*CProxy, error) {
	allProxies := make(map[string]*CProxy)
	st.blockedNodes = make([]string, 0)
	st.blockedNodeCount = 0

	for _, configPath := range strings.Split(st.config.ConfigPaths, ",") {
		var body []byte
		var err error
		if strings.HasPrefix(configPath, "http") {
			var resp *http.Response
			resp, err = http.Get(configPath)
			if err != nil {
				log.Warnln("failed to fetch config: %s", err)
				continue
			}
			body, err = io.ReadAll(resp.Body)
		} else {
			body, err = os.ReadFile(configPath)
		}
		if err != nil {
			log.Warnln("failed to read config: %s", err)
			continue
		}

		rawCfg := &RawConfig{
			Proxies: []map[string]any{},
		}
		if err := yaml.Unmarshal(body, rawCfg); err != nil {
			return nil, err
		}
		proxies := make(map[string]*CProxy)
		proxiesConfig := rawCfg.Proxies
		providersConfig := rawCfg.Providers

		for i, config := range proxiesConfig {
			proxy, err := adapter.ParseProxy(config)
			if err != nil {
				return nil, fmt.Errorf("proxy %d: %w", i, err)
			}

			if _, exist := proxies[proxy.Name()]; exist {
				return nil, fmt.Errorf("proxy %s is the duplicate name", proxy.Name())
			}
			proxies[proxy.Name()] = &CProxy{Proxy: proxy, Config: config}
		}
		for name, config := range providersConfig {
			if name == provider.ReservedName {
				return nil, fmt.Errorf("can not defined a provider called `%s`", provider.ReservedName)
			}
			pd, err := provider.ParseProxyProvider(name, config)
			if err != nil {
				return nil, fmt.Errorf("parse proxy provider %s error: %w", name, err)
			}
			if err := pd.Initial(); err != nil {
				return nil, fmt.Errorf("initial proxy provider %s error: %w", pd.Name(), err)
			}
			for _, proxy := range pd.Proxies() {
				proxies[fmt.Sprintf("[%s] %s", name, proxy.Name())] = &CProxy{
					Proxy:  proxy,
					Config: config,
				}
			}
		}
		for k, p := range proxies {
			switch p.Type() {
			case constant.Shadowsocks, constant.ShadowsocksR, constant.Snell, constant.Socks5, constant.Http,
				constant.Vmess, constant.Vless, constant.Trojan, constant.Hysteria, constant.Hysteria2,
				constant.WireGuard, constant.Tuic, constant.Ssh, constant.AnyTLS:
			default:
				continue
			}
			if _, ok := allProxies[k]; !ok {
				allProxies[k] = p
			}
		}
	}

	filterRegexp := regexp.MustCompile(st.config.FilterRegex)

	// 处理屏蔽关键词
	var blockKeywords []string
	if st.config.BlockRegex != "" {
		// 分割关键词，去除空格
		for _, keyword := range strings.Split(st.config.BlockRegex, "|") {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				blockKeywords = append(blockKeywords, strings.ToLower(keyword))
			}
		}
	}

	// 记录总节点数
	totalNodes := len(allProxies)

	filteredProxies := make(map[string]*CProxy)
	for name, proxy := range allProxies {
		// 检查是否包含屏蔽关键词
		shouldBlock := false
		if len(blockKeywords) > 0 {
			lowerName := strings.ToLower(name)
			for _, keyword := range blockKeywords {
				if strings.Contains(lowerName, keyword) {
					if st.debugMode {
						st.blockedNodes = append(st.blockedNodes, fmt.Sprintf("%s (匹配关键词: %s)", name, keyword))
						st.blockedNodeCount++
					}
					shouldBlock = true
					break
				}
			}
		}

		if shouldBlock {
			continue
		}

		if filterRegexp.MatchString(name) {
			filteredProxies[name] = proxy
		}
	}

	// 在Debug模式下输出屏蔽信息
	if st.debugMode && len(blockKeywords) > 0 {
		fmt.Printf("\n[Debug] 节点统计信息:\n")
		fmt.Printf("[Debug] 总节点数: %d\n", totalNodes)
		fmt.Printf("[Debug] 已屏蔽节点数: %d\n", st.blockedNodeCount)
		fmt.Printf("[Debug] 剩余节点数: %d\n", len(filteredProxies))
		if st.blockedNodeCount > 0 {
			fmt.Printf("\n[Debug] 被屏蔽的节点:\n")
			for _, name := range st.blockedNodes {
				fmt.Printf("[Debug] - %s\n", name)
			}
		}
		fmt.Println()
	}

	return filteredProxies, nil
}

func (st *SpeedTester) TestProxies(proxies map[string]*CProxy, fn func(result *Result)) {
	var htmlReporter *reporter.HTMLReporter
	var err error

	if st.config.HTMLReport != "" {
		htmlReporter, err = reporter.NewHTMLReporter(
			st.config.HTMLReport,
			st.config.EnableUnlock,
			st.config.ConfigPaths,
			len(proxies),
			st.config.OutputPath,
			st.config.FastMode,
		)
		if err != nil {
			log.Errorln("初始化 HTML 报告失败: %v", err)
			return
		}
	}

	for name, proxy := range proxies {
		result := st.testProxy(name, proxy)

		if htmlReporter != nil {
			// 转换结果为 HTML 报告格式
			htmlResult := &reporter.Result{
				ProxyName:    result.ProxyName,
				ProxyType:    result.ProxyType,
				Latency:      result.FormatLatency(),
				LatencyValue: result.Latency.Milliseconds(),
				LastUpdate:   time.Now(),
			}

			// 只在非快速模式下添加其他信息
			if !st.config.FastMode {
				htmlResult.Jitter = result.FormatJitter()
				htmlResult.JitterValue = result.Jitter.Milliseconds()
				htmlResult.PacketLoss = result.FormatPacketLoss()
				htmlResult.PacketLossValue = result.PacketLoss
				htmlResult.Location = reporter.FormatLocation(result.FormatLocation())
				htmlResult.StreamUnlock = result.FormatStreamUnlock()
				htmlResult.UnlockPlatforms = reporter.ParseStreamUnlock(result.FormatStreamUnlock())
				htmlResult.DownloadSpeed = result.FormatDownloadSpeed()
				htmlResult.DownloadSpeedMB = result.DownloadSpeed / (1024 * 1024)
				htmlResult.UploadSpeed = result.FormatUploadSpeed()
				htmlResult.UploadSpeedMB = result.UploadSpeed / (1024 * 1024)
			}

			if err := htmlReporter.AddResult(htmlResult); err != nil {
				log.Errorln("添加 HTML 报告结果失败: %v", err)
			}
		}

		// 回调函数在最后调用，确保 HTML 报告已更新
		fn(result)
	}
}

type testJob struct {
	name  string
	proxy *CProxy
}

type Result struct {
	ProxyName     string         `json:"proxy_name"`
	ProxyType     string         `json:"proxy_type"`
	ProxyConfig   map[string]any `json:"proxy_config"`
	Latency       time.Duration  `json:"latency"`
	Jitter        time.Duration  `json:"jitter"`
	PacketLoss    float64        `json:"packet_loss"`
	DownloadSize  float64        `json:"download_size"`
	DownloadTime  time.Duration  `json:"download_time"`
	DownloadSpeed float64        `json:"download_speed"`
	UploadSize    float64        `json:"upload_size"`
	UploadTime    time.Duration  `json:"upload_time"`
	UploadSpeed   float64        `json:"upload_speed"`
	Location      string         `json:"location"`
	StreamUnlock  string         `json:"stream_unlock"`
}

func (r *Result) FormatDownloadSpeed() string {
	return formatSpeed(r.DownloadSpeed)
}

func (r *Result) FormatLatency() string {
	if r.Latency == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Latency.Milliseconds())
}

func (r *Result) FormatJitter() string {
	if r.Jitter == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Jitter.Milliseconds())
}

func (r *Result) FormatPacketLoss() string {
	return fmt.Sprintf("%.1f%%", r.PacketLoss)
}

func (r *Result) FormatUploadSpeed() string {
	return formatSpeed(r.UploadSpeed)
}

func (r *Result) FormatLocation() string {
	if r.Location == "" {
		return "N/A"
	}
	return r.Location
}

func (r *Result) FormatStreamUnlock() string {
	if r.StreamUnlock == "" {
		return "N/A"
	}
	return r.StreamUnlock
}

func formatSpeed(bytesPerSecond float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s", "TB/s"}
	unit := 0
	speed := bytesPerSecond
	for speed >= 1024 && unit < len(units)-1 {
		speed /= 1024
		unit++
	}
	return fmt.Sprintf("%.2f%s", speed, units[unit])
}

func (st *SpeedTester) testProxy(name string, proxy *CProxy) *Result {
	result := &Result{
		ProxyName:   name,
		ProxyType:   proxy.Type().String(),
		ProxyConfig: proxy.Config,
	}

	// 1. 先进行延迟测试
	latencyResult := st.testLatency(proxy)
	result.Latency = latencyResult.avgLatency

	// 如果是快速模式，只测试延迟，不测试抖动和丢包率
	if !st.config.FastMode {
		result.Jitter = latencyResult.jitter
		result.PacketLoss = latencyResult.packetLoss
	}

	// 如果是快速模式，只测试延迟，直接返回结果
	if st.config.FastMode {
		return result
	}

	// 如果延迟测试失败（延迟为0或丢包率为100%），直接返回结果
	if result.Latency == 0 || result.PacketLoss == 100 {
		return result
	}

	client := st.createClient(proxy)

	// 2. 如果启用了解锁检测，进行地理位置和流媒体检测
	if st.config.EnableUnlock {
		// 先获取地理位置和风险值
		location, err := st.testLocation(client)
		if err == nil {
			result.Location = location
		}

		// 创建一个通道用于流媒体检测结果
		streamChan := make(chan string, 1)

		// 在后台进行流媒体检测
		go func() {
			streamChan <- unlock.TestAll(client, st.config.UnlockConcurrent, st.debugMode)
		}()

		// 如果不需要测速，立即返回结果
		if st.config.EnableUnlock {
			// 等待流媒体检测结果
			result.StreamUnlock = <-streamChan
			return result
		}
	}

	// 3. 如果不是解锁模式，或者需要测试速度，进行下载和上传测试
	if !st.config.EnableUnlock {
		// 并发进行下载和上传测试
		var wg sync.WaitGroup
		downloadResults := make(chan *downloadResult, st.config.Concurrent)

		// 计算每个并发连接的数据大小
		downloadChunkSize := st.config.DownloadSize / st.config.Concurrent
		uploadChunkSize := st.config.UploadSize / st.config.Concurrent

		// 启动下载测试
		for i := 0; i < st.config.Concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				downloadResults <- st.testDownload(proxy, downloadChunkSize)
			}()
		}
		wg.Wait()

		uploadResults := make(chan *downloadResult, st.config.Concurrent)

		// 启动上传测试
		for i := 0; i < st.config.Concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				uploadResults <- st.testUpload(proxy, uploadChunkSize)
			}()
		}
		wg.Wait()

		// 4. 汇总结果
		var totalDownloadBytes, totalUploadBytes int64
		var totalDownloadTime, totalUploadTime time.Duration
		var downloadCount, uploadCount int

		for i := 0; i < st.config.Concurrent; i++ {
			if dr := <-downloadResults; dr != nil {
				totalDownloadBytes += dr.bytes
				totalDownloadTime += dr.duration
				downloadCount++
			}
		}
		close(downloadResults)

		for i := 0; i < st.config.Concurrent; i++ {
			if ur := <-uploadResults; ur != nil {
				totalUploadBytes += ur.bytes
				totalUploadTime += ur.duration
				uploadCount++
			}
		}
		close(uploadResults)

		if downloadCount > 0 {
			result.DownloadSize = float64(totalDownloadBytes)
			result.DownloadTime = totalDownloadTime / time.Duration(downloadCount)
			result.DownloadSpeed = float64(totalDownloadBytes) / result.DownloadTime.Seconds()
		}
		if uploadCount > 0 {
			result.UploadSize = float64(totalUploadBytes)
			result.UploadTime = totalUploadTime / time.Duration(uploadCount)
			result.UploadSpeed = float64(totalUploadBytes) / result.UploadTime.Seconds()
		}
	}

	return result
}

func (st *SpeedTester) testLocation(client *http.Client) (string, error) {
	if st.config.EnableUnlock && st.config.EnableRisk {
		return unlock.GetLocationWithRisk(client, st.debugMode, st.config.EnableRisk)
	}
	return unlock.GetLocation(client, st.debugMode)
}

type latencyResult struct {
	avgLatency time.Duration
	jitter     time.Duration
	packetLoss float64
}

func (st *SpeedTester) testLatency(proxy constant.Proxy) *latencyResult {
	client := st.createClient(proxy)
	latencies := make([]time.Duration, 0, 6)
	failedPings := 0

	for i := 0; i < 6; i++ {
		time.Sleep(100 * time.Millisecond)

		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=0", st.config.ServerURL))
		if err != nil {
			failedPings++
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			latencies = append(latencies, time.Since(start))
		} else {
			failedPings++
		}
	}

	return calculateLatencyStats(latencies, failedPings)
}

type downloadResult struct {
	bytes    int64
	duration time.Duration
}

func (st *SpeedTester) testDownload(proxy constant.Proxy, size int) *downloadResult {
	client := st.createClient(proxy)
	start := time.Now()

	resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=%d", st.config.ServerURL, size))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	downloadBytes, _ := io.Copy(io.Discard, resp.Body)

	return &downloadResult{
		bytes:    downloadBytes,
		duration: time.Since(start),
	}
}

func (st *SpeedTester) testUpload(proxy constant.Proxy, size int) *downloadResult {
	client := st.createClient(proxy)
	reader := NewZeroReader(size)

	start := time.Now()
	resp, err := client.Post(
		fmt.Sprintf("%s/__up", st.config.ServerURL),
		"application/octet-stream",
		reader,
	)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	return &downloadResult{
		bytes:    reader.WrittenBytes(),
		duration: time.Since(start),
	}
}

func (st *SpeedTester) createClient(proxy constant.Proxy) *http.Client {
	return &http.Client{
		Timeout: st.config.Timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				var u16Port uint16
				if port, err := strconv.ParseUint(port, 10, 16); err == nil {
					u16Port = uint16(port)
				}
				return proxy.DialContext(ctx, &constant.Metadata{
					Host:    host,
					DstPort: u16Port,
				})
			},
		},
	}
}

func calculateLatencyStats(latencies []time.Duration, failedPings int) *latencyResult {
	result := &latencyResult{
		packetLoss: float64(failedPings) / 6.0 * 100,
	}

	if len(latencies) == 0 {
		return result
	}

	// 计算平均延迟
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	result.avgLatency = total / time.Duration(len(latencies))

	// 计算抖动
	var variance float64
	for _, l := range latencies {
		diff := float64(l - result.avgLatency)
		variance += diff * diff
	}
	variance /= float64(len(latencies))
	result.jitter = time.Duration(math.Sqrt(variance))

	return result
}

func (st *SpeedTester) testStreamUnlock(proxy *CProxy) (string, error) {
	client := st.createClient(proxy)
	return unlock.TestAll(client, st.config.UnlockConcurrent, st.debugMode), nil
}
