package unlock

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// TestDiscovery 测试 Discovery+ 解锁情况
func TestDiscovery(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Discovery+",
	}

	req, err := http.NewRequest("GET", "https://us1-prod-direct.discoveryplus.com/token?deviceId=d1a4a5d25212400f1b6cd3ee39f616cf&realm=go&shortlived=true", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("X-Request-Id", "b4ab7368-9614-4bab-9c82-3c4b7474a97d")

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

	var response struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	switch {
	case strings.Contains(response.Code, "geo_blocked"):
		result.Status = "Failed"
		result.Info = "Region Not Available"
	case strings.Contains(response.Message, "client not authorized"):
		result.Status = "Success"
		result.Region = "US"
	case strings.Contains(response.Message, "success"):
		result.Status = "Success"
		result.Region = "US"
	default:
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}
