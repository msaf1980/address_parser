package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	var path = flag.String("file", "address.xml", "XML file with addresses")
	flag.Parse()

	var parser AddressParser
	parser.Init(1)

	start := time.Now()
	if err := parser.readAddressFile(*path); err != nil {
		log.Fatalf("read error: %v", err)
	} else {
		duration := time.Now().Sub(start)
		parser.Stat.Dump()
		log.Printf("Load complete in %v", duration)
	}
}
