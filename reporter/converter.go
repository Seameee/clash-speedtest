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
    <title>Clash 配置转换工具</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/animate.css@4.1.1/animate.min.css" rel="stylesheet">
    <style>
        :root {
            --bs-body-bg: #f8f9fa;
            --bs-body-color: #212529;
            --primary-color: #4F46E5;
            --secondary-color: #6B7280;
            --success-color: #059669;
            --danger-color: #DC2626;
            --warning-color: #D97706;
            --info-color: #3B82F6;
        }
        
        @media (prefers-color-scheme: dark) {
            :root {
                --bs-body-bg: #1F2937;
                --bs-body-color: #F9FAFB;
                --primary-color: #6366F1;
            }
            .container {
                background: #374151 !important;
            }
            .form-control, .form-select {
                background-color: #1F2937;
                border-color: #495057;
                color: #f8f9fa;
            }
            .alert {
                background-color: rgba(55, 65, 81, 0.9);
                border-color: rgba(75, 85, 99, 0.3);
        }
        }

        body {
            padding: 20px;
            background-color: var(--bs-body-bg);
            color: var(--bs-body-color);
            min-height: 100vh;
        }

        .container {
            max-width: 1000px;
            background: white;
            padding: 30px;
            border-radius: 16px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
        }

        .converter-header {
            text-align: center;
            margin-bottom: 2.5rem;
            animation: fadeIn 1s ease-in-out;
        }

        .converter-header h3 {
            margin-bottom: 1.3rem;
            font-weight: 600;
            color: var(--primary-color);
        }

        .converter-header p {
            color: var(--secondary-color);
            margin-bottom: 0;
            font-size: 1.1rem;

        }

        .converter-content {
            background: var(--bs-body-bg);
            border-radius: 12px;
            padding: 25px;
            margin-top: 2rem;
            animation: fadeInUp 0.5s ease-out;
        }

        .config-box {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 2rem;
            border: 1px solid rgba(229, 231, 235, 0.1);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
        }

        .alert {
            margin-bottom: 1.5rem;
            border-radius: 8px;
            padding: 1rem 1.25rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-size: 0.95rem;
            line-height: 1.5;
            animation: slideIn 0.3s ease-out;
        }

        .alert i {
            font-size: 1.25rem;
            flex-shrink: 0;
        }

        .alert-success {
            background-color: rgba(16, 185, 129, 0.1);
            border: 1px solid rgba(16, 185, 129, 0.2);
            color: #10B981;
        }

        .alert-error {
            background-color: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.2);
            color: #EF4444;
        }

        .alert-warning {
            background-color: rgba(245, 158, 11, 0.1);
            border: 1px solid rgba(245, 158, 11, 0.2);
            color: #F59E0B;
        }

        .alert-info {
            background-color: rgba(59, 130, 246, 0.1);
            border: 1px solid rgba(59, 130, 246, 0.2);
            color: #3B82F6;
        }

        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        .form-control {
            background-color: var(--bs-body-bg);
            border: 1px solid var(--bs-border-color);
            color: var(--bs-body-color);
            border-radius: 8px;
            padding: 12px;
            font-size: 14px;
            transition: all 0.2s ease-in-out;
            height: 300px;
            resize: vertical;
        }
        
        .form-control:focus {
            background-color: var(--bs-body-bg);
            border-color: var(--primary-color);
            box-shadow: 0 0 0 0.25rem rgba(99, 102, 241, 0.25);
            color: var(--bs-body-color);
        }
        
        .convert-btn {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 12px 28px;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            gap: 10px;
            margin: 0 10px;
            min-width: 200px;
        }
        
        .convert-btn:hover {
            background: var(--primary-color);
            opacity: 0.9;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
        }
        
        .alert {
            margin-bottom: 1rem;
            border-radius: 8px;
            border: 1px solid transparent;
            padding: 1rem;
            animation: fadeInDown 0.5s ease-out;
        }
        
        .alert-success {
            background-color: rgba(5, 150, 105, 0.1);
            border-color: rgba(5, 150, 105, 0.2);
            color: var(--success-color);
        }
        
        .alert-danger {
            background-color: rgba(220, 38, 38, 0.1);
            border-color: rgba(220, 38, 38, 0.2);
            color: var(--danger-color);
        }
        
        .alert-info {
            background-color: rgba(59, 130, 246, 0.1);
            border-color: rgba(59, 130, 246, 0.2);
            color: var(--info-color);
        }
        
        .copy-btn {
            background: var(--secondary-color);
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }
        
        .copy-btn:hover {
            background: var(--secondary-color);
            opacity: 0.9;
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(107, 114, 128, 0.3);
        }

        .config-section {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 2rem;
            border: 1px solid rgba(229, 231, 235, 0.1);
        }

        .config-section-header {
            display: flex;
            align-items: center;
            margin-bottom: 1rem;
        }

        .config-section-header i {
            font-size: 1.5rem;
            margin-right: 0.75rem;
            color: var(--primary-color);
        }

        .config-section-title {
            font-size: 1.1rem;
            font-weight: 600;
            color: var(--bs-body-color);
            margin: 0;
        }

        .button-group {
            display: flex;
            justify-content: center;
            gap: 1rem;
            margin-top: 2rem;
            flex-wrap: wrap;
        }

        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }

        @keyframes fadeInUp {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        @keyframes fadeInDown {
            from {
                opacity: 0;
                transform: translateY(-20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        /* 模态对话框样式 */
        .message-modal {
            display: none;
            position: fixed;
            top: 24px;
            right: 24px;
            background: var(--bs-body-bg);
            border-radius: 12px;
            padding: 16px 20px;
            min-width: 300px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
            z-index: 1000;
            animation: slideInRight 0.3s ease-out;
            border: 1px solid rgba(229, 231, 235, 0.1);
        }
        
        .message-content {
            display: flex;
            align-items: center;
            gap: 12px;
        }
        
        .message-icon {
            font-size: 20px;
            flex-shrink: 0;
        }
        
        .message-text {
            font-size: 14px;
            color: var(--bs-body-color);
            font-weight: 500;
        }
                /* Footer styles */
        .footer {
            margin-top: 2rem;
            padding-top: 1rem;
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


        @keyframes slideInRight {
            from {
                opacity: 0;
                transform: translateX(100%);
            }
            to {
                opacity: 1;
                transform: translateX(0);
            }
        }
        
        @keyframes slideOutRight {
            from {
                opacity: 1;
                transform: translateX(0);
            }
            to {
                opacity: 0;
                transform: translateX(100%);
            }
        
        }
        
        .message-success .message-icon { color: var(--success-color); }
        .message-error .message-icon { color: var(--danger-color); }
        .message-warning .message-icon { color: var(--warning-color); }
        .message-info .message-icon { color: var(--info-color); }
        
        .message-success { background-color: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.2); }
        .message-error { background-color: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.2); }
        .message-warning { background-color: rgba(245, 158, 11, 0.1); border: 1px solid rgba(245, 158, 11, 0.2); }
        .message-info { background-color: rgba(59, 130, 246, 0.1); border: 1px solid rgba(59, 130, 246, 0.2); }
    </style>
</head>
<body>
    <div class="container">
        <div class="converter-header">
            <h3><i class="bi bi-gear-fill"></i> Clash 配置转换工具</h3>
            <p><i class="bi bi-info-circle-fill me-2"></i> 支持将 Clash/Mihomo 配置转换为 Xray/Sing-box 格式</p>
            <p><i class="bi bi-check-circle-fill me-2"></i> 支持转换的协议：Ss、Ssr、Vmess、Vless、Tuic、Trojan、Hysteria、Hysteria2</p>
        </div>

        <div class="converter-content">
            <div class="config-box">

                <textarea class="form-control" id="convertedConfig" rows="10" readonly 
                    placeholder="转换后的配置将显示在这里..."
                    style="font-family: monospace; white-space: pre-wrap; word-wrap: break-word;"></textarea>
                <div class="d-flex justify-content-end mt-2">
                    <button class="copy-btn" onclick="copyConfig()">
                        <i class="bi bi-clipboard"></i> 复制配置
                    </button>
                </div>
            </div>
            
            <div class="button-group">
                <button class="convert-btn" onclick="convertConfig('xray')">
                    <i class="bi bi-arrow-right-circle"></i> 转换为 Xray 链接
                </button>
                <button class="convert-btn" onclick="convertConfig('singbox')">
                    <i class="bi bi-box-arrow-right"></i> 转换为 Sing-box 配置
                </button>
            </div>
        </div>
    <div class="footer">
        <a href="https://github.com/faceair/clash-speedtest" target="_blank">
        <i class="bi bi-github"></i>原项目</a>
        <a href="https://github.com/OP404OP/clash-speedtest" target="_blank">
        <i class="bi bi-github"></i>修改版</a>
        </div>
    </div>

    <!-- 消息提示对话框 -->
    <div class="message-modal" id="messageModal">
        <div class="message-content">
            <i class="bi message-icon" id="messageIcon"></i>
            <div class="message-text" id="messageText"></div>
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
            
            if (proxies.length === 0) {
                showMessage('未找到可用的节点配置', 'error');
                return;
            }
            
            if (type === 'xray') {
            const convertedContent = proxies.map(proxy => convertToXray(proxy))
                .filter(Boolean)
                .join('\n');
            document.getElementById('convertedConfig').value = convertedContent;
                showMessage('Xray 转换完成！', 'success');
            } else if (type === 'singbox') {
                const {inbounds, outbounds} = convertToSingbox(proxies);
                
                const config = {
                    "log": {
                        "disabled": false,
                        "level": "warn",
                        "timestamp": true
                    },
                    "dns": {
                        "servers": [
                            {
                                "tag": "default-dns",
                                "address": "223.5.5.5",
                                "detour": "direct-out"
                            },
                            {
                                "tag": "system-dns",
                                "address": "local",
                                "detour": "direct-out"
                            },
                            {
                                "tag": "block-dns",
                                "address": "rcode://name_error"
                            },
                            {
                                "tag": "google",
                                "address": "https://dns.google/dns-query",
                                "address_resolver": "default-dns",
                                "address_strategy": "ipv4_only",
                                "strategy": "ipv4_only",
                                "client_subnet": "1.0.1.0"
                            }
                        ],
                        "rules": [
                            {
                                "outbound": "any",
                                "server": "default-dns"
                            },
                            {
                                "query_type": "HTTPS",
                                "server": "block-dns"
                            },
                            {
                                "clash_mode": "direct",
                                "server": "default-dns"
                            },
                            {
                                "clash_mode": "global",
                                "server": "google"
                            },
                            {
                                "rule_set": "cnsite",
                                "server": "default-dns"
                            },
                            {
                                "rule_set": "cnsite-!cn",
                                "server": "google"
                            }
                        ],
                        "strategy": "ipv4_only",
                        "disable_cache": false,
                        "disable_expire": false,
                        "independent_cache": false,
                        "final": "google"
                    },
                    "inbounds": inbounds,
                    "outbounds": [
                        ...outbounds,
                        {
                            "type": "urltest",
                            "tag": "auto",
                            "outbounds": outbounds.map(item => item.tag),
                            "url": "https://www.google.com/generate_204",
                            "interval": "1m",
                            "tolerance": 50,
                            "interrupt_exist_connections": false
                        },
                        {
                            "type": "selector",
                            "tag": "select",
                            "outbounds": outbounds.map(item => item.tag),
                            "default": outbounds[0]?.tag || "auto",
                            "interrupt_exist_connections": false
                        },
                        {
                            "type": "selector",
                            "tag": "解锁节点",
                            "outbounds": outbounds.map(item => item.tag),
                            "default": outbounds[0]?.tag || "",
                            "interrupt_exist_connections": false
                        },
                        {
                            "type": "direct",
                            "tag": "direct-out",
                            "routing_mark": 100
                        },
                        {
                            "type": "block",
                            "tag": "block-out"
                        },
                        {
                            "type": "dns",
                            "tag": "dns-out"
                        }
                    ],
                    "route": {
                        "rules": [
                            {
                                "protocol": "dns",
                                "outbound": "dns-out"
                            },
                            {
                                "protocol": "quic",
                                "outbound": "block-out"
                            },
                            {
                                "clash_mode": "block",
                                "outbound": "block-out"
                            },
                            {
                                "clash_mode": "direct",
                                "outbound": "direct-out"
                            },
                            {
                                "clash_mode": "global",
                                "outbound": "select"
                            },
                            {
                                "rule_set": [
                                    "geosite-netflix",
                                    "geosite-disney",
                                    "geosite-youtube",
                                    "geosite-google",
                                    "geosite-spotify",
                                    "geosite-reddit"
                                ],
                                "outbound": "select"
                            },
                            {
                                "rule_set": [
                                    "geosite-openai"
                                ],
                                "outbound": ""
                            },
                            {
                                "rule_set": [
                                    "cnip",
                                    "cnsite"
                                ],
                                "outbound": "direct-out"
                            },
                            {
                                "rule_set": "cnsite-!cn",
                                "outbound": "select"
                            }
                        ],
                        "rule_set": [
                            {
                                "type": "remote",
                                "tag": "cnsite-!cn",
                                "format": "binary",
                                "url": "https://github.com/SagerNet/sing-geosite/raw/rule-set/geosite-geolocation-!cn.srs",
                                "download_detour": "auto"
                            },
                            {
                                "type": "remote",
                                "tag": "cnip",
                                "format": "binary",
                                "url": "https://github.com/MetaCubeX/meta-rules-dat/raw/sing/geo-lite/geoip/cn.srs",
                                "download_detour": "auto"
                            },
                            {
                                "type": "remote",
                                "tag": "cnsite",
                                "format": "binary",
                                "url": "https://github.com/MetaCubeX/meta-rules-dat/raw/sing/geo-lite/geosite/cn.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-openai",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-openai.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-netflix",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-netflix.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-disney",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-disney.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-youtube",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-youtube.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-google",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-google.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-spotify",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-spotify.srs",
                                "download_detour": "auto"
                            },
                            {
                                "tag": "geosite-reddit",
                                "type": "remote",
                                "format": "binary",
                                "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-reddit.srs",
                                "download_detour": "auto"
                            }
                        ],
                        "auto_detect_interface": true,
                        "final": "select"
                    },
                    "experimental": {
                        "cache_file": {
                            "enabled": true,
                            "path": "cache.db",
                            "store_fakeip": true
                        },
                        "clash_api": {
                            "external_controller": "127.0.0.1:9090",
                            "external_ui": "ui",
                            "external_ui_download_url": "",
                            "external_ui_download_detour": "auto",
                            "default_mode": "rule"
                        }
                    },
                    "ntp": {
                        "enabled": true,
                        "server": "time.apple.com",
                        "server_port": 123,
                        "interval": "30m",
                        "detour": "direct-out"
                    }
                };
                
                document.getElementById('convertedConfig').value = JSON.stringify(config, null, 4);
                showMessage('Sing-box 转换完成！', 'success');
            }
        } catch (err) {
            console.error(err);
            showMessage('转换失败: ' + err.message, 'error');
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
    
    // 转换代理节点到 Sing-box 格式
    function convertToSingbox(proxies) {
        let outbounds = [];
        
        for (const proxy of proxies) {
            if (!proxy || !proxy.type || !proxy.server || !proxy.port) {
                continue;
            }
            
            let outbound = null;
            
            switch (proxy.type) {
                case 'vless':
                    outbound = {
                        "type": "vless",
                        "tag": proxy.name,
                        "server": proxy.server,
                        "server_port": parseInt(proxy.port),
                        "uuid": proxy.uuid,
                        "packet_encoding": "xudp",
                        "tls": {
                            "enabled": proxy.tls,
                            "server_name": proxy.servername || proxy.sni,
                            "insecure": false,
                            "utls": {"enabled": true, "fingerprint": proxy["client-fingerprint"]},
                        },
                    };
                    if (proxy["reality-opts"]) {
                        outbound["tls"]["reality"] = {
                            "enabled": true,
                            "public_key": proxy["reality-opts"]["public-key"],
                            "short_id": proxy["reality-opts"]["short-id"],
                        };
                         outbound["flow"] = proxy.flow;
                    } else if (proxy.tls) {
                         if (proxy.network == "ws") {
                             outbound["transport"] = {
                                "type": "ws",
                                "path": proxy["ws-opts"]?.path || proxy.ws_path || "/",
                                "headers": {"Host": proxy["ws-opts"]?.headers?.Host},
                            };
                         } else if (proxy.network == "tcp") {
                             outbound["flow"] = proxy.flow;
                         }
                    }
                    break;
                case 'vmess':
                    outbound = {
                        "type": "vmess",
                        "tag": proxy.name,
                        "server": proxy.server,
                        "server_port": parseInt(proxy.port),
                        "uuid": proxy.uuid,
                        "security": proxy.cipher || "auto",
                        "alter_id": proxy.alterId || 0,
                    };
                    break;
                case 'ss':
                    outbound = {
                        "type": "shadowsocks",
                        "tag": proxy.name,
                        "server": proxy.server,
                        "server_port": parseInt(proxy.port),
                        "method": proxy.cipher,
                        "password": proxy.password,
                    };
                    break;
                case 'trojan':
                     outbound = {
                        "type": "trojan",
                        "tag": proxy.name,
                        "server": proxy.server,
                        "server_port": parseInt(proxy.port),
                        "password": proxy.password,
                        "tls": {
                            "enabled": true,
                            "server_name": proxy.servername || proxy.sni,
                            "insecure": false,
                        },
                    };
                    break;
                 case 'hysteria2':
                    outbound = {
                        "type": "hysteria2",
                        "tag": proxy.name,
                        "server": proxy.server,
                        "server_port": parseInt(proxy.port),
                        "password": proxy.password,
                        "tls": {
                            "enabled": true,
                            "server_name": proxy.sni,
                            "insecure": proxy["skip-cert-verify"] == true,
                        },
                    };
                    break;
                default:
                    console.warn('Unsupported proxy type:', proxy.type);
                    continue;
            }
            
            if (outbound) {
                outbounds.push(outbound);
            }
        }
        
        const inbounds = [
            {
              "type": "tun",
              "inet4_address": "172.19.0.1/30",
              "inet6_address": "fd00::1/126",
              "auto_route": true,
              "strict_route": true,
              "sniff": true,
              "sniff_override_destination": true,
              "domain_strategy": "prefer_ipv4"
            }
          ];
        
        return {inbounds, outbounds};
    }

    // 显示消息
    function showMessage(message, type) {
        const modal = document.getElementById('messageModal');
        const messageText = document.getElementById('messageText');
        const messageIcon = document.getElementById('messageIcon');
        
        // 根据类型选择图标
        let icon = '';
        switch(type) {
            case 'success':
                icon = 'bi-check-circle-fill';
                break;
            case 'error':
            case 'danger':
                icon = 'bi-x-circle-fill';
                break;
            case 'warning':
                icon = 'bi-exclamation-triangle-fill';
                break;
            case 'info':
                icon = 'bi-info-circle-fill';
                break;
        }
        
        messageIcon.className = 'bi ' + icon + ' message-icon';
        messageText.textContent = message;
        modal.className = 'message-modal message-' + (type === 'danger' ? 'error' : type);
        
        modal.style.display = 'block';
        
        setTimeout(() => {
            modal.style.animation = 'slideOutRight 0.3s ease-out';
            setTimeout(() => {
                modal.style.display = 'none';
                modal.style.animation = 'slideInRight 0.3s ease-out';
            }, 300);
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
