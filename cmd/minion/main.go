package main

import (
	"flag"
	"log"

	"github.com/gtracer/overlord/pkg/boot"
)

func main() {
	customerName := flag.String("customer", "customer", "customer name")
	clusterName := flag.String("cluster", "overlord", "cluster name")
	flag.Parse()
	err := boot.Boot(*customerName, *clusterName)
	if err != nil {
		log.Print("error occured")
	}
}
