package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/metacubex/mihomo/log"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

var (
	configPathsConfig = flag.String("c", "", "config file path, also support http(s) url")
	filterRegexConfig = flag.String("f", ".+", "filter proxies by name, use regexp")
	serverURL         = flag.String("server-url", "https://speed.cloudflare.com", "server url")
	downloadSize      = flag.Int("download-size", 50*1024*1024, "download size for testing proxies")
	uploadSize        = flag.Int("upload-size", 20*1024*1024, "upload size for testing proxies")
	timeout           = flag.Duration("timeout", time.Second*5, "timeout for testing proxies")
	concurrent        = flag.Int("concurrent", 4, "download concurrent size")
	outputPath        = flag.String("output", "", "output config file path")
	maxLatency        = flag.Duration("max-latency", 800*time.Millisecond, "filter latency greater than this value")
	minSpeed          = flag.Float64("min-speed", 5, "filter speed less than this value(unit: MB/s)")
	enableUnlock      = flag.Bool("unlock", false, "enable streaming media unlock detection")
	unlockConcurrent  = flag.Int("unlock-concurrent", 5, "concurrent size for unlock testing")
	debugMode         = flag.Bool("debug", false, "enable debug mode for unlock testing")
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	Version     = "1.6.1"
)

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	fmt.Printf("Clash Speedtest Or Check Media Unlock %s\n\n", Version)

	if *configPathsConfig == "" {
		log.Fatalln("please specify the configuration file")
	}

	if *debugMode && !*enableUnlock {
		log.Fatalln("debug mode can only be used with unlock testing enabled")
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      *configPathsConfig,
		FilterRegex:      *filterRegexConfig,
		ServerURL:        *serverURL,
		DownloadSize:     *downloadSize,
		UploadSize:       *uploadSize,
		Timeout:          *timeout,
		Concurrent:       *concurrent,
		EnableUnlock:     *enableUnlock,
		UnlockConcurrent: *unlockConcurrent,
		DebugMode:        *debugMode,
	})

	allProxies, err := speedTester.LoadProxies()
	if err != nil {
		log.Fatalln("load proxies failed: %v", err)
	}

	bar := progressbar.Default(int64(len(allProxies)), "测试中...")
	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		bar.Add(1)
		bar.Describe(result.ProxyName)
		results = append(results, result)
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].DownloadSpeed > results[j].DownloadSpeed
	})

	printResults(results, *enableUnlock)

	if *outputPath != "" {
		err = saveConfig(results)
		if err != nil {
			log.Fatalln("save config file failed: %v", err)
		}
		fmt.Printf("\nsave config file to: %s\n", *outputPath)
	}
}

