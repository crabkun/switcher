package core

import (
	"regexp"
	"bytes"
)

var Config struct{
	ListenAddr string
	HTTPAddr string
	HTTPHostReplace string
	SSLAddr string
	SSHAddr string
	RDPAddr string
	VNCAddr string
	SOCKS5Addr string
	HTTPProxyAddr string
	DefaultAddr string
}

var RegExpMap map[string]*regexp.Regexp
var RegExp_HostRpl *regexp.Regexp
func InitRegExpMap(){
	RegExpMap=make(map[string]*regexp.Regexp)
	RegExpMap["http"]=regexp.MustCompile(`^(GET|POST|HEAD|DELETE|PUT|CONNECT|OPTIONS|TRACE)`)
	RegExpMap["ssh"]=regexp.MustCompile(`^SSH`)
	RegExpMap["ssl"]=regexp.MustCompile(`^\x16\x03`)
	RegExpMap["rdp"]=regexp.MustCompile(`^\x03\x00\x00\x13`)
	RegExpMap["socks5"]=regexp.MustCompile(`^\x05`)
	RegExpMap["httpProxy"]=regexp.MustCompile(`(^CONNECT)|(Proxy-Connection:)`)
	RegExp_HostRpl=regexp.MustCompile("Host: (.*)")
}

func GetAddrByRegExp(testbuf []byte,ptr *[]byte)(string,string){
	switch {
	case RegExpMap["http"].Match(testbuf) &&
		bytes.Index(testbuf,[]byte("Proxy-Connection:"))==-1 :
			if Config.HTTPHostReplace!=""{
				*ptr=RegExp_HostRpl.ReplaceAll(testbuf,[]byte("Host: "+Config.HTTPHostReplace+"\r"))
			}
		return "http",Config.HTTPAddr
	case RegExpMap["ssh"].Match(testbuf):
		return "ssh",Config.SSHAddr
	case RegExpMap["ssl"].Match(testbuf):
		return "ssl",Config.SSLAddr
	case RegExpMap["rdp"].Match(testbuf):
		return "rdp",Config.RDPAddr
	case RegExpMap["socks5"].Match(testbuf):
		return "socks5",Config.SOCKS5Addr
	case RegExpMap["httpProxy"].Match(testbuf):
		return "httpProxy",Config.HTTPProxyAddr
	default:
		return "default",Config.DefaultAddr
	}
}
