package main

import (
	"fmt"
	"os"

	"github.com/backendsystems/nibble/internal/parameters"
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/tui"
)

var version = "dev"

func main() {
	p := parameters.Parse(version)

	switch p.Mode {
	case parameters.ModeVersion:
		fmt.Println(version)

	case parameters.ModeHeadless:
		parameters.RunHeadless(p)

	case parameters.ModeTUI:
		s := scanner.New(p.DemoMode)

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
}
