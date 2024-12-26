package reporter

import (
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HTMLReporter struct {
	Results      []*Result
	mutex        sync.Mutex
	outputPath   string
	template     *template.Template
	lastUpdate   time.Time
	updateDelay  time.Duration
	enableUnlock bool
	fastMode     bool
	configPath   string
	totalCount   int
	outputConfig string
}

// Platform 表示流媒体平台信息
type Platform struct {
	Name   string // 平台名称
	Region string // 地区
}

// Result 表示测试结果
type Result struct {
	ProxyName       string        // 代理名称
	ProxyType       string        // 代理类型
	Latency         string        // 延迟
	LatencyValue    int64         // 延迟值(毫秒)
	Jitter          string        // 抖动
	JitterValue     int64         // 抖动值(毫秒)
	PacketLoss      string        // 丢包率
	PacketLossValue float64       // 丢包率值
	Location        template.HTML // 地理位置
	StreamUnlock    string        // 流媒体解锁
	UnlockPlatforms []Platform    // 解锁平台列表
	DownloadSpeed   string        // 下载速度
	DownloadSpeedMB float64       // 下载速度值(MB/s)
	UploadSpeed     string        // 上传速度
	UploadSpeedMB   float64       // 上传速度值(MB/s)
	LastUpdate      time.Time     // 最后更新时间
}

// templateData 用于传递给HTML模板的数据
type templateData struct {
	Results      []*Result
	EnableUnlock bool
	FastMode     bool
	LastUpdate   time.Time
	ConfigPath   string
	TotalCount   int
	OutputConfig string
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>节点报告</title>
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <!-- Bootstrap Icons -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css" rel="stylesheet">
    <!-- Flag Icons -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/lipis/flag-icons@7.2.3/css/flag-icons.min.css">
    <style>
		
        :root {
            --bs-body-font-size: 14px;
        }
        body {
            padding: 20px;
            background-color: #f8f9fa;
            font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", "Noto Sans", "Liberation Sans", Arial, sans-serif;
            line-height: 1.5;
            -webkit-text-size-adjust: 100%;
            -webkit-tap-highlight-color: transparent;
        }
        .container {
            background-color: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .title {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 15px;
        }
        .subtitle {
            color: #6c757d;
            font-size: 14px;
            margin-bottom: 20px;
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
        }
        .table-responsive {
            overflow-x: auto;
            -webkit-overflow-scrolling: touch;
        }
        .table {
            font-size: 13px;
            text-align: center;
            margin-bottom: 0;
            width: 100%;
            border-collapse: collapse;
        }
        .table th {
            text-align: center;
            background-color: #f8f9fa;
            font-weight: 600;
            white-space: nowrap;
            padding: 12px 8px;
            border: 1px solid #dee2e6;
        }
        .table td {
            padding: 8px;
            vertical-align: middle;
            border: 1px solid #dee2e6;
        }
        .table-hover tbody tr:hover {
            background-color: rgba(0,0,0,.075);
        }
        .platform-tag {
            display: inline-block;
            padding: 2px 6px;
            margin: 2px;
            border-radius: 4px;
            font-size: 12px;
            line-height: 1.5;
        }
        .platform-tag.na {
            background-color: #F44336;
            color: white;
            border-radius: 4px;
        }
        .proxy-type {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            background-color: #6c757d;
            color: white;
            white-space: nowrap;
        }
        .location-tag {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            background-color: #00956d;
            color: #ffffff;
            white-space: nowrap;
            gap: 4px;
            justify-content: center;
        }
        .location-tag.bg-danger {
            background-color: #D32F2F;
        }
        .risk-tag {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            white-space: nowrap;
        }
        .risk-tag.bg-success {
            background-color: #4CAF50;
            color: #ffffff;
        }
        .risk-tag.bg-warning {
            background-color: #FFC107;
            color: #000;
        }
        .risk-tag.bg-danger {
            background-color: #F44336;
            color: #ffffff;
        }
        .latency-tag, .jitter-tag, .loss-tag {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            white-space: nowrap;
        }
        .update-info {
            display: inline;
            color: #6c757d;
            font-size: 13px;
        }
        .button-group {
            display: flex;
            justify-content: center;
            gap: 10px;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            font-weight: 500;
            text-align: center;
            vertical-align: middle;
            cursor: pointer;
            user-select: none;
            padding: 8px 16px;
            font-size: .875rem;
            line-height: 1.5;
            border-radius: 6px;
            color: white;
            border: none;
            transition: all 0.2s ease;
            position: relative;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .btn-primary {
            background-color: #3B82F6;
        }
        .btn-secondary {
            background-color: #64748B;
        }
        .btn:hover {
            color: #fff;
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
        }
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }
        .proxy-name {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            white-space: nowrap !important;
            margin-left: 4px;
        }
        .proxy-name.unavailable {
            background-color: #F44336 !important;
            color: #ffffff !important;
            opacity: 0.8;
        }
        .fi {
            font-size: 1.2em;
            vertical-align: middle;
            margin-right: 4px;
        }
        .node-name {
            display: inline-flex;
            align-items: center;
            white-space: nowrap;
        }
        .speed-tag {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            white-space: nowrap;
        }

        /* 定义不同速度等级的样式 */
        .speed-tag.bg-success {
            background-color: #4CAF50;  // 绿色
            color: white;
        }
        .speed-tag.bg-info {
            background-color: #2196F3;  // 蓝色
            color: white;
        }
        .speed-tag.bg-warning {
            background-color: #FFC107;  // 黄色
            color: black;
        }
        .speed-tag.bg-danger {
            background-color: #F44336;  // 红色
            color: white;
        }
        .unavailable-tag {
            display: inline-block;
            padding: 2px 6px;
            margin: 2px;
            border-radius: 4px;
            font-size: 12px;
            background-color: #F44336;
            color: white;
        }
        .proxy-name {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            background-color: #2196F3;
            color: white;
        }
        .control-panel {
            text-align: center;
            margin-bottom: 40px;
        }
        /* 清除浮动 */
        td:after {
            content: "";
            display: table;
            clear: both;
        }
        /* 添加容器样式 */
        .location-container {
            display: flex;
            align-items: center;
            gap: 4px;
            justify-content: center;
        }
        /* Footer styles */
        .footer {
            margin-top: 3rem;
            padding-top: 1.5rem;
            border-top: 1px solid #eee;
            text-align: center;
            color: #6c757d;
            font-size: 0.9rem;
        }
        .footer a {
            display: inline-flex;
            align-items: center;
            color: inherit;
            text-decoration: none;
            padding: 0.5rem 0.75rem;
            margin: 0 0.25rem;
            border-radius: 6px;
            transition: all 0.15s ease;
        }
        .footer a:hover {
            color: #0d6efd;
            background-color: #f8f9fa;
        }
        .footer .bi-github {
            margin-right: 0.375rem;
        }
    </style>
    <!-- Bootstrap Bundle JS (includes Popper) -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js"></script>
    <!-- YAML Parser -->
    <script src="https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.min.js"></script>
</head>
<body>
    <div class="container">
        <div class="toast-container position-fixed top-0 start-50 translate-middle-x p-3">
            <div id="errorToast" class="toast align-items-center text-bg-danger border-0" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="d-flex">
                    <div class="toast-body">
                        <i class="bi bi-exclamation-circle me-2"></i>
                        配置转换服务无法启动，检查是否被终止！
                    </div>
                    <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="Close"></button>
                </div>
            </div>
        </div>
        <div class="header">
            <h3 class="title">节点报告</h3>
            <div class="subtitle">
                <span>测试订阅：{{.ConfigPath}}</span>
                <span>输出订阅：{{if eq .OutputConfig ""}}无{{else}}{{.OutputConfig}}{{end}}</span>
                <span>数量：({{len .Results}}/{{.TotalCount}})</span>
                <span class="update-info">最后更新时间: {{.LastUpdate.Format "2006-01-02 15:04:05"}}</span>
            </div>
        </div>
        <div class="control-panel">
            <div class="button-group">
                <button class="btn btn-primary" onclick="refreshResults()" title="刷新测试结果">
                    <i class="bi bi-arrow-clockwise"></i> 刷新
                </button>
                <div class="d-inline-block" 
                    data-bs-toggle="tooltip"
                    data-bs-placement="top"
                    title="{{if eq .OutputConfig ""}}未指定输出配置文件{{else if lt (len .Results) .TotalCount}}测试未完成，请等待{{else}}转换为Xray链接{{end}}">
                    <button class="btn btn-secondary" 
                        onclick="openConverter('{{.OutputConfig}}')"
                        {{if or (lt (len .Results) .TotalCount) (eq .OutputConfig "")}}
                        disabled 
                        {{end}}
                        style="cursor: {{if or (lt (len .Results) .TotalCount) (eq .OutputConfig "")}}not-allowed{{else}}pointer{{end}};">
                        <i class="bi bi-arrow-left-right"></i> 配置转换
                    </button>
                </div>
            </div>
        </div>
        <div class="table-responsive">
            <table class="table table-hover">
                <thead>
                    <tr>
                        {{if .FastMode}}
                        <th>序号</th>
                        <th>节点</th>
                        <th>协议</th>
                        <th>延迟</th>
                        {{else if .EnableUnlock}}
                        <th>序号</th>
                        <th>节点</th>
                        <th>协议</th>
                        <th>延迟</th>
                        <th>抖动</th>
                        <th>丢包率</th>
                        <th>地理/风险</th>
                        <th>流媒体</th>
                        {{else}}
                        <th>序号</th>
                        <th>节点</th>
                        <th>协议</th>
                        <th>延迟</th>
                        <th>抖动</th>
                        <th>丢包率</th>
                        <th>下载速度</th>
                        <th>上传速度</th>
                        {{end}}
                    </tr>
                </thead>
                <tbody id="results">
                    {{range $index, $result := .Results}}
                    <tr class="{{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}unavailable{{end}}">
                        <td>{{add $index 1}}</td>
                        <td>{{formatProxyName $result.ProxyName}}</td>
                        <td>
                            {{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}
                            <span class="unavailable-tag">{{$result.ProxyType}}</span>
                            {{else}}
                            <span class="proxy-type">{{$result.ProxyType}}</span>
                            {{end}}
                        </td>
                        <td>
                            {{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}
                            <span class="unavailable-tag">{{$result.Latency}}</span>
                            {{else}}
                            <span class="latency-tag" style="{{latencyColor $result.LatencyValue}}">{{$result.Latency}}</span>
                            {{end}}
                        </td>
                        {{if not $.FastMode}}
                            {{if $.EnableUnlock}}
                            <td>
                                {{if or (eq $result.Jitter "N/A") (eq $result.Jitter "0.00ms")}}
                                <span class="unavailable-tag">{{$result.Jitter}}</span>
                                {{else}}
                                <span class="jitter-tag" style="{{jitterColor $result.JitterValue}}">{{$result.Jitter}}</span>
                                {{end}}
                            </td>
                            <td>
                                <span class="loss-tag" style="{{lossColor $result.PacketLossValue}}">{{$result.PacketLoss}}</span>
                            </td>
                            <td>{{.Location}}</td>
                            <td>
                                {{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}
                                <span class="unavailable-tag">N/A</span>
                                {{else}}
                                {{if and $result.UnlockPlatforms (gt (len $result.UnlockPlatforms) 0)}}
                                {{range $result.UnlockPlatforms}}
                                <span class="platform-tag" style="{{randomColor .Name}}">{{.Name}} {{.Region}}</span>
                                {{end}}
                                {{else}}
                                <span class="platform-tag na">N/A</span>
                                {{end}}
                                {{end}}
                            </td>
                            {{else}}
                            <td>
                                {{if or (eq $result.Jitter "N/A") (eq $result.Jitter "0.00ms")}}
                                <span class="unavailable-tag">{{$result.Jitter}}</span>
                                {{else}}
                                <span class="jitter-tag" style="{{jitterColor $result.JitterValue}}">{{$result.Jitter}}</span>
                                {{end}}
                            </td>
                            <td>
                                <span class="loss-tag" style="{{lossColor $result.PacketLossValue}}">{{$result.PacketLoss}}</span>
                            </td>
                            <td>
                                {{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}
                                <span class="unavailable-tag">{{$result.DownloadSpeed}}</span>
                                {{else}}
                                <span class="speed-tag {{getSpeedClass $result.DownloadSpeed}}">{{$result.DownloadSpeed}}</span>
                                {{end}}
                            </td>
                            <td>
                                {{if or (eq $result.Latency "N/A") (eq $result.Latency "0.00ms")}}
                                <span class="unavailable-tag">{{$result.UploadSpeed}}</span>
                                {{else}}
                                <span class="speed-tag {{getSpeedClass $result.UploadSpeed}}">{{$result.UploadSpeed}}</span>
                                {{end}}
                            </td>
                            {{end}}
                        {{end}}
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        <div class="footer">
            <a href="https://github.com/faceair/clash-speedtest" target="_blank">
                <i class="bi bi-github"></i>原项目
            </a>
            <a href="https://github.com/OP404OP/clash-speedtest" target="_blank">
                <i class="bi bi-github"></i>修改版
            </a>
        </div>
    </div>
    <script>
        let refreshTimer = null;

        // 初始化所有的 tooltips
        document.addEventListener('DOMContentLoaded', function() {
            var tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
            tooltipTriggerList.forEach(function(el) {
                new bootstrap.Tooltip(el);
            });
        });

        // 检查测试是否已完成
        function isTestFinished() {
            return {{len .Results}} >= {{.TotalCount}};
        }

        // 手动刷新
        function refreshResults() {
            window.location.reload();
        }

        // 自动刷新
        function startAutoRefresh() {
            // 清除可能存在的旧定时器
            if (refreshTimer) {
                clearInterval(refreshTimer);
                refreshTimer = null;
            }

            // 检查是否需要继续刷新
            if (isTestFinished()) {
                console.log('测试已完成，停止刷新');
                return;
            }

            // 设置5秒定时刷新
            refreshTimer = setInterval(function() {
                if (isTestFinished()) {
                    if (refreshTimer) {
                        clearInterval(refreshTimer);
                        refreshTimer = null;
                    }
                    return;
                }
                window.location.reload();
            }, 5000);
        }

        // 页面加载时启动自动刷新
        window.onload = function() {
            if (!isTestFinished()) {
                startAutoRefresh();
            }
        };

        // 页面卸载时清理定时器
        window.addEventListener('beforeunload', function() {
            stopRefresh();
        });

        // 添加错误消息处理
        function handleTestError() {
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error-message';
            errorDiv.style.display = 'none';
            document.body.appendChild(errorDiv);
        }

        // 检测页面加载出错
        window.addEventListener('error', function(e) {
            handleTestError();
        });

        // 打开配置转换页面
        function openConverter(configPath) {
            window.open('http://127.0.0.1:8080/convert?config=' + encodeURIComponent(configPath), 
                'ConfigConverter', 
                'width=1178,height=904,resizable=yes,scrollbars=yes');
        }
    </script>
</body>
</html>
`

func NewHTMLReporter(outputPath string, enableUnlock bool, configPath string, totalCount int, outputConfig string, fastMode bool) (*HTMLReporter, error) {
	reporter := &HTMLReporter{
		Results:      make([]*Result, 0),
		outputPath:   outputPath,
		updateDelay:  time.Second * 2,
		enableUnlock: enableUnlock,
		fastMode:     fastMode,
		configPath:   configPath,
		totalCount:   totalCount,
		outputConfig: outputConfig,
	}

	// 解析 HTML 模板
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatProxyName": formatProxyName,
		"latencyColor":    generateLatencyColor,
		"jitterColor":     generateJitterColor,
		"lossColor":       generateLossColor,
		"randomColor":     generateRandomColor,
		"getSpeedClass":   getSpeedClass,
	}).Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析模板失败: %v", err)
	}

	reporter.template = tmpl

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer file.Close()

	// 写入初始内容
	data := templateData{
		Results:      reporter.Results,
		EnableUnlock: reporter.enableUnlock,
		FastMode:     reporter.fastMode,
		LastUpdate:   time.Now(),
		ConfigPath:   reporter.configPath,
		TotalCount:   reporter.totalCount,
		OutputConfig: reporter.outputConfig,
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return nil, fmt.Errorf("写入初始内容失败: %v", err)
	}

	return reporter, nil
}

func (r *HTMLReporter) AddResult(result *Result) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 更新结果列表
	r.Results = append(r.Results, result)
	r.lastUpdate = time.Now()

	// 立即更新文件
	file, err := os.Create(r.outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer file.Close()

	// 写入更新内容
	data := templateData{
		Results:      r.Results,
		EnableUnlock: r.enableUnlock,
		FastMode:     r.fastMode,
		LastUpdate:   r.lastUpdate,
		ConfigPath:   r.configPath,
		TotalCount:   r.totalCount,
		OutputConfig: r.outputConfig,
	}

	err = r.template.Execute(file, data)
	if err != nil {
		return fmt.Errorf("写入更新内容失败: %v", err)
	}

	return nil
}

// FormatLocation 格式化地理位置信息
func FormatLocation(location string) template.HTML {
	if location == "N/A" {
		return template.HTML(fmt.Sprintf(`<div class="location-container"><span class="location-tag bg-danger">%s</span></div>`, location))
	}

	// 移除 ANSI 颜色代码
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	location = re.ReplaceAllString(location, "")

	// 分离国家代码和风险值
	parts := strings.Fields(location)
	if len(parts) > 1 {
		country := parts[0]
		riskParts := strings.Split(strings.Trim(parts[1], "[]"), " ")
		var riskValue, riskLevel string
		if len(riskParts) > 0 {
			riskValue = riskParts[0]
			if len(riskParts) > 1 {
				riskLevel = riskParts[1]
			}
		}

		// 根据风险值设置不同的颜色
		var riskClass string
		if riskValue == "--" {
			riskClass = "bg-danger" // 非常差
		} else {
			// 根据风险等级文本直接判断
			switch {
			case strings.Contains(riskLevel, "纯净"):
				riskClass = "bg-success"
			case strings.Contains(riskLevel, "一般"):
				riskClass = "bg-warning"
			case strings.Contains(riskLevel, "较差") || strings.Contains(riskLevel, "非常差"):
				riskClass = "bg-danger"
			default:
				// 如果没匹配到风险等级，使用红色
				riskClass = "bg-danger"
			}
		}

		// 如果有风险等级，显示 "风险值 风险等级"，否则只显示风险值
		var riskText string
		if riskLevel != "" {
			switch riskLevel {
			case "纯净":
				riskText = fmt.Sprintf("%s&nbsp;&nbsp;纯净", riskValue)
			case "一般":
				riskText = fmt.Sprintf("%s&nbsp;&nbsp;一般", riskValue)
			case "较差":
				riskText = fmt.Sprintf("%s&nbsp;&nbsp;较差", riskValue)
			case "非常差":
				riskText = "--&nbsp;&nbsp;非常差"
			default:
				riskText = fmt.Sprintf("%s&nbsp;&nbsp;%s", riskValue, riskLevel)
			}
		} else {
			riskText = riskValue
		}

		return template.HTML(fmt.Sprintf(`<div class="location-container"><span class="location-tag">%s</span><span class="risk-tag %s">%s</span></div>`,
			country, riskClass, riskText))
	}

	return template.HTML(fmt.Sprintf(`<div class="location-container"><span class="location-tag">%s</span></div>`, strings.TrimSpace(location)))
}

// ParseStreamUnlock 解析流媒体解锁信息
func ParseStreamUnlock(unlock string) []Platform {
	if unlock == "N/A" {
		return nil
	}

	platforms := make([]Platform, 0)
	// 首行按逗号分割但忽括号内的号
	var parts []string
	var currentPart string
	var inBrackets bool

	for i := 0; i < len(unlock); i++ {
		char := unlock[i]
		switch char {
		case '[':
			inBrackets = true
			currentPart += string(char)
		case ']':
			inBrackets = false
			currentPart += string(char)
		case ',':
			if inBrackets {
				currentPart += string(char)
			} else {
				if len(strings.TrimSpace(currentPart)) > 0 {
					parts = append(parts, strings.TrimSpace(currentPart))
				}
				currentPart = ""
			}
		default:
			currentPart += string(char)
		}
	}
	if len(strings.TrimSpace(currentPart)) > 0 {
		parts = append(parts, strings.TrimSpace(currentPart))
	}

	for _, part := range parts {
		// 移除方括号
		part = strings.TrimPrefix(part, "[")
		part = strings.TrimSuffix(part, "]")

		// 分割平台和地区
		platformParts := strings.Split(part, ":")
		if len(platformParts) >= 2 {
			platform := Platform{
				Name:   strings.TrimSpace(platformParts[0]),
				Region: strings.TrimSpace(strings.Join(platformParts[1:], ":")),
			}
			platforms = append(platforms, platform)
		} else if len(platformParts) == 1 {
			// 处理没有地区信息的平台
			platform := Platform{
				Name:   strings.TrimSpace(platformParts[0]),
				Region: "",
			}
			platforms = append(platforms, platform)
		}
	}
	return platforms
}

// 生成随机颜色
func generateRandomColor(name string) template.CSS {
	// 预定义一些鲜艳的颜色组（背景色, 文字色）
	colors := []struct {
		bg string
		fg string
	}{
		{"#FF4B4B", "#FFFFFF"}, // 红色背景，白色文字
		{"#4CAF50", "#FFFFFF"}, // 绿色背景，白色文字
		{"#2196F3", "#FFFFFF"}, // 蓝色背景，白色文字
		{"#FF9800", "#FFFFFF"}, // 橙色背景，白色文字
		{"#9C27B0", "#FFFFFF"}, // 紫色背景，白色文字
		{"#00BCD4", "#FFFFFF"}, // 青色背景，白色文字
		{"#FFEB3B", "#880015"}, // 黄色背景，白色文字
		{"#795548", "#FFFFFF"}, // 棕色背景，白色文字
		{"#607D8B", "#FFFFFF"}, // 灰色背景，白色文字
		{"#E91E63", "#FFFFFF"}, // 粉色背景，白色文字
		{"#673AB7", "#FFFFFF"}, // 深紫色背景，白色文字
		{"#3F51B5", "#FFFFFF"}, // 蓝色背景，白色文字
		{"#009688", "#FFFFFF"}, // 茶色背景，白色文字
		{"#FFC107", "#FFFFFF"}, // 琥珀色背景，白色文字
		{"#FF5722", "#FFFFFF"}, // 深橙色背景，白色文字
		{"#8BC34A", "#FFFFFF"}, // 浅绿色背景，白色文字
		{"#CDDC39", "#FFFFFF"}, // 酸橙色背景，白色文字
	}

	// 用名称作为子生成固定的索引
	hash := 0
	for i := 0; i < len(name); i++ {
		hash = int(name[i]) + ((hash << 5) - hash)
	}
	index := hash % len(colors)
	if index < 0 {
		index = -index
	}

	color := colors[index]
	return template.CSS(fmt.Sprintf("background-color: %s; color: %s", color.bg, color.fg))
}

// 格式化代理名称，将国家代码转换为国旗图标
func formatProxyName(name string) template.HTML {
	// 国家代码映射
	countryFlags := map[string]string{
		// 东亚地区
		"🇨🇳": "cn", "CN": "cn", "cn": "cn", // 中国
		"🇭🇰": "hk", "HK": "hk", "hk": "hk", // 香港
		"🇹🇼": "tw", "TW": "tw", "tw": "tw", // 台湾
		"🇯🇵": "jp", "JP": "jp", "jp": "jp", // 日本
		"🇰🇷": "kr", "KR": "kr", "kr": "kr", // 韩国
		// 东南亚地区
		"🇸🇬": "sg", "SG": "sg", "sg": "sg", // 新加坡
		"🇳🇳": "vn", "VN": "vn", "vn": "vn", // 越南
		"🇹🇭": "th", "TH": "th", "th": "th", // 泰国
		"🇮🇩": "id", "ID": "id", "id": "id", // 印度尼西亚
		"🇲🇾": "my", "MY": "my", "my": "my", // 马来西亚
		"🇵🇭": "ph", "PH": "ph", "ph": "ph", // 菲律宾
		// 北美地区
		"🇺🇸": "us", "US": "us", "us": "us", // 美国
		"🇦🇦": "ca", "CA": "ca", "ca": "ca", // 加拿大
		"🇲🇽": "mx", "MX": "mx", "mx": "mx", // 墨西哥
		// 欧地区
		"🇬🇧": "gb", "GB": "gb", "gb": "gb", "UK": "gb", "uk": "gb", // 英国
		"🇫🇷": "fr", "FR": "fr", "fr": "fr", // 法国
		"🇩🇪": "de", "DE": "de", "de": "de", // 德国
		"🇮🇹": "it", "IT": "it", "it": "it", // 意大利
		"🇪🇸": "es", "ES": "es", "es": "es", // 西班牙
		"🇳🇱": "nl", "NL": "nl", "nl": "nl", // 荷兰
		"🇷🇺": "ru", "RU": "ru", "ru": "ru", // 俄罗斯
		"🇨🇭": "ch", "CH": "ch", "ch": "ch", // 瑞士
		"🇸🇪": "se", "SE": "se", "se": "se", // 瑞典
		"🇳🇴": "no", "NO": "no", "no": "no", // 挪威
		"🇫🇮": "fi", "FI": "fi", "fi": "fi", // 芬兰
		"🇵🇱": "pl", "PL": "pl", "pl": "pl", // 波兰
		"🇹🇷": "tr", "TR": "tr", "tr": "tr", // 土耳其
		// 大洋洲
		"🇦🇺": "au", "AU": "au", "au": "au", // 澳大利亚
		"🇳🇿": "nz", "NZ": "nz", "nz": "nz", // 新西兰
		// 其他地区
		"🇮🇳": "in", "IN": "in", "in": "in", // 印度
		"🇧🇷": "br", "BR": "br", "br": "br", // 巴西
		"🇦🇪": "ae", "AE": "ae", "ae": "ae", // 阿联酋
		"🇿🇦": "za", "ZA": "za", "za": "za", // 南非
		"🇮🇱": "il", "IL": "il", "il": "il", // 以色列
	}

	// 辅助函数：生成带国旗的节点名称 HTML
	generateFlagHTML := func(code, name string, isUnavailable bool) template.HTML {
		color := generateRandomColor(name)
		proxyClass := "proxy-name"
		if isUnavailable {
			proxyClass = "proxy-name unavailable"
		}
		return template.HTML(fmt.Sprintf(`<span class="node-name"><span class="fi fi-%s fis"></span><span class="%s" style="%s">%s</span></span>`,
			code, proxyClass, color, name))
	}

	// 1. 首先尝试提取国旗表情号
	emojiRe := regexp.MustCompile(`^([\x{1F1E6}-\x{1F1FF}]{2})\s*(.+)`)
	if matches := emojiRe.FindStringSubmatch(name); len(matches) == 3 {
		flag := matches[1]
		if code, ok := countryFlags[flag]; ok {
			return generateFlagHTML(code, name, strings.Contains(name, "N/A") || strings.Contains(name, "0.00ms"))
		}
	}

	// 2. 尝试从名称中提取国家代码
	codeRe := regexp.MustCompile(`(?i)(^|\||\s+)(US|HK|JP|CN|SG|TW|GB|KR|VN|TH|ID|MY|PH|CA|MX|FR|DE|IT|ES|NL|RU|CH|SE|NO|FI|PL|TR|AU|NZ|IN|BR|AE|ZA|IL|UK)[-_ ]?(.+)`)
	if matches := codeRe.FindStringSubmatch(name); len(matches) > 0 {
		code := strings.ToLower(matches[2])
		if _, ok := countryFlags[code]; ok {
			return generateFlagHTML(code, name, strings.Contains(name, "N/A") || strings.Contains(name, "0.00ms"))
		}
	}

	// 3. 如果都没找到，返回带样式的原始文本
	color := generateRandomColor(name)
	return template.HTML(fmt.Sprintf(`<span class="proxy-name" style="%s">%s</span>`, color, name))
}

// 生成延迟颜色
func generateLatencyColor(latency int64) template.CSS {
	switch {
	case latency <= 100:
		return template.CSS("background-color: #4CAF50; color: white") // 绿色
	case latency <= 200:
		return template.CSS("background-color: #FFC107; color: white") // 黄色
	case latency <= 300:
		return template.CSS("background-color: #FF9800; color: white") // 橙色
	default:
		return template.CSS("background-color: #F44336; color: white") // 红色
	}
}

// 生成抖动颜色
func generateJitterColor(jitter int64) template.CSS {
	switch {
	case jitter <= 50:
		return template.CSS("background-color: #4CAF50; color: white") // 绿色
	case jitter <= 100:
		return template.CSS("background-color: #FFC107; color: white") // 黄色
	case jitter <= 150:
		return template.CSS("background-color: #FF9800; color: white") // 黄色
	default:
		return template.CSS("background-color: #F44336; color: white") // 红色
	}
}

// 生成丢包率颜色
func generateLossColor(loss float64) template.CSS {
	switch {
	case loss <= 1:
		return template.CSS("background-color: #4CAF50; color: white") // 绿色
	case loss <= 5:
		return template.CSS("background-color: #FFC107; color: white") // 黄色
	case loss <= 10:
		return template.CSS("background-color: #FF9800; color: white") // 黄色
	default:
		return template.CSS("background-color: #F44336; color: white") // 红色
	}
}

// 获取速度类
func getSpeedClass(speed string) string {
	// 处理 N/A 情况
	if speed == "N/A" {
		return "bg-danger"
	}

	// 将速度字符串转为数值进行比较
	speedValue := parseSpeedValue(speed)

	switch {
	case speedValue >= 10: // >=10MB/s
		return "bg-success"
	case speedValue >= 5: // >=5MB/s
		return "bg-info"
	case speedValue >= 2: // >=2MB/s
		return "bg-warning"
	default: // <2MB/s 或解析失败
		return "bg-danger"
	}
}

// 辅助函数解析速度值
func parseSpeedValue(speed string) float64 {
	// 移除空格和单位，只保留数字部分
	re := regexp.MustCompile(`[\d.]+`)
	matches := re.FindString(speed)
	if matches == "" {
		return 0
	}

	value, err := strconv.ParseFloat(matches, 64)
	if err != nil {
		return 0
	}
	return value
}
