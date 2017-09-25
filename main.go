package main

import (
	"io/ioutil"
	"encoding/json"
	"net"
	"log"
	"github.com/crabkun/switcher/core"
	"github.com/crabkun/switcher/bridge"
)

func main(){
	var err error
	log.Println("Switcher V1.0 by Crabkun")
	buf,err:=ioutil.ReadFile("config.json")
	if err!=nil{
		panic("配置文件config.json载入失败:"+err.Error())
	}
	if err=json.Unmarshal(buf,&core.Config);err!=nil{
		panic("配置文件config.json解析失败:"+err.Error())
	}
	core.InitRegExpMap()
	l,err:=net.Listen("tcp",core.Config.ListenAddr)
	if err!=nil{
		panic("监听失败:"+err.Error())
	}
	log.Println("万能端口成功监听于",l.Addr())
	for{
		client,err:=l.Accept()
		if err!=nil{
			panic("客户端接受失败:"+err.Error())
		}
		go bridge.NewClientComming(client)
	}
}