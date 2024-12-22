package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestPeacock 测试 Peacock TV 解锁情况
func TestPeacock(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Peacock",
	}

	req, err := http.NewRequest("GET", "https://www.peacocktv.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查重定向URL
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "unavailable") {
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

	if strings.Contains(htmlContent, "unavailable in your location") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
	} else if strings.Contains(htmlContent, "choose-plan") || strings.Contains(htmlContent, "watch-online") {
		result.Status = "Success"
		result.Region = "US"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
