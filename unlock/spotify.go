package unlock

import (
	"encoding/json"
	"io"
	"net/http"
)

// TestSpotify 测试 Spotify 解锁情况
func TestSpotify(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "Spotify",
	}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", UA_Browser)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查响应状态码
	switch resp.StatusCode {
	case 200:
		// 成功获取用户信息
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Status = "Failed"
			result.Info = "Read Response Error"
			return result
		}

		// 解析 JSON 响应
		var data struct {
			Country string `json:"country"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			result.Status = "Failed"
			result.Info = "Parse Error"
			return result
		}

		result.Status = "Success"
		if data.Country != "" {
			result.Region = data.Country
		} else {
			result.Region = "Available"
		}
		return result

	case 401:
		// 需要登录
		result.Status = "Failed"
		result.Info = "Login Required"
		return result

	case 403:
		// 地区限制
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result

	case 404:
		// 服务不可用
		result.Status = "Failed"
		result.Info = "Not Available"
		return result

	default:
		result.Status = "Failed"
		result.Info = "Unknown Error"
		return result
	}
}

func init() {
	// 注册 Spotify 测试
	streamTests = append(streamTests, TestSpotify)
}
