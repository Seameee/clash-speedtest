package unlock

import (
	"io"
	"net/http"
	"regexp"
	"strings"
)

// TestYouTube 测试 YouTube Premium 解锁情况
func TestYouTube(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "YouTube",
	}

	req, err := http.NewRequest("GET", "https://www.youtube.com/premium", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

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

	// 检查是否被阻止访问
	if strings.Contains(htmlContent, "Access to this page has been denied") {
		result.Status = "Failed"
		result.Info = "Access Denied"
		return result
	}

	// 尝试获取地区信息
	regionPattern := `"countryCode":"([^"]+)"`
	re := regexp.MustCompile(regionPattern)
	if matches := re.FindStringSubmatch(htmlContent); len(matches) > 1 {
		result.Status = "Success"
		result.Region = matches[1]
		return result
	}

	if strings.Contains(htmlContent, "Premium is not available") {
		result.Status = "Failed"
		result.Info = "Not Available"
	} else if strings.Contains(htmlContent, "YouTube and YouTube Music ad-free") {
		result.Status = "Success"
		result.Region = "Available"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}

func init() {
	// 注册 YouTube 测试
	streamTests = append(streamTests, TestYouTube)
}
