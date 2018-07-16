package main

import (
	"github.com/Sirupsen/logrus"
	"sync"
	"net"
)

func listen(config *configStruct, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := net.Listen(config.Net, config.Listen)
	if err != nil {
		logrus.Errorf("[%s] failed to listen at %s(%s)", config.Name, config.Listen, config.Net)
		return
	}
	logrus.Errorf("[%s] listing at %s(%s)", config.Name, config.Listen, config.Net)
	return
}
func main() {
	wg := &sync.WaitGroup{}
	for _, v := range config {
		wg.Add(1)
		go listen(v, wg)
	}
	wg.Wait()
	logrus.Infof("program exited")
}
