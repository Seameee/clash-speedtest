package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestKKTV 测试 KKTV 解锁情况
func TestKKTV(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "KKTV",
	}

	req, err := http.NewRequest("GET", "https://api.kktv.me/v3/ipcheck", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept-Language", "zh-TW")

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

	if strings.Contains(htmlContent, `"country":"TW"`) {
		result.Status = "Success"
		result.Region = "TW"
	} else if strings.Contains(htmlContent, "country") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
