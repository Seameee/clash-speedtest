package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestTVer 测试 TVer 解锁情况
func TestTVer(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "TVer",
	}

	req, err := http.NewRequest("GET", "https://edge.api.brightcove.com/playback/v1/accounts/5102072605001/videos/ref%3Adesign_5102072605001", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept", "application/json;pk=BCpkADawqM0Z5JvJ1ZXP-nKGEIxOyYVK1zF6Y2FyEV2Zk7Zc")

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

	if strings.Contains(string(body), "geo") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	result.Status = "Success"
	result.Region = "JPN"
	return result
}
