# Clash-SpeedTest

基于 Clash/Mihomo 核心的测速工具，快速测试你的节点速度和流媒体解锁情况。

Features:
1. 无需额外的配置，直接将 Clash/Mihomo 配置本地文件路径或者订阅地址作为参数传入即可
2. 支持 Proxies 和 Proxy Provider 中定义的全部类型代理节点，兼容性跟 Mihomo 一致
3. 不依赖额外的 Clash/Mihomo 进程实例，单一工具即可完成测试
4. 代码简单而且开源，不发布构建好的二进制文件，保证你的节点安全
5. 支持流媒体解锁检测功能，可以测试 40+ 个主流流媒体平台
6. 支持显示节点的地理位置信息和 IP 纯净度，支持 IP 历史记录检测
7. 支持显示节点延迟、抖动和丢包率
8. 支持调试模式查看详细的解锁测试信息和节点屏蔽信息
9. 支持自定义并发数，提高测试效率
10. 在开启 -unlock 模式下，将跳过上传速度和下载速度测试
11. 支持快速测试模式（-fast），仅测试节点延迟
12. 支持生成美观的 HTML 报告
    - 支持实时刷新和配置转换功能
    - 支持表格排序和过滤
    - 支持一键导出测试结果截图
    - 支持配置转换（Clash/Mihomo -> sing-box/Xray）
13.支持节点测试结果的颜色标记，直观显示节点状态
14.支持国旗图标显示，美化节点显示效果
15.支持关键词屏蔽功能，可以屏蔽不需要的节点（如倍率节点等）
16.支持 IP 检测的 Cloudflare 验证伪装，提高检测准确性

<img width="1332" alt="image" src="https://github.com/user-attachments/assets/fdc47ec5-b626-45a3-a38a-6d88c326c588">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/main/unlock.png?raw=true">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/main/unlockdebug.png?raw=true">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/af56ceadad44bc93f5525e2ac962ce8fff1cb00f/unlockrisk.png?raw=true">

## HTML报告
- -fast: 快速测试模式，仅测试节点延迟

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/ab8b8b356cb18726c6b07ecd2a8d2620b5f32ed0/fast.png?raw=true">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/ab8b8b356cb18726c6b07ecd2a8d2620b5f32ed0/html_fast.png?raw=true">

- 测速全部节点

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/ab8b8b356cb18726c6b07ecd2a8d2620b5f32ed0/html_speed.png?raw=true">

- -unlock -risk: 启用流媒体解锁检测和IP风险检测

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/ab8b8b356cb18726c6b07ecd2a8d2620b5f32ed0/2024-12-26-09.png?raw=true">

- HTML报告配置转换（Clash/Mihomo -> Xray/sing-box）

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/ab8b8b356cb18726c6b07ecd2a8d2620b5f32ed0/html_convert.png?raw=true">

- -b '倍率|x' -debug：屏蔽指定关键词节点并开启DEBUG模式


## 使用方法

