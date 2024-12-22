package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestHotstar 测试 Hotstar 解锁情况
func TestHotstar(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Hotstar",
	}

	req, err := http.NewRequest("GET", "https://www.hotstar.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept-Language", "en")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查重定向URL
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "/in/") {
		result.Status = "Success"
		result.Region = "IN"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)
	if strings.Contains(htmlContent, "unavailable in your region") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
	} else if strings.Contains(htmlContent, "hotstar.com/in") {
		result.Status = "Success"
		result.Region = "IN"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
