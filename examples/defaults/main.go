package main

import (
	"fmt"

	"github.com/leonidasdeim/goconfig"
)

type Config struct {
	Ip   string `default:"localhost"`
	Port string `default:"8080"`
}

func main() {
	c, err := goconfig.Init[Config]()
	if err != nil {
		fmt.Printf("Error at initialize goconfig: %v", err)
		return
	}

	fmt.Println("There is no initial configuration file.")
	fmt.Println("Initialized config has default values provided with tags:")
	fmt.Printf("Default IP: %s \n", c.GetCfg().Ip)
	fmt.Printf("Default Port: %s \n", c.GetCfg().Port)
}
