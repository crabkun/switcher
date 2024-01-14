package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// 定义一个全局的流量统计map
var trafficMap = make(map[string]int64)
var trafficMutex = &sync.Mutex{}

func listen(rule *ruleStructure, wg *sync.WaitGroup) {
	defer wg.Done()
	//监听
	listener, err := net.Listen("tcp", rule.Listen)
	if err != nil {
		logrus.Errorf("[%s] failed to listen at %s", rule.Name, rule.Listen)
		return
	}
	logrus.Infof("[%s] listing at %s", rule.Name, rule.Listen)
	for {
		//处理客户端连接
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("[%s] failed to accept at %s", rule.Name, rule.Listen)
			time.Sleep(time.Second * 1)
			continue
		}
		//判断黑名单
		blacklistMutex.Lock()
		blacklist := rule.blacklistMap
		blacklistMutex.Unlock()
		if len(blacklist) != 0 {
			clientIP := conn.RemoteAddr().String()
			clientIP = clientIP[0:strings.LastIndex(clientIP, ":")]
			if blacklist[clientIP] {
				logrus.Infof("[%s] disconnected ip in blacklist: %s", rule.Name, clientIP)
				conn.Close()
				continue
			}
		}
		//判断是否是正则模式
		if rule.EnableRegexp {
			go handleRegexp(conn, rule)
		} else {
			go handleNormal(conn, rule)
		}
	}
}

func handleNormal(conn net.Conn, rule *ruleStructure) {
	defer conn.Close()

	var target net.Conn
	var targetAddress string
	for _, v := range rule.Targets {
		c, err := net.Dial("tcp", v.Address)
		if err != nil {
			logrus.Errorf("[%s] try to handle connection %s failed because target %s connected failed, try next target.",
				rule.Name, conn.RemoteAddr(), v.Address)
			continue
		}
		target = c
		targetAddress = v.Address
		break
	}
	if target == nil {
		logrus.Errorf("[%s] unable to handle connection %s because all targets connected failed",
			rule.Name, conn.RemoteAddr())
		return
	}
	logrus.Debugf("[%s] handle connection %s to target %s", rule.Name, conn.RemoteAddr(), target.RemoteAddr())

	defer target.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	var traffic1, traffic2 int64
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	go func() {
		defer wg.Done()
		traffic1 = copyWithTrafficCount(conn, target)
		trafficMutex.Lock()
		trafficMap[ip] += traffic1
		trafficMutex.Unlock()
	}()
	go func() {
		defer wg.Done()
		traffic2 = copyWithTrafficCount(target, conn)
		trafficMutex.Lock()
		trafficMap[ip] += traffic2
		trafficMutex.Unlock()
	}()

	wg.Wait()

	trafficMutex.Lock()
	logrus.Infof("[%s] %s to target %s: This connection traffic: %.2f MB, Total traffic: %.2f MB", rule.Name, conn.RemoteAddr().String(), targetAddress, float64(traffic1 + traffic2) / (1024 * 1024), float64(trafficMap[ip]) / (1024 * 1024))
	trafficMutex.Unlock()
}

func handleRegexp(conn net.Conn, rule *ruleStructure) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(rule.FirstPacketTimeout)))
	firstPacket, err := waitFirstPacket(conn)
	if err != nil {
		logrus.Errorf("[%s] unable to handle connection %s because failed to get first packet : %s",
			rule.Name, conn.RemoteAddr(), err.Error())
		return
	}

	var target net.Conn
	var targetAddress string
	for _, v := range rule.Targets {
		if !v.regexp.Match(firstPacket) {
			continue
		}
		c, err := net.Dial("tcp", v.Address)
		if err != nil {
			logrus.Errorf("[%s] try to handle connection %s failed because target %s connected failed, try next match target.",
				rule.Name, conn.RemoteAddr(), v.Address)
			continue
		}
		target = c
		targetAddress = v.Address
		break
	}
	if target == nil {
		logrus.Errorf("[%s] unable to handle connection %s because no match target",
			rule.Name, conn.RemoteAddr())
		return
	}

	logrus.Debugf("[%s] handle connection %s to target %s", rule.Name, conn.RemoteAddr(), target.RemoteAddr())
	conn.SetReadDeadline(time.Time{})
	io.Copy(target, bytes.NewReader(firstPacket))

	defer target.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	var traffic1, traffic2 int64
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	go func() {
		defer wg.Done()
		traffic1 = copyWithTrafficCount(conn, target)
		trafficMutex.Lock()
		trafficMap[ip] += traffic1
		trafficMutex.Unlock()
	}()
	go func() {
		defer wg.Done()
		traffic2 = copyWithTrafficCount(target, conn)
		trafficMutex.Lock()
		trafficMap[ip] += traffic2
		trafficMutex.Unlock()
	}()

	wg.Wait()

	trafficMutex.Lock()
	logrus.Infof("[%s] %s to target %s: This connection traffic: %.2f MB, Total traffic: %.2f MB", rule.Name, conn.RemoteAddr().String(), targetAddress, float64(traffic1 + traffic2) / (1024 * 1024), float64(trafficMap[ip]) / (1024 * 1024))
	trafficMutex.Unlock()
}

func waitFirstPacket(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func copyWithTrafficCount(dst io.Writer, src io.Reader) int64 {
	buf := make([]byte, 32*1024)
	var traffic int64 = 0
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				traffic += int64(nw)
			}
			if ew != nil {
				break
			}
			if nr != nw {
				logrus.Errorf("partial write")
				break
			}
		}
		if er != nil {
			break
		}
	}
	return traffic
}
