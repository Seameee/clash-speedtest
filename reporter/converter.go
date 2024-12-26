package reporter

import (
	"html/template"
	"net/http"
	"os"
)

const converterTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>配置转换</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        :root {
            --bs-body-bg: #f8f9fa;
            --bs-body-color: #212529;
        }
        @media (prefers-color-scheme: dark) {
            :root {
                --bs-body-bg: #212529;
                --bs-body-color: #f8f9fa;
            }
            .container {
                background: #2c3034 !important;
            }
            .form-control, .form-select {
                background-color: #1a1d20;
                border-color: #495057;
                color: #f8f9fa;
            }
        }
        body {
            padding: 20px;
            background-color: var(--bs-body-bg);
            color: var(--bs-body-color);
        }
        .container {
            max-width: 800px;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .converter-header {
            text-align: center;
            margin-bottom: 2rem;
        }
        .converter-header h3 {
            margin-bottom: 1rem;
        }
        .converter-header p {
            color: #6c757d;
            margin-bottom: 0;
        }
        .rule-section {
            margin: 15px 0;
            padding: 15px;
            border: 1px solid #dee2e6;
            border-radius: 4px;
        }
        .btn-convert {
            min-width: 150px;
            transition: all 0.3s;
        }
        .btn-convert:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .converter-content {
            background: var(--bs-body-bg);
            border-radius: 8px;
            padding: 20px;
            margin-top: 2rem;
        }
        
        .form-control {
            background-color: var(--bs-body-bg);
            border: 1px solid var(--bs-border-color);
            color: var(--bs-body-color);
        }
        
        .form-control:focus {
            background-color: var(--bs-body-bg);
            border-color: #86b7fe;
            box-shadow: 0 0 0 0.25rem rgba(13,110,253,.25);
            color: var(--bs-body-color);
        }
        
        .convert-btn {
            background: #4CAF50;
            color: white;
            border: none;
            padding: 10px 24px;
            border-radius: 6px;
            font-size: 16px;
            cursor: pointer;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }
        
        .convert-btn:hover {
            background: #45a049;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        
        .convert-btn i {
            font-size: 1.2em;
        }
        
        .alert {
            margin-bottom: 1rem;
            border-radius: 6px;
            border: none;
        }
        
        .alert-success {
            background-color: #d4edda;
            color: #155724;
        }
        
        .alert-danger {
            background-color: #f8d7da;
            color: #721c24;
        }
        
        .alert-info {
            background-color: #cce5ff;
            color: #004085;
        }
        
        .copy-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            font-size: 14px;
            cursor: pointer;
            transition: all 0.2s ease;
            display: inline-flex;
            align-items: center;
            gap: 6px;
        }
        
        .copy-btn:hover {
            background: #5a6268;
            transform: translateY(-1px);
        }
        
        .copy-btn i {
            font-size: 1.1em;
        }

        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.5);
        }
        .modal-content {
            background-color: #fefefe;
            margin: 15% auto;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            width: 400px;
            text-align: center;
        }
        .modal-title {
            margin-bottom: 15px;
            color: #333;
            font-size: 1.2em;
        }
        .modal-buttons {
            margin-top: 20px;
        }
        .modal-button {
            padding: 8px 16px;
            margin: 0 5px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 500;
        }
        .modal-button.primary {
            background-color: #3B82F6;
            color: white;
        }
        .modal-button.secondary {
            background-color: #64748B;
            color: white;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="converter-header">
            <h3>配置转换工具</h3>
            <p>将 Clash/Mihomo 配置转换为 Xray 协议链接</p>
        </div>

        <div class="converter-content">
            <div class="mb-4">
                <label class="form-label fw-bold">Xray 配置</label>
                <div class="alert alert-primary" role="alert">
                    <i class="bi bi-info-circle-fill me-2"></i>
                    支持转换的协议：SS、SSR、VMess、VLESS、Trojan、Hysteria、Hysteria2、TUIC
                </div>
                <textarea class="form-control" id="convertedConfig" rows="10" readonly 
                    placeholder="转换后的配置将显示在这里..."
                    style="font-family: monospace; white-space: pre-wrap; word-wrap: break-word;"></textarea>
                <div class="d-flex justify-content-end mt-2">
                    <button class="copy-btn" onclick="copyConfig()">
                        <i class="bi bi-clipboard"></i> 复制配置
                    </button>
                </div>
            </div>
            
            <div class="text-center">
                <button class="convert-btn" onclick="convertConfig('xray')">
                    <i class="bi bi-arrow-right-circle"></i> 开始转换
                </button>
            </div>
        </div>
    </div>

    <!-- 提示对话框 -->
    <div id="alertModal" class="modal">
        <div class="modal-content">
            <div class="modal-title" id="modalMessage"></div>
            <div class="modal-buttons">
                <button class="modal-button primary" onclick="closeModal()">确定</button>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.min.js"></script>
    <script>
    // Base64 编码函数，支持 UTF-8
    function safeBase64Encode(str) {
        return btoa(encodeURIComponent(str).replace(/%([0-9A-F]{2})/g,
            function toSolidBytes(match, p1) {
                return String.fromCharCode('0x' + p1);
            }));
    }

    // Base64 解码函数，支持 UTF-8
    function safeBase64Decode(str) {
        try {
            return decodeURIComponent(atob(str).split('').map(function(c) {
                return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
        } catch (e) {
            return atob(str);
        }
    }

    // 转换配置
    async function convertConfig(type) {
        try {
            showMessage('正在读取配置文件...', 'info');
            const configPath = '{{.ConfigPath}}';
            const response = await fetch('/readfile?path=' + encodeURIComponent(configPath));
            const yamlContent = await response.text();
            
            showMessage('正在转换节点配置...', 'info');
            const config = jsyaml.load(yamlContent);
            
            // 处理不同格式的配置文件
            let proxies = [];
            if (config.proxies) {
                proxies = config.proxies;
            } else if (config['proxy-providers'] && Array.isArray(config['proxy-providers'])) {
                proxies = config['proxy-providers'];
            } else if (config['proxy-providers'] && typeof config['proxy-providers'] === 'object') {
                const providers = Object.values(config['proxy-providers']);
                if (providers.length > 0 && providers[0].proxies) {
                    proxies = providers[0].proxies;
                }
            }
            
            if (!proxies || proxies.length === 0) {
                throw new Error('未找到可用的代理节点');
            }
            
            const convertedContent = proxies.map(proxy => convertToXray(proxy))
                .filter(Boolean)
                .join('\n');
            
            // 显示转换结果
            document.getElementById('convertedConfig').value = convertedContent;
            
            showMessage('转换完成！配置已生成', 'success');
        } catch (error) {
            console.error('转换失败:', error);
            showMessage('转换失败: ' + error.message, 'danger');
        }
    }

    // 转换代理节点到 Xray 格式
    function convertToXray(proxy) {
        if (!proxy || !proxy.type || !proxy.server || !proxy.port) {
            return null;
        }

        switch (proxy.type) {
            case 'ss':
                // 构建基本配置
                const method = proxy.cipher;
                const password = proxy.password;
                const server = proxy.server;
                const port = proxy.port;
                
                // 构建 SS URL
                let uri = 'ss://' + safeBase64Encode(method + ':' + password) + '@' + server + ':' + port;
                
                // 添加插件配置
                if (proxy.plugin) {
                    let pluginStr = '';
                    // 处理 simple-obfs 插件
                    if (proxy.plugin === 'obfs') {
                        pluginStr = 'simple-obfs';
                        if (proxy["plugin-opts"]) {
                            pluginStr += ';obfs=' + (proxy["plugin-opts"].mode || 'http');
                            if (proxy["plugin-opts"].host) {
                                pluginStr += ';obfs-host=' + proxy["plugin-opts"].host;
                            }
                        }
                    }
                    // 处理其他插件...
                    
                    if (pluginStr) {
                        uri += '/?plugin=' + encodeURIComponent(pluginStr);
                    }
                }
                
                // 添加别名
                uri += '#' + encodeURIComponent(proxy.name || '');
                
                return uri;

            case 'ssr':
                const ssrConfig = proxy.server + ':' + proxy.port + ':' + proxy.protocol + ':' + proxy.cipher + ':' + proxy.obfs + ':' + 
                    safeBase64Encode(proxy.password) + '/?obfsparam=' + safeBase64Encode(proxy.obfs_param || '') + 
                    '&protoparam=' + safeBase64Encode(proxy.protocol_param || '') + '&remarks=' + safeBase64Encode(proxy.name);
                return 'ssr://' + safeBase64Encode(ssrConfig);

            case 'vmess':
                const vmessConfig = {
                    v: "2",
                    ps: proxy.name,
                    add: proxy.server,
                    port: proxy.port,
                    id: proxy.uuid,
                    aid: proxy.alterId || 0,
                    net: proxy.network === "grpc" ? "grpc" : (proxy.network || "tcp"),
                    type: "none",
                    host: proxy["ws-opts"]?.headers?.Host || proxy["ws-headers"]?.Host || "",
                    path: proxy.network === "grpc" ? (proxy["grpc-service-name"] || "zhs") : (proxy["ws-opts"]?.path || proxy.ws_path || ""),
                    tls: proxy.tls ? "tls" : "none",
                    sni: proxy.network === "grpc" ? "jpssv5.mry.best" : (proxy.servername || proxy.sni || ""),
                    alpn: proxy.alpn?.join(",") || "",
                    fp: proxy["client-fingerprint"] || "chrome",
                    scy: proxy.cipher || "auto"
                };
                return 'vmess://' + safeBase64Encode(JSON.stringify(vmessConfig));

            case 'vless':
                const params = new URLSearchParams();
                params.append('encryption', 'none');
                params.append('security', proxy["reality-opts"] ? 'reality' : (proxy.tls ? 'tls' : 'none'));
                params.append('type', proxy.network || 'tcp');
                
                if (proxy.servername || proxy.sni) {
                    params.append('sni', proxy.servername || proxy.sni);
                }
                
                if (proxy["reality-opts"]) {
                    params.append('pbk', proxy["reality-opts"]["public-key"]);
                    params.append('sid', proxy["reality-opts"]["short-id"]);
                }
                
                if (proxy["client-fingerprint"]) {
                    params.append('fp', proxy["client-fingerprint"]);
                }
                
                if (proxy.network === 'tcp') {
                    params.append('headerType', 'none');
                }
                
                if (proxy.flow) {
                    params.append('flow', proxy.flow);
                }
                
                if (proxy.network === 'ws') {
                    params.append('path', proxy["ws-opts"]?.path || proxy.ws_path || '/');
                    if (proxy["ws-opts"]?.headers?.Host) {
                        params.append('host', proxy["ws-opts"].headers.Host);
                    }
                }
                
                if (proxy.network === 'grpc') {
                    params.append('serviceName', proxy["grpc-service-name"] || '');
                }
                
                const vlessConfig = proxy.uuid + '@' + proxy.server + ':' + proxy.port + '?' + params.toString() + '#' + encodeURIComponent(proxy.name);
                return 'vless://' + vlessConfig;

            case 'trojan':
                const trojanParams = new URLSearchParams();
                trojanParams.append('security', 'tls');
                trojanParams.append('type', proxy.network || 'tcp');
                
                if (proxy.sni || proxy.servername) {
                    trojanParams.append('sni', proxy.sni || proxy.servername);
                }
                
                if (proxy.network === 'ws') {
                    trojanParams.append('path', proxy["ws-opts"]?.path || proxy.ws_path || '/');
                    if (proxy["ws-opts"]?.headers?.Host) {
                        trojanParams.append('host', proxy["ws-opts"].headers.Host);
                    }
                }
                
                if (proxy.network === 'grpc') {
                    trojanParams.append('serviceName', proxy["grpc-service-name"] || 'zhs');
                }
                
                const trojanConfig = proxy.password + '@' + proxy.server + ':' + proxy.port + '?' + trojanParams.toString() + '#' + encodeURIComponent(proxy.name);
                return 'trojan://' + trojanConfig;

            case 'hysteria':
                let hyConfig = proxy.auth_str + '@' + proxy.server + ':' + proxy.port + '?' +
                    'protocol=' + (proxy.protocol || "udp") +
                    '&up=' + proxy.up_mbps +
                    '&down=' + proxy.down_mbps +
                    (proxy.obfs ? '&obfs=' + proxy.obfs : '') +
                    '&insecure=' + (proxy["skip-cert-verify"] ? "1" : "0") +
                    (proxy.alpn ? '&alpn=' + proxy.alpn.join(',') : '') +
                    '#' + encodeURIComponent(proxy.name);
                return 'hysteria://' + hyConfig;

            case 'hysteria2':
                let hy2Config = proxy.password + '@' + proxy.server + ':' + proxy.port + '?' +
                    'insecure=' + (proxy["skip-cert-verify"] ? "1" : "0") +
                    (proxy.obfs ? '&obfs=' + proxy.obfs.type + '&obfs-password=' + proxy.obfs.password : '') +
                    (proxy.alpn ? '&alpn=' + proxy.alpn.join(',') : '') +
                    (proxy.sni ? '&sni=' + proxy.sni : '') +
                    '#' + encodeURIComponent(proxy.name);
                return 'hysteria2://' + hy2Config;

            case 'tuic':
                let tuicConfig = proxy.uuid + ':' + proxy.password + '@' + proxy.server + ':' + proxy.port + '?' +
                    'congestion_control=' + (proxy.congestion_control || "cubic") +
                    '&udp_relay_mode=' + (proxy.udp_relay_mode || "native") +
                    (proxy.zero_rtt_handshake ? '&zero_rtt_handshake=1' : '') +
                    '&alpn=' + encodeURIComponent(proxy.alpn?.join(",") || "h3") +
                    '#' + encodeURIComponent(proxy.name);
                return 'tuic://' + tuicConfig;

            default:
                console.warn('Unsupported proxy type:', proxy.type);
                return null;
        }
    }

    // 显示消息
    function showMessage(message, type) {
        const alertDiv = document.createElement('div');
        alertDiv.className = 'alert alert-' + type + ' alert-dismissible fade show';
        alertDiv.role = 'alert';
        alertDiv.innerHTML = 
            message +
            '<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>';
        
        const container = document.querySelector('.container');
        container.insertBefore(alertDiv, container.firstChild);
        
        setTimeout(() => {
            alertDiv.remove();
        }, 3000);
    }

    // 复制配置
    async function copyConfig() {
        const configText = document.getElementById('convertedConfig').value;
        if (!configText) {
            showMessage('没有可复制的配置', 'warning');
            return;
        }
        
        try {
            await navigator.clipboard.writeText(configText);
            showMessage('配置已复制到剪贴板', 'success');
        } catch (err) {
            showMessage('复制失败: ' + err.message, 'danger');
        }
    }

    function showModal(message) {
        const modal = document.getElementById('alertModal');
        const messageEl = document.getElementById('modalMessage');
        messageEl.textContent = message;
        modal.style.display = 'block';
    }

    function closeModal() {
        const modal = document.getElementById('alertModal');
        modal.style.display = 'none';
    }

    // 点击模态框外部关闭
    window.onclick = function(event) {
        const modal = document.getElementById('alertModal');
        if (event.target == modal) {
            modal.style.display = 'none';
        }
    }
    </script>
</body>
</html>
`

// HandleConverter 处理配置转换请求
func HandleConverter(w http.ResponseWriter, r *http.Request) {
	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理 OPTIONS 请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	tmpl, err := template.New("converter").Parse(converterTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	configPath := r.URL.Query().Get("config")
	if configPath == "" {
		http.Error(w, "Missing config parameter", http.StatusBadRequest)
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"ConfigPath": configPath,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleReadFile 处理文件读取请求
func HandleReadFile(w http.ResponseWriter, r *http.Request) {
	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理 OPTIONS 请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/yaml")
	w.Write(content)
}
