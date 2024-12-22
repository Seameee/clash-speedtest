package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestDisney 测试 Disney+ 解锁情况
func TestDisney(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Disney+",
	}

	req, err := http.NewRequest("GET", "https://www.disneyplus.com", nil)
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

	// 检查重定向
	location := resp.Request.URL.String()
	switch {
	case strings.Contains(location, "/unavailable"):
		result.Status = "Failed"
		result.Info = "Not Available"
		return result
	case strings.Contains(location, "/blocked"):
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	// 检查是否有地区限制信息
	if strings.Contains(htmlContent, "not available in your region") ||
		strings.Contains(htmlContent, "Disney+ is not available in your country") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	// 检查是否显示订阅界面
	if strings.Contains(htmlContent, "subscription") ||
		strings.Contains(htmlContent, "hero-collection") ||
		strings.Contains(htmlContent, "sign-up") {
		result.Status = "Success"
		// 尝试获取地区信息
		if strings.Contains(htmlContent, `"region":"`) {
			start := strings.Index(htmlContent, `"region":"`) + 9
			end := strings.Index(htmlContent[start:], `"`) + start
			if end > start {
				result.Region = htmlContent[start:end]
				return result
			}
		}
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 Disney+ 测试
	streamTests = append(streamTests, TestDisney)
}
