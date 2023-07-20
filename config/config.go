package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type InboundConfig struct {
	Tag      string
	Protocol string // 'socks5' or 'tun'
	Listen   string // for socks5, ip addr + port; for tun, tun device name
	Route    string
}

type Addr struct {
	IP   string
	Port uint16
}

type OutboundConfig struct {
	Tag      string
	Protocol string // 'direct' or 'socks5'
	Addrs    []Addr
	Mark     uint32
}

type RouteConfig struct {
	Tag  string
	Outs []string
}

type Config struct {
	Inbound  []InboundConfig
	Outbound []OutboundConfig
	Route    []RouteConfig
}

func GetConfig(filename string) (Config, error) {
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("open config error. ", err.Error())
	}
	configStr, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Println("read config error. ", err.Error())
	}
	var c Config
	err = json.Unmarshal(configStr, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
