# 详细运行参数说明
> :warning: 本文所有命令, Windows环境使用"openp2p.exe", Linux环境使用"./openp2p"


## 安装和监听
```
./openp2p install -node OFFICEPC1 -user USERNAME1 -password PASSWORD1  
或
./openp2p -d -node OFFICEPC1 -user USERNAME1 -password PASSWORD1  
# 注意Windows系统把“./openp2p” 换成“openp2p.exe”
```
>* install: 安装模式【推荐】，会安装成系统服务，这样它就能随系统自动启动
>* -d: daemon模式。发现worker进程意外退出就会自动启动新的worker进程
>* -node: 独一无二的节点名字，唯一标识
>* -user: 独一无二的用户名字，该节点属于这个user
>* -password: 密码
>* -sharebandwidth: 作为共享节点时提供带宽，默认10mbps. 如果是光纤大带宽，设置越大效果越好. -1表示不共享，该节点只在私有的P2P网络使用。不加入共享的P2P网络，这样也意味着无法使用别人的共享节点
>* -loglevel: 需要查看更多调试日志，设置0；默认是1

## 连接
```
./openp2p -d -node HOMEPC123 -user USERNAME1 -password PASSWORD1 -appname OfficeWindowsRemote -peernode OFFICEPC1 -dstip 127.0.0.1 -dstport 3389 -srcport 23389 -protocol tcp
使用配置文件，建立多个P2PApp
./openp2p -d -f    
./openp2p -f 
```
>* -appname: 这个P2P应用名字
>* -peernode: 目标节点名字
>* -dstip: 目标服务地址，默认本机127.0.0.1
>* -dstport: 目标服务端口，常见的如windows远程桌面3389，Linux ssh 22
>* -protocol: 目标服务协议 tcp、udp
>* -peeruser: 目标用户，如果是同一个用户下的节点，则无需设置
>* -peerpassword: 目标密码，如果是同一个用户下的节点，则无需设置

## 配置文件
一般保存在当前目录，安装模式下会保存到 `C:\Program Files\OpenP2P\config.json` 或 `/usr/local/openp2p/config.json`
希望修改参数，或者配置多个P2PApp可手动修改配置文件

配置实例
```
{
  "network": {
    "Node": "hhd1207-222",
    "User": "USERNAME1",
    "Password": "PASSWORD1",
    "ShareBandwidth": -1,
    "ServerHost": "api.openp2p.cn",
    "ServerPort": 27182,
    "UDPPort1": 27182,
    "UDPPort2": 27183
  },
  "apps": [
    {
      "AppName": "OfficeWindowsPC",
      "Protocol": "tcp",
      "SrcPort": 23389,
      "PeerNode": "OFFICEPC1",
      "DstPort": 3389,
      "DstHost": "localhost",
      "PeerUser": "",
      "PeerPassword": ""
    },
    {
      "AppName": "OfficeServerSSH",
      "Protocol": "tcp",
      "SrcPort": 22,
      "PeerNode": "OFFICEPC1",
      "DstPort": 22,
      "DstHost": "192.168.1.5",
      "PeerUser": "",
      "PeerPassword": ""
    }
  ]
}
```

## 升级客户端
```
# update local client
./openp2p update  
# update remote client
curl --insecure 'https://openp2p.cn:27182/api/v1/device/YOUR-NODE-NAME/update?user=&password='
```

Windows系统需要设置防火墙放行本程序，程序会自动设置，如果设置失败会影响连接功能。
Linux系统（Ubuntu和CentOS7）的防火墙默认配置均不会有影响，如果不行可尝试关闭防火墙
```
systemctl stop firewalld.service
systemctl start firewalld.service
firewall-cmd --state
```

## 卸载
```
./openp2p uninstall
```