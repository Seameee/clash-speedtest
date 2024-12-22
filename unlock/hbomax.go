package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestHBOMax 测试 HBO Max 解锁情况
func TestHBOMax(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "HBO Max",
	}

	req, err := http.NewRequest("GET", "https://www.max.com/", nil)
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
	if strings.Contains(location, "/geo-availability") {
		result.Status = "Failed"
		result.Info = "Not Available"
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
	if strings.Contains(htmlContent, "currently not available in your region") ||
		strings.Contains(htmlContent, "HBO Max is not available in your territory") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	// 检查是否显示订阅界面
	if strings.Contains(htmlContent, "subscription") ||
		strings.Contains(htmlContent, "sign-up") ||
		strings.Contains(htmlContent, "choose-plan") {
		result.Status = "Success"
		// 尝试获取地区信息
		if strings.Contains(htmlContent, `"territory":"`) {
			start := strings.Index(htmlContent, `"territory":"`) + 12
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
	// 注册 HBO Max 测试
	streamTests = append(streamTests, TestHBOMax)
}
