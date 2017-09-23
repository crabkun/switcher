## Switcher
一个用GO语言编写的端口复用工具，能让HTTP/(HTTPS/SSL/TLS)/SSH/RDP/SOCKS5/HTTPProxy/Other跑在同一个端口上。支持复用本地端口，也支持复用其他IP的端口！
## 使用方法
配置好目录下的config.json后，直接运行就行
## 配置
打开程序目录下的config.json，你会看到类似下面的内容

    {
    "ListenAddr": "0.0.0.0:5294",
    "HTTPAddr": "127.0.0.1:80",
	"HTTPHostReplace":"www.hello.com",
    "SSLAddr": "",
    "SSHAddr": "123.123.123.123:22",
    "RDPAddr": "",
	"SOCKS5Addr": "",
	"HTTPProxyAddr":"",
    "DefaultAddr": "127.0.0.1:1234"
    }

上面的配置，如果别人用浏览器连接你电脑的5294端口，就相当于以www.hello.com为host去访问你电脑的80端口

如果别人用SSH客户端去连接你电脑的5294端口，就相当于连接123.123.123.123:22

如果别人用上面列出的协议以外的客户端去连接你电脑的5294端口，就相当于连接你电脑的1234端口



- ListenAddr是本地监听地址，也就是需要复用的端口
- HTTPAddr是目标的HTTP服务器地址
- HTTPHostReplace是HTTP请求的HOST字段替换值，留空为不替换
- SSLAddr是目标的SSL服务器地址（HTTPS/TLS都可以填在这里）
- SSHAddr是目标的SSH服务器地址
- RDPAddr是目标的微软远程桌面服务器地址
- SOCKS5Addr是目标的Socks5代理服务器地址
- HTTPProxyAddr是目标的HTTP代理服务器地址
- DefaultAddr是当上面所有协议都不满足的时候的目标地址

所有地址都要带端口！

用不到的协议可以直接留空。

支持复用本地端口，也支持复用其他IP的端口！所以上面的地址也可以是127.0.0.1，也可以是外网地址。

## 注意事项
本工具的原理是根据客户端建立好连接后第一个数据包的特征进行判断是什么协议，然后再中转到目标地址。
这种方式已知有两个缺陷：

1. 不支持连接建立之后服务器主动握手的协议，例如VNC，FTP，MYSQL…所以DefaultAddr里面不要填写这些协议的地址，没法复用的。
2. SSH无法连接请更换最新版putty或MobaXterm，因为SSH本来属于服务器主动握手的协议，但有些软件遵守有些软件不遵守，所以请选择客户端主动握手的软件。

## 想增加新协议？遇到了问题？
欢迎提issue和Pull Request

## 开源协议
BSD 3-Clause License