```bash
# 支持从源码安装，或从 Release 里下载由 Github Action 自动构建的二进制文件
> go install github.com/faceair/clash-speedtest
> go install github.com/OP404OP/clash-speedtest (二次修改)

# 查看帮助
> clash-speedtest -h
Usage of clash-speedtest:
  -c string
        configuration file path, also support http(s) url
  -f string
        filter proxies by name, use regexp (default ".*")
  -b string
        block proxies by keywords, use | to separate multiple keywords (example: -b 'rate|x1|1x')
  -server-url string
        server url for testing proxies (default "https://speed.cloudflare.com")
  -download-size int
        download size for testing proxies (default 50MB)
  -upload-size int
        upload size for testing proxies (default 20MB)
  -timeout duration
        timeout for testing proxies (default 5s)
  -concurrent int
        download concurrent size (default 4)
  -output string
        output config file path (default "")
  -max-latency duration
        If not specified, the default filter latency is greater than 0) Latency filtering threshold, in ms, nodes greater than this value will be filtered, for example -max-latency 1000ms means filtering nodes with latency greater than 1000 ms
  -min-speed float
        If not specified, the default filter is nodes with latency greater than 0) Speed ​​filter threshold, in MB/s, nodes with a speed less than this value will be filtered, for example -min-speed 10 means filtering nodes with a speed less than 10 MB/s
  -unlock
        enable streaming media unlock detection(Unlock detection with streaming media when OUTPUT is enabled, only nodes with delay greater than 0 are saved by default)
  -unlock-concurrent int
        concurrent size for unlock testing (default 5)
  -debug
        enable debug mode (for unlock testing and node blocking)
  -risk
        enable IP risk checking when unlock testing is enabled
  -html string
        output HTML report path (default "")
        By default, the configuration conversion service is started on the local port 8080.
  -fast
        enable fast mode, only test latency

# 演示：

# 1. 测试全部节点，使用 HTTP 订阅地址
# 请在订阅地址后面带上 flag=meta 参数，否则无法识别出节点类型
> clash-speedtest -c 'https://domain.com/api/v1/client/subscribe?token=secret&flag=meta'

# 2. 测试香港节点，使用正则表达式过滤，使用本地文件
> clash-speedtest -c ~/.config/clash/config.yaml -f 'HK|港'
节点                                        	带宽          	延迟
Premium|广港|IEPL|01                        	484.80KB/s  	815.00ms
Premium|广港|IEPL|02                        	N/A         	N/A
Premium|广港|IEPL|03                        	2.62MB/s    	333.00ms
Premium|广港|IEPL|04                        	1.46MB/s    	272.00ms
Premium|广港|IEPL|05                        	3.87MB/s    	249.00ms

# 3. 当然你也可以混合使用
> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta,/home/.config/clash/config.yaml"

# 4. 屏蔽指定关键词节点并查看详细信息
> clash-speedtest -c config.yaml -b "倍率|x1|1x" -debug
Debug 模式已启用
[Debug] 节点统计信息:
[Debug] 总节点数: 47
[Debug] 已屏蔽节点数: 16
[Debug] 剩余节点数: 31
[Debug] 被屏蔽的节点:
[Debug] - 🇯🇵 日本W01 | x0.8 (匹配关键词: x)
[Debug] - 🇯🇵 日本W06 | 下载专用 | x0.01 (匹配关键词: x)
[Debug] - 🇯🇵 日本W04 | x0.8 (匹配关键词: x)
[Debug] - 🇯🇵 日本W07 | x0.8 (匹配关键词: x)
[Debug] - 🇭🇰 香港W06 | x0.8 (匹配关键词: x)
[Debug] - 🇭🇰 香港W08 | x0.8 (匹配关键词: x)

# 5. 筛选出延迟低于 800ms 且下载速度大于 5MB/s 的节点，并输出到 filtered.yaml
> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.yaml -max-latency 800ms -min-speed 5
# 筛选后的配置文件可以直接粘贴到 Clash/Mihomo 中使用，或是贴到 Github\Gist 上通过 Proxy Provider 引用。

# 6. 启用流媒体解锁检测
> clash-speedtest -c config.yaml -unlock
# 此命令将测试节点对各个流媒体平台的解锁情况，支持以下功能：
# - 测试 40+ 个主流流媒体平台，包括 Netflix、Disney+、HBO Max、Prime Video 等
# - 显示解锁区域信息（例如 Netflix:SG 表示解锁新加坡区）
# - 自动跳过延迟过高或无法连接的节点
# - 支持并发检测以提高测试速度
# - 支持使用 -f 参数过滤要测试的节点，例如：
#   > clash-speedtest -c config.yaml -unlock -f 'HK|港'  # 只测试香港节点
#   > clash-speedtest -c config.yaml -unlock -f 'US|美'  # 只测试美国节点

# 7. 启用调试模式查看详细解锁信息
> clash-speedtest -c config.yaml -unlock -debug
# 在解锁检测过程中显示详细的测试信息，包括：
# - 每个节点测试的具体流媒体平台
# - 测试结果（成功/失败）
# - 解锁区域
# - 错误信息（如果有）
# - 节点统计信息（如果有）

# 8. 控制解锁测试并发数
> clash-speedtest -c config.yaml -unlock -unlock-concurrent 10
# 默认并发数为 5，可以根据需要调整以平衡速度和稳定性

# 9. 启用 IP 纯净度检测
> clash-speedtest -c config.yaml -unlock -risk
# 此命令将检测节点 IP 的纯净度，显示格式为：[分数 级别]
# - [0 纯净]：绿色，表示 IP 非常干净
# - [1-65 一般]：黄色，表示 IP 有轻微风险
# - [66+ 较差]：橙色，表示 IP 风险较高
# - [-- 非常差]：红色，表示 IP 风险非常高
# - [值 未知]：白色，表示无法获取风险信息

# 10. 生成 HTML 报告
> clash-speedtest -c config.yaml -html report.html
# 此命令将生成一个美观的 HTML 报告，支持以下功能：
# - 实时自动刷新测试结果(5秒刷新一次)
# - 支持手动刷新
# - 配置转换功能
# - 颜色标记显示节点质量
# - 国旗图标显示
# - 生成测试报告长截图

# 11. 快速测试模式
> clash-speedtest -c config.yaml -fast
# 此命令将只测试节点延迟，跳过其他测试项目，适用于：
# - 快速检查节点是否可用
# - 只需要检查延迟的场景
# - 需要快速得到测试结果的场景

检测结果示例：
节点                 延迟       抖动    	丢包率	  地理位置	        流媒体
Premium|广港|IEPL|01 180.00ms  12.5ms    0.0%	  香港 [0 纯净]	  Netflix:SG, Disney+, HBO Max, Prime Video
Premium|广港|IEPL|02 165.00ms  8.2ms     0.0%	  香港 [50 一般]	  Netflix:HK, Disney+, HBO Max, Prime Video
Premium|广港|IEPL|03 195.00ms  15.8ms    1.2%	  香港 [66 较差]	  Netflix:HK, Disney+, HBO Max, Prime Video

演示项目：[https://github.com/faceair/freesub](https://github.com/faceair/freesub) 通过 Github Action 使用本工具对免费订阅进行测速，并保存结果。

```
## 测速原理

