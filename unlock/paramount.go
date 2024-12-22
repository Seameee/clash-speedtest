package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestParamount 测试 Paramount+ 解锁情况
func TestParamount(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Paramount+",
	}

	req, err := http.NewRequest("GET", "https://www.paramountplus.com/", nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	switch {
	case strings.Contains(htmlContent, "geo-availability"):
		result.Status = "Failed"
		result.Info = "Region Not Available"
	case strings.Contains(htmlContent, "paramount-plus-is-here"):
		result.Status = "Success"
		result.Region = "US"
	case strings.Contains(htmlContent, "choose-plan"):
		result.Status = "Success"
		result.Region = "US"
	default:
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
