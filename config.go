package main

import (
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type configStruct struct {
	Name   string `json:"name"`
	Listen string `json:"listen"`
	Net    string `json:"net"`
	Mode   int    `json:"mode"`
	Targets []*struct {
		Regexp  string         `json:"regexp"`
		regexp  *regexp.Regexp `json:"-"`
		Address string         `json:"address"`
	} `json:"targets"`
}

var config []*configStruct

func init() {
	buf, err := ioutil.ReadFile("config.json")
	if err != nil {
		logrus.Fatalf("failed to load config.json: %s", err.Error())
	}

	if err := json.Unmarshal(buf, &config); err != nil {
		logrus.Fatalf("failed to load config.json: %s", err.Error())
	}

	if len(config) == 0 {
		logrus.Fatalf("empty config", err.Error())
	}

	for i, v := range config {
		v.Net = strings.ToLower(v.Net)
		if err := v.verify(); err != nil {
			logrus.Fatalf("verity config failed at pos %d : %s", i, err.Error())
		}
	}
}

func (c *configStruct) verify() error {
	if c.Name == "" {
		return fmt.Errorf("empty name")
	}
	if c.Listen == "" {
		return fmt.Errorf("invalid listen address")
	}
	if c.Net != "tcp" && c.Net != "udp" {
		return fmt.Errorf("invalid network protocol")
	}
	if c.Mode != MODE_NORMAL &&
		c.Mode != MODE_REGEXP_SERVER_FIRST &&
		c.Mode != MODE_REGEXP_CLIENT_FIRST {
		return fmt.Errorf("invalid mode %d", c.Mode)
	}
	if len(c.Targets) == 0 {
		return fmt.Errorf("invalid targets")
	}
	for i, v := range c.Targets {
		if v.Address == "" {
			return fmt.Errorf("invalid address at pos %d", i)
		}
		if c.Mode == MODE_REGEXP_CLIENT_FIRST || c.Mode == MODE_REGEXP_SERVER_FIRST {
			r, err := regexp.Compile(v.Regexp)
			if err != nil {
				return fmt.Errorf("invalid regexp at pos %d : %s", i, err.Error())
			}
			v.regexp = r
		}
	}
	return nil
}
