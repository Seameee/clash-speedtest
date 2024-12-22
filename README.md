# Clash-SpeedTest

基于 Clash/Mihomo 核心的测速工具，快速测试你的节点速度和流媒体解锁情况。

Features:
1. 无需额外的配置，直接将 Clash/Mihomo 配置本地文件路径或者订阅地址作为参数传入即可
2. 支持 Proxies 和 Proxy Provider 中定义的全部类型代理节点，兼容性跟 Mihomo 一致
3. 不依赖额外的 Clash/Mihomo 进程实例，单一工具即可完成测试
4. 代码简单而且开源，不发布构建好的二进制文件，保证你的节点安全
5. 支持流媒体解锁检测功能，可以测试 40+ 个主流流媒体平台
6. 支持显示节点的地理位置信息
7. 支持显示节点延迟、抖动和丢包率
8. 支持调试模式查看详细的解锁测试信息
9. 支持自定义并发数，提高测试效率
10. 在开启 -unlock 模式下，将跳过上传速度和下载速度测试

<img width="1332" alt="image" src="https://github.com/user-attachments/assets/fdc47ec5-b626-45a3-a38a-6d88c326c588">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/main/unlock.png?raw=true">

<img width="1332" alt="image" src="https://github.com/OP404OP/clash-speedtest/blob/main/unlockdebug.png?raw=true">

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
        filter latency greater than this value (default 800ms)
  -min-speed float
        filter speed less than this value(unit: MB/s) (default 5)
  -unlock
        enable streaming media unlock detection
  -unlock-concurrent int
        concurrent size for unlock testing (default 5)
  -debug
        enable debug mode for unlock testing

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

# 4. 筛选出延迟低于 800ms 且下载速度大于 5MB/s 的节点，并输出到 filtered.yaml
> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.yaml -max-latency 800ms -min-speed 5
# 筛选后的配置文件可以直接粘贴到 Clash/Mihomo 中使用，或是贴到 Github\Gist 上通过 Proxy Provider 引用。

# 5. 启用流媒体解锁检测
> clash-speedtest -c config.yaml -unlock
# 此命令将测试节点对各个流媒体平台的解锁情况，支持以下功能：
# - 测试 40+ 个主流流媒体平台，包括 Netflix、Disney+、HBO Max、Prime Video 等
# - 显示解锁区域信息（例如 Netflix:SG 表示解锁新加坡区）
# - 自动跳过延迟过高或无法连接的节点
# - 支持并发检测以提高测试速度
# - 支持使用 -f 参数过滤要测试的节点，例如：
#   > clash-speedtest -c config.yaml -unlock -f 'HK|港'  # 只测试香港节点
#   > clash-speedtest -c config.yaml -unlock -f 'US|美'  # 只测试美国节点

# 6. 启用调试模式查看详细解锁信息
> clash-speedtest -c config.yaml -unlock -debug
# 在解锁检测过程中显示详细的测试信息，包括：
# - 每个节点测试的具体流媒体平台
# - 测试结果（成功/失败）
# - 解锁区域
# - 错误信息（如果有）

# 7. 控制解锁测试并发数
> clash-speedtest -c config.yaml -unlock -unlock-concurrent 10
# 默认并发数为 5，可以根据需要调整以平衡速度和稳定性

检测结果示例：
节点                 延迟       抖动    	丢包率	  地理位置	 流媒体
Premium|广港|IEPL|01 180.00ms  12.5ms    0.0%	   香港	   Netflix:SG, Disney+, HBO Max, Prime Video, Bilibili China Mainland Only
Premium|广港|IEPL|02 165.00ms  8.2ms     0.0%	   香港	   Netflix:HK, Disney+, HBO Max, Prime Video, Bilibili HongKong/Macau/Taiwan
Premium|广港|IEPL|03 195.00ms  15.8ms    1.2%	   香港	   Netflix:HK, Disney+, HBO Max, Prime Video, Bilibili Taiwan Only

演示项目：[https://github.com/faceair/freesub](https://github.com/faceair/freesub) 通过 Github Action 使用本工具对免费订阅进行测速，并保存结果。

## 测速原理

通过 HTTP GET 请求下载指定大小的文件，默认使用 https://speed.cloudflare.com (50MB) 进行测试，计算下载时间得到下载速度。

测试结果：
1. 带宽 是指下载指定大小文件的速度，即一般理解中的下载速度。当这个数值越高时表明节点的出口带宽越大。
2. 延迟 是指 HTTP GET 请求拿到第一个字节的的响应时间，即一般理解中的 TTFB。当这个数值越低时表明你本地到达节点的延迟越低，可能意味着中转节点有 BGP 部署、出海线路是 IEPL、IPLC 等。

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
