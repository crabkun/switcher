package bridge

import (
	"net"
	"log"
	"../core"
)

func IOBridge(A net.Conn,B net.Conn){
	buf:=make([]byte,1024)
	for{
		n,err:=A.Read(buf)
		if err!=nil{
			A.Close()
			B.Close()
			return
		}
		B.Write(buf[:n])
	}
}
func NewBridge(client net.Conn,remoteAddr string,connType string,firstPacket []byte){
	r,err:=net.Dial("tcp",remoteAddr)
	if err!=nil{
		log.Printf("客户端%v无法中转到[%s]%s，原因：%s\n",client.RemoteAddr(),connType,remoteAddr,err.Error())
		return
	}
	r.Write(firstPacket)
	go IOBridge(client,r)
	go IOBridge(r,client)
	log.Printf("客户端%v被中转到[%s]%s\n",client.RemoteAddr(),connType,remoteAddr)
}

func NewClientComming(client net.Conn){
	buf:=make([]byte,20480)
	n,err:=client.Read(buf)
	if err!=nil{
		log.Printf("客户端%v处理错误，原因:%s\n",client.RemoteAddr(),err)
		client.Close()
	}
	testbuf:=buf[:n]
	connType,addr:=core.GetAddrByRegExp(testbuf,&testbuf)
	NewBridge(client,addr,connType,testbuf)
}