func printResults(results []*speedtester.Result, enableUnlock bool) {
	table := tablewriter.NewWriter(os.Stdout)

	var headers []string
	if enableUnlock {
		headers = []string{
			"序号",
			"节点名称",
			"类型",
			"延迟",
			"抖动",
			"丢包率",
			"地理",
			"流媒体",
		}
	} else {
		headers = []string{
			"序号",
			"节点名称",
			"类型",
			"延迟",
			"抖动",
			"丢包率",
			"下载速度",
			"上传速度",
		}
	}
	table.SetHeader(headers)

	// 设置表格样式
	table.SetAutoWrapText(false) // 默认关闭自动换行
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// 设置列宽度
	if enableUnlock {
		table.SetColMinWidth(0, 4)  // 序号
		table.SetColMinWidth(1, 20) // 节点名称
		table.SetColMinWidth(2, 8)  // 类型
		table.SetColMinWidth(3, 8)  // 延迟
		table.SetColMinWidth(4, 8)  // 抖动
		table.SetColMinWidth(5, 8)  // 丢包率
		table.SetColMinWidth(6, 10) // 地理
		table.SetColMinWidth(7, 40) // 流媒体
	} else {
		table.SetColMinWidth(0, 4)  // 序号
		table.SetColMinWidth(1, 20) // 节点名称
		table.SetColMinWidth(2, 8)  // 类型
		table.SetColMinWidth(3, 8)  // 延迟
		table.SetColMinWidth(4, 8)  // 抖动
		table.SetColMinWidth(5, 8)  // 丢包率
		table.SetColMinWidth(6, 12) // 下载速度
		table.SetColMinWidth(7, 12) // 上传速度
	}

	// 处理流媒体结果的换行
	formatStreamUnlock := func(unlock string) string {
		if unlock == "N/A" {
			return unlock
		}
		// 每4个��台换一行
		parts := strings.Split(unlock, ", ")
		var lines []string
		for i := 0; i < len(parts); i += 4 {
			end := i + 4
			if end > len(parts) {
				end = len(parts)
			}
			lineItems := parts[i:end]
			// 为每个平台添加颜色
			for j := range lineItems {
				lineItems[j] = colorGreen + lineItems[j] + colorReset
			}
			lines = append(lines, strings.Join(lineItems, ", "))
		}
		return strings.Join(lines, "\n")
	}

	for i, result := range results {
		idStr := fmt.Sprintf("%d.", i+1)

		// 延迟颜色
		latencyStr := result.FormatLatency()
		if result.Latency > 0 {
			if result.Latency < 800*time.Millisecond {
				latencyStr = colorGreen + latencyStr + colorReset
			} else if result.Latency < 1500*time.Millisecond {
				latencyStr = colorYellow + latencyStr + colorReset
			} else {
				latencyStr = colorRed + latencyStr + colorReset
			}
		} else {
			latencyStr = colorRed + latencyStr + colorReset
		}

		// 如果是解锁测试模式且延迟为0或丢包率为100%，则跳过后续测试
		if enableUnlock && (result.Latency == 0 || result.PacketLoss == 100) {
			row := []string{
				idStr,
				colorRed + result.ProxyName + colorReset,
				colorRed + result.ProxyType + colorReset,
				latencyStr,
				colorRed + "N/A" + colorReset,
				colorRed + "N/A" + colorReset,
				colorRed + "N/A" + colorReset,
				colorRed + "N/A" + colorReset,
			}
			table.Append(row)
			continue
		}

		// 抖动颜色
		jitterStr := result.FormatJitter()
		if result.Jitter > 0 {
			if result.Jitter < 800*time.Millisecond {
				jitterStr = colorGreen + jitterStr + colorReset
			} else if result.Jitter < 1500*time.Millisecond {
				jitterStr = colorYellow + jitterStr + colorReset
			} else {
				jitterStr = colorRed + jitterStr + colorReset
			}
		} else {
			jitterStr = colorRed + jitterStr + colorReset
		}

		// 丢包率颜色
		packetLossStr := result.FormatPacketLoss()
		if result.PacketLoss < 10 {
			packetLossStr = colorGreen + packetLossStr + colorReset
		} else if result.PacketLoss < 20 {
			packetLossStr = colorYellow + packetLossStr + colorReset
		} else {
			packetLossStr = colorRed + packetLossStr + colorReset
		}

		// 节点名称和类型颜色
		proxyNameStr := result.ProxyName
		proxyTypeStr := result.ProxyType
		if result.Latency > 0 && result.PacketLoss < 100 {
			proxyNameStr = colorGreen + proxyNameStr + colorReset
			proxyTypeStr = colorGreen + proxyTypeStr + colorReset
		} else {
			proxyNameStr = colorRed + proxyNameStr + colorReset
			proxyTypeStr = colorRed + proxyTypeStr + colorReset
		}

		var row []string
		if enableUnlock {
			// 地理位置颜色
			locationStr := result.FormatLocation()
			if locationStr != "N/A" {
				locationStr = colorGreen + locationStr + colorReset
			} else {
				locationStr = colorRed + locationStr + colorReset
			}

			// 流媒体解锁颜色
			streamUnlock := result.FormatStreamUnlock()
			if *debugMode {
				fmt.Printf("节点 %s 的流媒体结果: %s\n", result.ProxyName, streamUnlock)
			}
			var unlockStr string
			if streamUnlock != "N/A" {
				unlockStr = formatStreamUnlock(streamUnlock)
			} else {
				unlockStr = colorRed + "N/A" + colorReset
			}

			row = []string{
				idStr,
				proxyNameStr,
				proxyTypeStr,
				latencyStr,
				jitterStr,
				packetLossStr,
				locationStr,
				unlockStr,
			}
		} else {
			// 下载速度颜色
			downloadSpeed := result.DownloadSpeed / (1024 * 1024)
			downloadSpeedStr := result.FormatDownloadSpeed()
			if downloadSpeed >= 10 {
				downloadSpeedStr = colorGreen + downloadSpeedStr + colorReset
			} else if downloadSpeed >= 5 {
				downloadSpeedStr = colorYellow + downloadSpeedStr + colorReset
			} else {
				downloadSpeedStr = colorRed + downloadSpeedStr + colorReset
			}

			// 上传速度颜色
			uploadSpeed := result.UploadSpeed / (1024 * 1024)
			uploadSpeedStr := result.FormatUploadSpeed()
			if uploadSpeed >= 5 {
				uploadSpeedStr = colorGreen + uploadSpeedStr + colorReset
			} else if uploadSpeed >= 2 {
				uploadSpeedStr = colorYellow + uploadSpeedStr + colorReset
			} else {
				uploadSpeedStr = colorRed + uploadSpeedStr + colorReset
			}

			row = []string{
				idStr,
				proxyNameStr,
				proxyTypeStr,
				latencyStr,
				jitterStr,
				packetLossStr,
				downloadSpeedStr,
				uploadSpeedStr,
			}
		}

		table.Append(row)
	}

	fmt.Println()
	table.Render()
	fmt.Println()
}

func saveConfig(results []*speedtester.Result) error {
	filteredResults := make([]*speedtester.Result, 0)
	for _, result := range results {
		if *maxLatency > 0 && result.Latency > *maxLatency {
			continue
		}
		if *minSpeed > 0 && float64(result.DownloadSpeed)/(1024*1024) < *minSpeed {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	proxies := make([]map[string]any, 0)
	for _, result := range filteredResults {
		proxies = append(proxies, result.ProxyConfig)
	}

	config := &speedtester.RawConfig{
		Proxies: proxies,
	}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(*outputPath, yamlData, 0o644)
}
