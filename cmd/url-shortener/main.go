package main

import (
	"AwakeList/internal/config"
	"fmt"
)

func main() {

	cfg := config.MustLoad()

	fmt.Print(cfg)

}
