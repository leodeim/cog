package main

import (
	"errors"
	"fmt"

	"github.com/leonidasdeim/cog"
)

type Config struct {
	Name string `default:"my-app"`
}

func Subscriber1(config Config) error {
	fmt.Println("Subscriber1 has been called, new app name: ", config.Name)

	return nil
}

func Subscriber2(config Config) error {
	fmt.Println("Subscriber2 has been called, new app name: ", config.Name)

	return nil
}

func SubscriberBad(config Config) error {
	fmt.Println("Subscriber1Bad returns an error")
	fmt.Println("It triggers rollback")

	return errors.New("some error")
}

func main() {
	c, err := cog.Init[Config]()
	if err != nil {
		fmt.Printf("Error at initialize cog: %v", err)
		return
	}

	fmt.Println("Adding subscribers one by one")
	c.AddSubscriber(Subscriber1)
	c.AddSubscriber(Subscriber2)
	c.AddSubscriber(SubscriberBad)

	fmt.Println("Updating config, it will trigger subscribers")
	c.Update(Config{
		Name: "new-name",
	})
}
