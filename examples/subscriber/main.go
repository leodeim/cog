package main

import (
	"fmt"
	"time"

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

	_ = InitService(c)

	time.Sleep(1 * time.Second)
	updateConfig(c, "192.168.1.1", "8081")
	time.Sleep(1 * time.Second)
	updateConfig(c, "1.1.1.1", "8082")
	time.Sleep(1 * time.Second)
}

func updateConfig(c *cog.Config[Config], ip string, port string) {
	config := c.GetCfg()

	config.Ip = ip
	config.Port = port

	fmt.Println("Updating config...")
	c.Update(config)
}

// Service who uses config and acts as subscriber

type Service struct {
	config *cog.Config[Config]
}

func InitService(c *cog.Config[Config]) *Service {
	s := &Service{
		config: c,
	}

	c.AddSubscriber(s.serviceName())
	go s.configRunner()

	return s
}

func (s *Service) serviceName() string {
	return "SERVICE_1"
}

func (s *Service) connect() {
	fmt.Printf("Service connecting to host: %s:%s \n", s.config.GetCfg().Ip, s.config.GetCfg().Port)
}

func (s *Service) configRunner() {
	for {
		s.connect()

		ch, err := s.config.GetSubscriber(s.serviceName())
		if err != nil {
			fmt.Println("Subscriber not registered error")
			return
		}
		<-ch
		fmt.Println("Subscriber notified, because config has been updated")
	}
}
