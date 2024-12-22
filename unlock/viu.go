package unlock

import (
	"net/http"
	"strings"
)

// TestViu 测试 Viu 解锁情况
func TestViu(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Viu",
	}

	req, err := http.NewRequest("GET", "https://www.viu.com", nil)
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

	// 检查重定向地址
	if location := resp.Header.Get("location"); location != "" {
		parts := strings.Split(location, "/")
		if len(parts) >= 5 {
			region := parts[4]
			if region == "no-service" {
				result.Status = "Failed"
				result.Info = "Region Restricted"
				return result
			}
			result.Status = "Success"
			result.Region = strings.ToUpper(region)
			return result
		}
	}

	result.Status = "Failed"
	result.Info = "Region Not Found"
	return result
}
