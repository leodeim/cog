package main

import (
	"fmt"

	"github.com/leonidasdeim/cog"
)

type Config struct {
	Ip   string `default:"localhost"`
	Port string `default:"8080"`
}

func Subscriber(config Config) error {
	fmt.Println("Subscriber has been called")
	fmt.Printf("IP: %s \n", config.Ip)
	fmt.Printf("Port: %s \n", config.Port)

	return nil
}

func main() {
	c, err := cog.Init[Config]()
	if err != nil {
		fmt.Printf("Error at initialize cog: %v", err)
		return
	}
	c.AddSubscriber(Subscriber)

	fmt.Printf("Initial IP: %s \n", c.Config().Ip)
	fmt.Printf("Initial Port: %s \n", c.Config().Port)

	UpdateConfig(c)
}

func UpdateConfig(c *cog.C[Config]) {
	config := c.Config()

	config.Ip = "192.168.1.1"
	config.Port = "8081"

	fmt.Println("Updating config...")
	c.Update(config)
}
