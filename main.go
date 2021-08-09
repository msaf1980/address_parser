package main

import (
	"flag"
	"log"
)

func main() {
	var path = flag.String("file", "address.xml", "XML file with addresses")
	flag.Parse()

	var parser AddressParser
	parser.Init()

	if err := parser.readAddressFile(*path); err != nil {
		log.Fatalf("read error: %v", err)
	} else {
		parser.Stat.Dump()
	}
}
