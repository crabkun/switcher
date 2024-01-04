## Switcher V2
一个多功能的端口转发工具，支持转发本地或远程地址的端口，支持正则表达式转发（实现端口复用）。

**这是v2版，如需v1版请切换到v1分支**
## 使用方法
配置好目录下的config.json后，直接运行就行  
也可以使用--config运行参数来指定config.json的路径
## 配置
打开程序目录下的config.json，你会看到类似下面的内容

### 主结构

    {
      "log_level": "debug",
      "rules": [
        规则配置
      ]
    }

### 规则配置
    {
      "name": "test",
      "listen": "0.0.0.0:1234",
      "enable_regexp": false,
      "first_packet_timeout": 5000,
      "blacklist":{
        "1.2.3.4":true,
        "114.114.114.114":true
      },
      "targets": [
        目标配置
      ]
    }
### 目标配置
    {
      "regexp": "正则表达式",
      "address": "127.0.0.1:80"
    }
### 字段解释
#### 主结构 
1. log_level代表日志等级，有info/debug/error可以选
1. rules是规则配置数组，看下面
#### 规则配置
1. name是这个规则的名字，为了在日志中区分不同规则，建议取不同的名字
2. listen是这个规则监听的地址，0.0.0.0:1234代表监听所有网卡的1234端口
3. enable_regexp为是否开启正则表达式模式，后面有解释
4. first_packet_timeout为等待客户端第一个数据包的超时时间(**毫秒**)，仅开启正则表达式模式后有效，后面有解释
5. blacklist为黑名单IP，在黑名单里面的IP且为true的时候则直接断开链接。如不需要使用黑名单可留null
5. targets为目标配置数组，看下面

#### 目标配置
目标配置有两种模式：**普通模式**和**正则模式**。  

**上面**规则配置的**enable_regexp**为true或false决定了这个目标配置是普通模式还是正则模式。  

**普通模式**，即上面的**enable_regexp**为**false**，当存在多个目标的时候，程序会从第一个目标开始尝试连接，如果失败则尝试下一个目标，直到成功为止  

**正则模式**，即上面的**enable_regexp**为**true**，程序会根据客户端第一个数据包来匹配正则表达式，匹配成功就转发到指定的目标。
为了防止客户端长时间不发第一个数据包，故可以通过上面的规则配置的**first_packet_timeout**字段来配置超时时间（毫秒）  

目标配置有两个字段：  
1.regexp字段在正则模式才有用，代表正则表达式。  
2.address字段代表要转发的目标地址和端口，可以是本地的地址，也可以是远程地址

## 示例配置

    {
      "log_level": "debug",
      "rules": [
        {
          "name": "普通模式示例",
          "listen": "0.0.0.0:1234",
          "blacklist":{
            "1.2.3.4":true,
            "114.114.114.114":true
          },
          "targets": [
            {
              "address": "127.0.0.1:80"
            }
          ]
        },
        {
          "name": "正则模式示例",
          "listen": "0.0.0.0:5555",
          "enable_regexp": true,
          "first_packet_timeout": 5000,
          "blacklist":{
            "1.2.3.4":true,
            "114.114.114.114":true
          },
          "targets": [
            {
              "regexp": "^(GET|POST|HEAD|DELETE|PUT|CONNECT|OPTIONS|TRACE)",
              "address": "127.0.0.1:80"
            },
            {
              "regexp": "^SSH",
              "address": "123.123.123.123:22"
            }
          ]
        }
      ]
    }

上面的配置开了两个规则，分别监听本机的1234端口和5555端口。  
1234端口为普通模式，只要有客户端连接，就一股脑转发到127.0.0.1:80，  
5555端口为正则模式，只要有HTTP浏览器连接，就会转发到127.0.0.1:80。只要有SSH客户端连接，就会转发到123.123.123.123:22。            

## 常见协议正则表达式
|协议|正则表达式|
| --- | ---|
|HTTP|^(GET\|POST\|HEAD\|DELETE\|PUT\|CONNECT\|OPTIONS\|TRACE)|
|SSH|^SSH|
|HTTPS(SSL)|^\x16\x03|
|RDP|^\x03\x00\x00|
|SOCKS5|^\x05|
|HTTP代理|(^CONNECT)\|(Proxy-Connection:)|

**复制到JSON中记得注意特殊符号呀，例如^\\x16\\x03得改成^\\\\x16\\\\x03**

## 注意事项
本工具的正则模式的原理是根据客户端建立好连接后第一个数据包的特征进行判断是什么协议，然后再中转到目标地址。
这种方式已知有两个缺陷：

1. 不支持连接建立之后服务器主动握手的协议，例如VNC，FTP，MYSQL…。
2. SSH无法连接请更换最新版putty或MobaXterm，因为SSH本来属于服务器主动握手的协议，但有些软件遵守有些软件不遵守，所以请选择客户端主动握手的软件。

## 遇到了问题？
欢迎提issue或Pull Request

## 开源协议
BSD 3-Clause License
