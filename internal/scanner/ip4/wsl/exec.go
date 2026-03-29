package wsl

import (
	"bytes"
	"os/exec"
)

func runWinCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type cidrAddr struct {
	cidr string
}

func (c cidrAddr) Network() string { return "ip+net" }
func (c cidrAddr) String() string  { return c.cidr }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [3]byte{}
	pos := 2
	for n > 0 {
		buf[pos] = byte('0' + n%10)
		n /= 10
		pos--
	}
	return string(buf[pos+1:])
}
