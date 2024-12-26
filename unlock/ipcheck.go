package unlock

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
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

// geoResult 用于在通道中传递地理位置信息
type geoResult struct {
	country string
	ip      string
	err     error
}

// riskResult 用于在通道中传递风险值信息
type riskResult struct {
	risk interface{}
	err  error
}

func readCompressedBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	default:
		reader = resp.Body
	}

	return io.ReadAll(reader)
}

// GetLocation 获取地理位置信息
func GetLocation(client *http.Client, debugMode bool) (string, error) {
	req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
	if err != nil {
		if debugMode {
			fmt.Printf("创建请求失败: %v\n", err)
		}
		return "N/A", err
	}

	// 设置请求头
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-encoding", "gzip, br")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "Windows")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	if debugMode {
		fmt.Println("发送请求头:")
		for k, v := range req.Header {
			fmt.Printf("%s: %v\n", k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if debugMode {
			fmt.Printf("发送请求失败: %v\n", err)
		}
		return "N/A", err
	}
	defer resp.Body.Close()

	body, err := readCompressedBody(resp)
	if err != nil {
		if debugMode {
			fmt.Printf("读取响应失败: %v\n", err)
		}
		return "N/A", err
	}

	if debugMode {
		fmt.Printf("请求 URL: %s\n", req.URL)
		fmt.Printf("响应状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应头: %v\n", resp.Header)
		fmt.Printf("地理位置 API 响应: %s\n", string(body))
	}

	var geoResp GeoResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		if debugMode {
			fmt.Printf("JSON 解析错误: %v\n", err)
		}
		return "N/A", err
	}

	if geoResp.Country != "" {
		if debugMode {
			fmt.Printf("成功获取到国家信息: %s\n", geoResp.Country)
		}
		return geoResp.Country, nil
	}
	if debugMode {
		fmt.Println("响应中没有国家信息")
	}
	return "N/A", fmt.Errorf("no country information in response")
}

// GetLocationWithRisk 获取地理位置和IP纯净度信息
func GetLocationWithRisk(client *http.Client, debugMode bool) (string, error) {
	if debugMode {
		fmt.Println("开始获取地理位置信息...")
	}

	// 设置更短的超时时间
	client.Timeout = 3 * time.Second

	// 创建通道
	geoChan := make(chan geoResult, 1)
	riskChan := make(chan riskResult, 1)

	// 设置超时通道
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		timeout <- true
	}()

	// 并发获取地理位置
	go func() {
		city, err := GetLocation(client, debugMode)
		if err != nil || city == "N/A" {
			geoChan <- geoResult{"N/A", "", err}
			return
		}
		geoChan <- geoResult{city, "", nil}
	}()

	// 并发获取 IP 和风险值
	go func() {
		req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		// 设置请求头
		req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("accept-encoding", "gzip, br")
		req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("cache-control", "no-cache")
		req.Header.Set("pragma", "no-cache")
		req.Header.Set("priority", "u=0, i")
		req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", "Windows")
		req.Header.Set("sec-fetch-dest", "document")
		req.Header.Set("sec-fetch-mode", "navigate")
		req.Header.Set("sec-fetch-site", "cross-site")
		req.Header.Set("sec-fetch-user", "?1")
		req.Header.Set("upgrade-insecure-requests", "1")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
		resp, err := client.Do(req)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}
		defer resp.Body.Close()

		body, err := readCompressedBody(resp)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		var geoResp struct {
			IP string `json:"ip"`
		}
		if err := json.Unmarshal(body, &geoResp); err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		// 获取 IP 风险值
		riskReq, err := http.NewRequest("GET", fmt.Sprintf("https://ipcheck.ing/api/ipchecking?ip=%s&lang=zh-CN", geoResp.IP), nil)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		// 设置相同的请求头
		for key, values := range req.Header {
			riskReq.Header[key] = values
		}
		// 添加必要的额外头
		riskReq.Header.Set("Referer", "https://ipcheck.ing/")
		riskReq.Header.Set("Origin", "https://ipcheck.ing/")

		riskResp, err := client.Do(riskReq)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}
		defer riskResp.Body.Close()

		riskBody, err := readCompressedBody(riskResp)
		if err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		if debugMode {
			fmt.Printf("风险值响应: %s\n", string(riskBody))
		}

		var riskData struct {
			ProxyDetect struct {
				Risk interface{} `json:"risk"`
			} `json:"proxyDetect"`
		}
		if err := json.Unmarshal(riskBody, &riskData); err != nil {
			riskChan <- riskResult{nil, err}
			return
		}

		riskChan <- riskResult{riskData.ProxyDetect.Risk, nil}
	}()

	// 等待结果或超时
	select {
	case geoRes := <-geoChan:
		// 获取风险值结果
		var riskRes riskResult
		select {
		case riskRes = <-riskChan:
			// 成功获取风险值
		case <-timeout:
			if debugMode {
				fmt.Println("获取风险值超时，将只显示地理位置")
			}
			return geoRes.country, nil
		}

		if geoRes.err != nil || geoRes.country == "N/A" {
			if debugMode {
				fmt.Printf("获取地理位置失败: %v\n", geoRes.err)
			}
			return "N/A", geoRes.err
		}

		location := geoRes.country

		if riskRes.err != nil {
			if debugMode {
				fmt.Printf("获取风险值失败，将只显示地理位置: %v\n", riskRes.err)
			}
			return location, nil
		}

		// 根据风险值返回不同结果
		var riskLevel string
		if debugMode {
			fmt.Printf("风险值类型: %T, 值: %v\n", riskRes.risk, riskRes.risk)
		}

		switch v := riskRes.risk.(type) {
		case float64:
			if v == 0 {
				riskLevel = fmt.Sprintf("[%.0f纯净]", v)
			} else if v < 66 {
				riskLevel = fmt.Sprintf("[%.0f一般]", v)
			} else {
				riskLevel = fmt.Sprintf("[%.0f较差]", v)
			}
		case json.Number:
			f, _ := v.Float64()
			if f == 0 {
				riskLevel = fmt.Sprintf("[%.0f纯净]", f)
			} else if f < 66 {
				riskLevel = fmt.Sprintf("[%.0f一般]", f)
			} else {
				riskLevel = fmt.Sprintf("[%.0f较差]", f)
			}
		case nil:
			riskLevel = fmt.Sprintf("[-- 非常差]")
		default:
			riskLevel = fmt.Sprintf("[%v 未知]", v)
		}

		// 如果启用了风险检测，返回带风险值的结果
		if riskRes.risk != nil {
			return fmt.Sprintf("%s %s", location, riskLevel), nil
		}
		// 否则只返回地理位置
		return location, nil
	case <-timeout:
		if debugMode {
			fmt.Println("获取地理位置超时")
		}
		return "N/A", fmt.Errorf("timeout")
	}
}
