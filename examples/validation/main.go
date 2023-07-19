package main

import (
	"fmt"

	"github.com/leonidasdeim/cog"
	fh "github.com/leonidasdeim/cog/pkg/filehandler"
)

type Config struct {
	Ip       string `default:"localhost"`
	Port     string `default:"8080"`
	Username string `validate:"required"`
	Password string `validate:"required"`
}

func main() {
	h1, _ := fh.New(fh.WithName("file1"))
	_, err := cog.Init[Config](h1)
	if err == nil {
		fmt.Println("Config with 'file1' successfully initialized.")
	}

	h2, _ := fh.New(fh.WithName("file2"))
	_, err = cog.Init[Config](h2)
	if err != nil {
		fmt.Println("Validation failed for config using 'file2':")
		fmt.Println(err)
	}
}
