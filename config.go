package main

type sockConfig struct {
	remoteAddr string
	localAddr  string
}

func NewsockConfig(remote, local string) *sockConfig {
	return &sockConfig{remote, local}
}
