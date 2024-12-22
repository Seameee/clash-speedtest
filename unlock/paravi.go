package unlock

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// TestParavi 测试 Paravi 解锁情况
func TestParavi(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Paravi",
	}

	data := strings.NewReader(`{"meta_id":17414,"vuid":"3b64a775a4e38d90cc43ea4c7214702b","device_code":1,"app_id":1}`)
	req, err := http.NewRequest("POST", "https://api.paravi.jp/api/v1/playback/auth", data)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.paravi.jp")

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
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Error.Type == "Forbidden" {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	if response.Error.Type == "Unauthorized" {
		result.Status = "Success"
		result.Region = "JPN"
		return result
	}

	result.Status = "Success"
	result.Region = "JPN"
	return result
}