通过 HTTP GET 请求下载指定大小的文件，默认使用 https://speed.cloudflare.com (50MB) 进行测试，计算下载时间得到下载速度。

测试结果：
1. 带宽 是指下载指定大小文件的速度，即一般理解中的下载速度。当这个数值越高时表明节点的出口带宽越大。
2. 延迟 是指 HTTP GET 请求拿到第一个字节的的响应时间，即一般理解中的 TTFB。当这个数值越低时表明你本地到达节点的延迟越低，可能意味着中转节点有 BGP 部署、出海线路是 IEPL、IPLC 等。
3. IP 纯净度 是指节点 IP 的风险等级，分为纯净、一般、较差、非常差四个等级。数值越低表示 IP 越干净，被各种网站或服务屏蔽的可能性越小。

请注意带宽跟延迟是两个独立的指标，两者并不关联：
1. 可能带宽很高但是延迟也很高，这种情况下你下载速度很快但是打开网页的时候却很慢，可能是是中转节点没有 BGP 加速，但出海线路带宽很充足。
2. 可能带宽很低但是延迟也很低，这种情况下你打开网页的时候很快但是下载速度很慢，可能是中转节点有 BGP 加速，但出海线路的 IEPL、IPLC 带宽很小。

Cloudflare 是全球知名的 CDN 服务商，其提供的测速服务器到海外绝大部分的点速度都很快，一般情况下都没有必要自建测速服务器。

如果你不想使用 Cloudflare 的测速服务器，可以自己搭建一个测速服务器。

```shell
# 在您需要进行测速的服务器上安装和启动测速服务器
> go install github.com/faceair/clash-speedtest/download-server
> download-server

# 此时在本地使用 http://your-server-ip:8080 作为 server-url 即可
> clash-speedtest --server-url "http://your-server-ip:8080"
```

## License

[MIT](LICENSE)
