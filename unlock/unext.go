package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestUNEXT 测试 U-NEXT 解锁情况
func TestUNEXT(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "U-NEXT",
	}

	req, err := http.NewRequest("GET", "https://video.unext.jp/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept-Language", "ja-JP")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查重定向URL
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "restrict") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	if strings.Contains(htmlContent, "access from your country") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
	} else if strings.Contains(htmlContent, "u-next") && !strings.Contains(htmlContent, "not available") {
		result.Status = "Success"
		result.Region = "JP"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
