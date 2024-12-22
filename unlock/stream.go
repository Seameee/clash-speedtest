package unlock

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const (
	UA_Browser = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// StreamTest 定义流媒体测试函数类型
type StreamTest func(*http.Client) *StreamResult

// 全局变量
var (
	streamTests   []StreamTest
	platformNames = []string{
		"Steam", "Netflix", "Disney+", "YouTube", "YouTube CDN", "ChatGPT", "Google Gemini", "Meta AI", "Abema",
		"Bahamut", "Bilibili China Mainland Only", "Bilibili HongKong/Macau/Taiwan", "Bilibili Taiwan Only",
		"DAZN", "Discovery+", "DMM", "HBO Go Asia", "HBO Max", "Hotstar", "Hulu", "KKTV",
		"LINE TV", "Paramount+", "Peacock", "Prime Video", "Spotify",
		"TVB", "TVer", "U-NEXT", "GooglePlayStore",
		// 新增平台
		"4GTV", "Catchplay+", "encoreTVB", "ESPN+", "Funimation",
		"GYAO", "HamiVideo", "Paravi", "Radiko", "Telasa",
		"VideoMarket", "Viu",
	}
)

// StreamResult 表示流媒体检测结果
type StreamResult struct {
	Platform string // 平台名称
	Status   string // 状态：Success/Failed
	Region   string // 地区/货币代码
	Info     string // 额外信息
}

// TestSteam 测试 Steam 商店货币区域
func TestSteam(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Steam",
	}

	req, err := http.NewRequest("GET", "https://store.steampowered.com/app/761830", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	// 尝试多种方式匹配货币信息
	patterns := []string{
		`"priceCurrency":"([^"]+)"`,
		`data-price-final[^>]+>([A-Z]{2,3})\s`,
		`\$([A-Z]{2,3})\s+\d+\.\d+`,
		`¥\s*\d+`,    // 日元
		`₩\s*\d+`,    // 韩元
		`NT\$\s*\d+`, // 新台币
		`HK\$\s*\d+`, // 港币
		`S\$\s*\d+`,  // 新加坡元
		`A\$\s*\d+`,  // 澳元
		`₹\s*\d+`,    // 印度卢比
		`€\s*\d+`,    // 欧元
		`£\s*\d+`,    // 英镑
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(htmlContent); len(matches) > 0 {
			result.Status = "Success"
			switch {
			case strings.Contains(matches[0], "¥"):
				result.Region = "JPY"
			case strings.Contains(matches[0], "₩"):
				result.Region = "KRW"
			case strings.Contains(matches[0], "NT$"):
				result.Region = "TWD"
			case strings.Contains(matches[0], "HK$"):
				result.Region = "HKD"
			case strings.Contains(matches[0], "S$"):
				result.Region = "SGD"
			case strings.Contains(matches[0], "A$"):
				result.Region = "AUD"
			case strings.Contains(matches[0], "₹"):
				result.Region = "INR"
			case strings.Contains(matches[0], "€"):
				result.Region = "EUR"
			case strings.Contains(matches[0], "£"):
				result.Region = "GBP"
			default:
				if len(matches) > 1 {
					result.Region = matches[1]
				} else {
					result.Region = matches[0]
				}
			}
			return result
		}
	}

	// 检查是否被重定向到年龄验证页面
	if strings.Contains(htmlContent, "agecheck") || strings.Contains(htmlContent, "age_check") {
		result.Status = "Failed"
		result.Info = "Age Check Required"
		return result
	}

	// 检查是否在维护
	if strings.Contains(htmlContent, "maintenance") {
		result.Status = "Failed"
		result.Info = "Store Maintenance"
		return result
	}

	result.Status = "Failed"
	result.Info = "Currency Not Found"
	return result
}

// FormatResult 格式化检测结果为字符串
func (r *StreamResult) FormatResult() string {
	if r.Status == "Success" {
		if r.Info != "" {
			return r.Region + " (" + r.Info + ")"
		}
		return r.Region
	}
	if r.Info != "" {
		return "Failed (" + r.Info + ")"
	}
	return "Failed"
}

// TestAll 并发测试所有流媒体平台
func TestAll(client *http.Client, concurrency int, debug bool) string {
	if concurrency <= 0 {
		concurrency = 5 // 默认并发数
	}

	// 使用 map 记录已注册的平台
	registeredPlatforms := make(map[string]bool)
	uniqueTests := make([]StreamTest, 0, len(streamTests))

	// 使用全局的 platformNames
	for _, name := range platformNames {
		registeredPlatforms[name] = true
	}

	// 只保留一个实例
	for _, test := range streamTests {
		if result := test(client); result != nil {
			if registeredPlatforms[result.Platform] {
				uniqueTests = append(uniqueTests, test)
				registeredPlatforms[result.Platform] = false // 标记为已添加
			}
		}
	}

	if debug {
		fmt.Printf("\n开始流媒体并发检测，并发数: %d，总平台数: %d\n", concurrency, len(uniqueTests))
	}

	resultChan := make(chan *StreamResult, len(uniqueTests))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency) // 用于控制并发数的信号量

	// 启动所有测试
	for i := range uniqueTests {
		wg.Add(1)
		test := uniqueTests[i] // 在循环内创建局部变量，避免闭包问题
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 执行测试
			result := test(client)
			if result != nil {
				if debug {
					fmt.Printf("检测结果: %s - 状态: %s, 区域: %s, 信息: %s\n",
						result.Platform, result.Status, result.Region, result.Info)
				}
				resultChan <- result
			}
		}()
	}

	// 等待所有测试完成
	go func() {
		wg.Wait()
		close(resultChan)
		if debug {
			fmt.Printf("所有流媒体检测完成\n\n")
		}
	}()

	// 收集并处理结果
	var successResults []string
	resultMap := make(map[string]bool) // 用于去重

	for result := range resultChan {
		if result.Status == "Success" && !resultMap[result.Platform] {
			resultMap[result.Platform] = true
			formatted := ""
			if result.Region != "" && result.Region != "Available" {
				formatted = fmt.Sprintf("%s:%s", result.Platform, result.Region)
			} else {
				formatted = result.Platform
			}
			successResults = append(successResults, formatted)
		}
	}

	// 按字母顺序排序结果
	sort.Strings(successResults)

	// 返回所有成功的结果
	if len(successResults) > 0 {
		return strings.Join(successResults, ", ")
	}
	return "N/A"
}

func init() {
	// 注册所有流媒体测试
	streamTests = []StreamTest{
		TestSteam,
		TestNetflix,
		TestDisney,
		TestYouTube,
		TestYouTubeCDN,
		TestOpenAI,
		TestGemini,
		TestMetaAI,
		TestAbema,
		TestBahamut,
		TestBilibiliMainland,
		TestBilibiliHKMCTW,
		TestBilibiliTW,
		TestDAZN,
		TestDiscovery,
		TestDMM,
		TestHBOGoAsia,
		TestHBOMax,
		TestHotstar,
		TestHulu,
		TestKKTV,
		TestLineTV,
		TestParamount,
		TestPeacock,
		TestPrimeVideo,
		TestSpotify,
		TestTVB,
		TestTVer,
		TestUNEXT,
		TestGooglePlayStore,
		Test4GTV,
		TestParavi,
		TestRadiko,
		TestCatchplay,
		TestEncoreTVB,
		TestESPN,
		TestFunimation,
		TestGYAO,
		TestHamiVideo,
		TestTelasa,
		TestVideoMarket,
		TestViu,
	}
}
