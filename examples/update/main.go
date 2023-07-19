package main

import (
	"fmt"

	"github.com/leonidasdeim/cog"
)

type Config struct {
	Ip   string `default:"localhost"`
	Port string `default:"8080"`
}

func main() {
	c, err := cog.Init[Config]()
	if err != nil {
		fmt.Printf("Error at initialize cog: %v", err)
		return
	}

	fmt.Printf("Default IP: %s \n", c.GetCfg().Ip)
	fmt.Printf("Default Port: %s \n", c.GetCfg().Port)

	UpdateConfig(c)

	fmt.Printf("Updated IP: %s \n", c.GetCfg().Ip)
	fmt.Printf("Updated Port: %s \n", c.GetCfg().Port)
}

func UpdateConfig(c *cog.Config[Config]) {
	config := c.GetCfg()

	config.Ip = "192.168.1.1"
	config.Port = "8081"

	fmt.Println("Updating config...")
	c.Update(config)
}
