package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/tui"
)

var version = "dev"

func main() {
	var demoMode bool
	var showVersion bool
	flag.BoolVar(&demoMode, "demo", false, "use demo interfaces")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return
	}

	s := scanner.New(demoMode)

	ifaces, addrsByIface, err := s.GetInterfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		os.Exit(1)
	}

	if len(ifaces) == 0 {
		fmt.Println("No valid network interfaces found with IPv4 addresses")
		os.Exit(1)
	}

	if err := tui.Run(s, ifaces, addrsByIface); err != nil {
		fmt.Printf("Error starting the program: %v", err)
		os.Exit(1)
	}
}
