package main

import (
	"AwakeList"
	"log"
)

func main() {
	srv := new(AwakeList.Server)
	if err := srv.Run("8080"); err != nil {
		log.Fatal("Error while running server: $s", err.Error())
	}
}
