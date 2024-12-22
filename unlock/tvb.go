package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestTVB 测试 TVB 解锁情况
func TestTVB(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "TVB",
	}

	req, err := http.NewRequest("GET", "https://www.mytvsuper.com/iptest.php", nil)
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

	// 检查是否有地区限制信息
	if strings.Contains(htmlContent, "HK") {
		result.Status = "Success"
		result.Region = "HK"
		return result
	}

	// 检查是否被封锁
	if strings.Contains(htmlContent, "blocked") {
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 TVB 测试
	streamTests = append(streamTests, TestTVB)
}
