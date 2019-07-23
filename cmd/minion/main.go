package main

import (
	"log"

	"github.com/gtracer/overlord/pkg/boot"
)

func main() {
	err := boot.Boot()
	if err != nil {
		log.Print("error occured")
	}
}
