package main

import (
	"log"
	"resock"
)

func main() {
	if err := resock.RunClient(); err != nil {
		log.Fatalln(err)
	}
}
