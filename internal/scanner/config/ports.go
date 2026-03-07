package config

import (
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

func SetPorts(scanner shared.Scanner, ports []int) {
	switch typed := scanner.(type) {
	case *ip4.Scanner:
		typed.Ports = ports
	case *demo.Scanner:
		typed.Ports = ports
	}
}

func WithPorts(base shared.Scanner, ports []int) shared.Scanner {
	switch base.(type) {
	case *ip4.Scanner:
		return &ip4.Scanner{Ports: ports}
	case *demo.Scanner:
		return &demo.Scanner{Ports: ports}
	default:
		return base
	}
}
