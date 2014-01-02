package tun

import (
	"os/exec"
)

func Ifconfig(name, my_ip, remote_ip string) error {
	cmd := exec.Command("ifconfig", name, my_ip, "pointopoint", remote_ip, "up")

	return cmd.Run()
}
