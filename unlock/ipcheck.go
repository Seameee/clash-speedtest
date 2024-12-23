package unlock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorOrange = "\033[38;5;208m"
	colorWhite  = "\033[37m"
	colorReset  = "\033[0m"
)

type IPRiskResponse struct {
	Risk interface{} `json:"risk"`
}

type GeoResponse struct {
	Country string `json:"country"`
	IP      string `json:"ip"`
}

// GetLocation 获取地理位置信息
func GetLocation(client *http.Client, debugMode bool) (string, error) {
	req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
	if err != nil {
		return "N/A", nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "N/A", nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "N/A", nil
	}

	if debugMode {
		fmt.Printf("地理位置 API 响应: %s\n", string(body))
	}

	var geoResp GeoResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		if debugMode {
			fmt.Printf("JSON 解析错误: %v\n", err)
		}
		return "N/A", nil
	}

	if geoResp.Country != "" {
		return geoResp.Country, nil
	}
	return "N/A", nil
}

// GetLocationWithRisk 获取地理位置和IP纯净度信息
func GetLocationWithRisk(client *http.Client, debugMode bool) (string, error) {
	// 首先获取IP和城市信息
	city, err := GetLocation(client, debugMode)
	if err != nil || city == "N/A" {
		return "N/A", nil
	}

	// 获取IP地址
	req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
	if err != nil {
		return city, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return city, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return city, nil
	}

	var geoResp struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return city, nil
	}

	if geoResp.IP == "" {
		return city, nil
	}

	// 获取IP纯净度
	riskReq, err := http.NewRequest("GET", fmt.Sprintf("https://ipcheck.ing/api/ipchecking?ip=%s&lang=zh-CN", geoResp.IP), nil)
	if err != nil {
		return city, nil
	}

	// 设置必要的请求头
	riskReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	riskReq.Header.Set("Accept", "application/json")
	riskReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	riskReq.Header.Set("Referer", "https://ipcheck.ing/")
	riskReq.Header.Set("Origin", "https://ipcheck.ing")

	riskResp, err := client.Do(riskReq)
	if err != nil {
		return city, nil
	}
	defer riskResp.Body.Close()

	riskBody, err := io.ReadAll(riskResp.Body)
	if err != nil {
		return city, nil
	}

	if debugMode {
		fmt.Printf("风险值 API 响应: %s\n", string(riskBody))
	}

	var riskResult struct {
		ProxyDetect struct {
			Risk interface{} `json:"risk"`
		} `json:"proxyDetect"`
	}
	if err := json.Unmarshal(riskBody, &riskResult); err != nil {
		if debugMode {
			fmt.Printf("风险值 JSON 解析错误: %v\n", err)
		}
		return city, nil
	}

	// 根据风险值返回不同结果
	var riskLevel string
	if debugMode {
		fmt.Printf("风险值类型: %T, 值: %v\n", riskResult.ProxyDetect.Risk, riskResult.ProxyDetect.Risk)
	}

	switch v := riskResult.ProxyDetect.Risk.(type) {
	case float64:
		if v == 0 {
			riskLevel = fmt.Sprintf("%s[%.0f 纯净]%s", colorGreen, v, colorReset)
		} else if v < 66 {
			riskLevel = fmt.Sprintf("%s[%.0f 一般]%s", colorYellow, v, colorReset)
		} else {
			riskLevel = fmt.Sprintf("%s[%.0f 较差]%s", colorOrange, v, colorReset)
		}
	case json.Number:
		f, _ := v.Float64()
		if f == 0 {
			riskLevel = fmt.Sprintf("%s[%.0f 纯净]%s", colorGreen, f, colorReset)
		} else if f < 66 {
			riskLevel = fmt.Sprintf("%s[%.0f 一般]%s", colorYellow, f, colorReset)
		} else {
			riskLevel = fmt.Sprintf("%s[%.0f 较差]%s", colorOrange, f, colorReset)
		}
	case nil:
		riskLevel = fmt.Sprintf("%s[-- 非常差]%s", colorRed, colorReset)
	default:
		riskLevel = fmt.Sprintf("%s[%v 未知]%s", colorWhite, v, colorReset)
	}

	return fmt.Sprintf("%s %s", city, riskLevel), nil
}
