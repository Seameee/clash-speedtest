package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"reporter"

	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/metacubex/mihomo/log"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

var (
	configPathsConfig = flag.String("c", "", "配置文件路径，支持 http(s) 链接")
	filterRegexConfig = flag.String("f", ".+", "使用正则表达式过滤节点名称(例如：-f 'HK|港')")
	blockKeywords     = flag.String("b", "", "使用关键词屏蔽节点，多个关键词用竖线|分隔(例如：-b '倍率|x1|1x|0.5x|试用|体验')")
	serverURL         = flag.String("server-url", "https://speed.cloudflare.com", "测速服务器地址")
	downloadSize      = flag.Int("download-size", 50*1024*1024, "下载测试的数据大小")
	uploadSize        = flag.Int("upload-size", 20*1024*1024, "上传测试的数据大小")
	timeout           = flag.Duration("timeout", time.Second*5, "测试超时时间")
	concurrent        = flag.Int("concurrent", 4, "下载并发数")
	outputPath        = flag.String("output", "", "输出配置文件路径+名称")
	maxLatency        = flag.Duration("max-latency", 0, "(如果没有指定，默认过滤延迟大于0的节点)延迟过滤阈值，单位 ms，大于此值的节点将被过滤，例如 -max-latency 1000ms 表示过滤延迟大于 1000 ms 的节点")
	minSpeed          = flag.Float64("min-speed", 0, "(如果没有指定，默认过滤延迟大于0的节点)速度过滤阈值，单位 MB/s，小于此值的节点将被过滤，例如 -min-speed 10 表示过滤速度小于 10 MB/s 的节点")
	enableUnlock      = flag.Bool("unlock", false, "启用流媒体解锁检测(启用OUTPUT时，默认只保存延迟大于0的节点)")
	unlockConcurrent  = flag.Int("unlock-concurrent", 5, "解锁测试并发数，默认 5 (仅在-unlock模式下有效)")
	debugMode         = flag.Bool("debug", false, "启用调试模式，可用于查看节点屏蔽信息或解锁测试详情")
	enableRisk        = flag.Bool("risk", false, "启用解锁测试时的 IP 风险检测(仅在-unlock模式下有效)")
	htmlReport        = flag.String("html", "", "输出 HTML 报告的路径+名称(默认5秒自动刷新，支持手动刷新)")
	fastMode          = flag.Bool("fast", false, "快速测试模式，仅测试节点延迟")
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	Version     = "1.6.4"
)

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	fmt.Printf("Clash Speedtest Or Check Media Unlock %s\n\n", Version)

	if *configPathsConfig == "" {
		log.Fatalln("please specify the configuration file")
	}

	if *debugMode && !*enableUnlock && *blockKeywords == "" {
		log.Fatalln("debug mode can only be used with unlock testing or node blocking enabled")
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      *configPathsConfig,
		FilterRegex:      *filterRegexConfig,
		BlockRegex:       *blockKeywords,
		ServerURL:        *serverURL,
		DownloadSize:     *downloadSize,
		UploadSize:       *uploadSize,
		Timeout:          *timeout,
		Concurrent:       *concurrent,
		EnableUnlock:     *enableUnlock,
		UnlockConcurrent: *unlockConcurrent,
		DebugMode:        *debugMode,
		EnableRisk:       *enableRisk,
		HTMLReport:       *htmlReport,
		OutputPath:       *outputPath,
		FastMode:         *fastMode,
	}, *debugMode)

	if *debugMode {
		fmt.Println("Debug 模式已启用")
	}

	allProxies, err := speedTester.LoadProxies()
	if err != nil {
		log.Fatalln("load proxies failed: %v", err)
	}

	bar := progressbar.Default(int64(len(allProxies)), "测试中...")
	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		bar.Add(1)
		bar.Describe(result.ProxyName)
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

	if *htmlReport != "" {
		quit := make(chan struct{})
		mux := http.NewServeMux()
		mux.HandleFunc("/convert", reporter.HandleConverter)
		mux.HandleFunc("/readfile", reporter.HandleReadFile)

		server := &http.Server{
			Addr:    "127.0.0.1:8080",
			Handler: mux,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Errorln("HTTP server error: %v", err)
			}
		}()

		fmt.Printf("\n配置转换服务已启动 [127.0.0.1 端口: 8080]\n")
		fmt.Printf("按 Enter 键或 Ctrl+C 退出程序...\n")

		go func() {
			fmt.Scanln()
			close(quit)
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case <-quit:
			fmt.Println("\n收到退出信号，正在关闭服务器...")
		case <-sigChan:
			fmt.Println("\n收到中断信号，正在关闭服务器...")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Errorln("服务器关闭出错: %v", err)
		} else {
			fmt.Println("服务器已关闭，端口已释放")
		}
	}
}

func printResults(results []*speedtester.Result, enableUnlock bool) {
	table := tablewriter.NewWriter(os.Stdout)

	// 处理流媒体结果的换行
	formatStreamUnlock := func(unlock string) string {
		if unlock == "N/A" {
			return unlock
		}
		// 每4个台换一行
		parts := strings.Split(unlock, ", ")
		var lines []string
		for i := 0; i < len(parts); i += 4 {
			end := i + 4
			if end > len(parts) {
				end = len(parts)
			}
			lineItems := parts[i:end]
			// 为每个平台添加色
			for j := range lineItems {
				lineItems[j] = colorGreen + lineItems[j] + colorReset
			}
			lines = append(lines, strings.Join(lineItems, ", "))
		}
		return strings.Join(lines, "\n")
	}

	var headers []string
	if *fastMode {
		headers = []string{
			"序号",
			"节点名称",
			"类型",
			"延迟",
		}
	} else if enableUnlock {
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
	table.SetAutoWrapText(false)
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
	if *fastMode {
		table.SetColMinWidth(0, 4)  // 序号
		table.SetColMinWidth(1, 20) // 节点名称
		table.SetColMinWidth(2, 8)  // 类型
		table.SetColMinWidth(3, 8)  // 延迟
	} else if enableUnlock {
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
		if *fastMode {
			row = []string{
				idStr,
				proxyNameStr,
				proxyTypeStr,
				latencyStr,
			}
		} else if enableUnlock {
			// 如果是解锁测试模式且延迟为0或丢包率为100%，则跳过后续测试
			if result.Latency == 0 || result.PacketLoss == 100 {
				row = []string{
					idStr,
					proxyNameStr,
					proxyTypeStr,
					latencyStr,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
				}
			} else {
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

				// 地理位置颜色
				locationStr := result.FormatLocation()
				if locationStr != "N/A" {
					locationStr = colorGreen + locationStr + colorReset
				} else {
					locationStr = colorRed + locationStr + colorReset
				}

				// 流媒体解锁颜色
				streamUnlock := result.FormatStreamUnlock()
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
			}
		} else {
			// 如果延迟为0或丢包率为100%，则跳过后续测试
			if result.Latency == 0 || result.PacketLoss == 100 {
				row = []string{
					idStr,
					proxyNameStr,
					proxyTypeStr,
					latencyStr,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
					colorRed + "N/A" + colorReset,
				}
			} else {
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
		// 检查延迟是否大于0
		if result.Latency <= 0 {
			continue
		}

		if *enableUnlock {
			// 解锁模式：只要延迟大于0就保存
			filteredResults = append(filteredResults, result)
			continue
		}

		// 检查延迟条件
		if *maxLatency > 0 && result.Latency >= *maxLatency {
			continue
		}

		// 检查速度条件
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
