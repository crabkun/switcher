package main

import (
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	VERSION = "2.0"
)

func main() {
	logrus.Infof("switcher %s", VERSION)
	wg := &sync.WaitGroup{}
	for _, v := range config.Rules {
		wg.Add(1)
		go listen(v, wg)
	}
	wg.Wait()
	logrus.Infof("program exited")
}
