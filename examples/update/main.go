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

	fmt.Printf("Default IP: %s \n", c.Config().Ip)
	fmt.Printf("Default Port: %s \n", c.Config().Port)

	UpdateConfig(c)

	fmt.Printf("Updated IP: %s \n", c.Config().Ip)
	fmt.Printf("Updated Port: %s \n", c.Config().Port)
}

func UpdateConfig(c *cog.C[Config]) {
	config := c.Config()

	config.Ip = "192.168.1.1"
	config.Port = "8081"

	fmt.Println("Updating config...")
	c.Update(config)
}
