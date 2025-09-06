package main

import (
	"fmt"

	"github.com/stianeikeland/go-rpio"
)

func main() {
	fmt.Println("GPIO Setup")
	err := rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}
	defer rpio.Close()
}
