package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type configStructure struct {
	LogLevel string           `json:"log_level"`
	Rules    []*ruleStructure `json:"rules"`
}

type ruleStructure struct {
	Name         string `json:"name"`
	Listen       string `json:"listen"`
	EnableRegexp bool   `json:"enable_regexp"`
	Targets      []*struct {
		Regexp  string         `json:"regexp"`
		regexp  *regexp.Regexp `json:"-"`
		Address string         `json:"address"`
	} `json:"targets"`
	FirstPacketTimeout uint64          `json:"first_packet_timeout"`
	BlacklistFile      string          `json:"blacklist_file"`
	blacklistMap       map[string]bool `json:"-"`
}

var config *configStructure
var blacklistMutex = &sync.Mutex{}

func init() {
	cfgPath := flag.String("config", "config.json", "config.json file path")
	flag.Parse()

	buf, err := ioutil.ReadFile(*cfgPath)
	if err != nil {
		logrus.Fatalf("failed to load config json: %s", err.Error())
	}

	if err := json.Unmarshal(buf, &config); err != nil {
		logrus.Fatalf("failed to load config json: %s", err.Error())
	}

	if len(config.Rules) == 0 {
		logrus.Fatalf("empty rule")
	}
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logrus.Fatalf("invalid log_level")
	}
	logrus.SetLevel(lvl)

	for i, v := range config.Rules {
		if err := v.verify(); err != nil {
			logrus.Fatalf("verity rule failed at pos %d : %s", i, err.Error())
		}
	}
}

func (c *ruleStructure) verify() error {
	if c.Name == "" {
		return fmt.Errorf("empty name")
	}
	if c.Listen == "" {
		return fmt.Errorf("invalid listen address")
	}
	if len(c.Targets) == 0 {
		return fmt.Errorf("invalid targets")
	}
	if c.EnableRegexp {
		if c.FirstPacketTimeout == 0 {
			c.FirstPacketTimeout = 5000
		}
	}
	for i, v := range c.Targets {
		if v.Address == "" {
			return fmt.Errorf("invalid address at pos %d", i)
		}
		if c.EnableRegexp {
			r, err := regexp.Compile(v.Regexp)
			if err != nil {
				return fmt.Errorf("invalid regexp at pos %d : %s", i, err.Error())
			}
			v.regexp = r
		}
	}
	if c.BlacklistFile != "" {
		err := loadBlacklist(c.BlacklistFile, &c.blacklistMap)
		if err != nil {
			return fmt.Errorf("failed to load blacklist: %s", err.Error())
		}
		go watchBlacklist(c.BlacklistFile, &c.blacklistMap)
	}
	return nil
}

func loadBlacklist(path string, blacklist *map[string]bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	newBlacklist := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		newBlacklist[ip] = true
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	blacklistMutex.Lock()
	oldBlacklist := *blacklist
	*blacklist = newBlacklist
	blacklistMutex.Unlock()

	// 打印出被移除的IP地址
	for ip := range oldBlacklist {
		if !newBlacklist[ip] {
			logrus.Infof("At %s, IP %s move out the Blacklist", time.Now().Format(time.RFC3339), ip)
		}
	}

	// 打印出新添加的IP地址
	for ip := range newBlacklist {
		if !oldBlacklist[ip] {
			logrus.Infof("At %s, IP %s add to the Blacklist", time.Now().Format(time.RFC3339), ip)
		}
	}

	return nil
}


func watchBlacklist(path string, blacklist *map[string]bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := loadBlacklist(path, blacklist)
			if err != nil {
				logrus.Errorf("failed to reload blacklist: %s", err.Error())
			}
		}
	}
}

