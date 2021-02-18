package main

import (
	"log"
	"resock"
)

func main() {
	if err := resock.RunServer(); err != nil {
		log.Fatalln(err)
	}
}
