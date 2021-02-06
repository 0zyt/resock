package main

import (
	"flag"
	"sync"
)

type sockConfig struct {
	remoteAddr string
	localAddr  string
}

var (
	cfg *sockConfig
	so  sync.Once
)

func getConfig() *sockConfig {
	so.Do(func() {
		local := flag.String("local", "127.0.0.1:1080", "please enter local proxy ip")
		remote := flag.String("server", "127.0.0.1:2001", "please enter tcp tunnel server ip")
		cfg = &sockConfig{*remote, *local}
	})
	return cfg
}
