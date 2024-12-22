package unlock

import (
	"io"
	"net/http"
	"strings"
)

// TestOpenAI 测试 OpenAI/ChatGPT 区域限制
func TestOpenAI(client *http.Client) *StreamResult {
	result := &StreamResult{
		Platform: "ChatGPT",
	}

	// 第一个请求：检查API访问
	req1, err := http.NewRequest("GET", "https://api.openai.com/compliance/cookie_requirements", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	// 设置请求头
	req1.Header.Set("User-Agent", UA_Browser)
	req1.Header.Set("Authority", "api.openai.com")
	req1.Header.Set("Accept", "*/*")
	req1.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req1.Header.Set("Authorization", "Bearer null")
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://platform.openai.com")
	req1.Header.Set("Referer", "https://platform.openai.com/")
	req1.Header.Set("Sec-Ch-Ua", `"Chromium";v="120", "Not_A Brand";v="24", "Google Chrome";v="120"`)
	req1.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req1.Header.Set("Sec-Ch-Ua-Platform", "Windows")
	req1.Header.Set("Sec-Fetch-Dest", "empty")
	req1.Header.Set("Sec-Fetch-Mode", "cors")
	req1.Header.Set("Sec-Fetch-Site", "same-site")

	resp1, err := client.Do(req1)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp1.Body.Close()

	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	// 第二个请求：检查iOS客户端访问
	req2, err := http.NewRequest("GET", "https://ios.chat.openai.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	// 设置请求头
	req2.Header.Set("User-Agent", UA_Browser)
	req2.Header.Set("Authority", "ios.chat.openai.com")
	req2.Header.Set("Accept", "*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req2.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req2.Header.Set("Sec-Ch-Ua", `"Chromium";v="120", "Not_A Brand";v="24", "Google Chrome";v="120"`)
	req2.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req2.Header.Set("Sec-Ch-Ua-Platform", "Windows")
	req2.Header.Set("Sec-Fetch-Dest", "document")
	req2.Header.Set("Sec-Fetch-Mode", "navigate")
	req2.Header.Set("Sec-Fetch-Site", "none")
	req2.Header.Set("Sec-Fetch-User", "?1")
	req2.Header.Set("Upgrade-Insecure-Requests", "1")

	resp2, err := client.Do(req2)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	// 检查响应内容
	hasUnsupportedCountry := strings.Contains(strings.ToLower(string(body1)), "unsupported_country")
	hasVPNBlock := strings.Contains(strings.ToLower(string(body2)), "vpn")

	// 根据不同情况返回结果
	if !hasVPNBlock && !hasUnsupportedCountry {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	if hasVPNBlock && hasUnsupportedCountry {
		result.Info = "Not Available"
	} else if !hasUnsupportedCountry && hasVPNBlock {
		result.Info = "Only Available with Web Browser"
	} else if hasUnsupportedCountry && !hasVPNBlock {
		result.Info = "Only Available with Mobile APP"
	} else {
		result.Info = "Unknown Error"
	}

	return result
}
