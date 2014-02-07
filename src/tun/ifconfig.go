package tun

import (
	"fmt"
	"os/exec"
)

func Ifconfig(name, my_ip, remote_ip string) error {
	cmd := exec.Command("ifconfig", name, my_ip, "pointopoint", remote_ip, "up")

	return cmd.Run()
}

func SetMTU(name string, mtu int) error {
	cmd := exec.Command("ifconfig", name, "mtu", fmt.Sprintf("%d", mtu))

	return cmd.Run()
}
