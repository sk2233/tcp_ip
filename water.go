/*
@author: sk
@date: 2024/6/4
*/
package main

import (
	"fmt"
	"os/exec"

	"github.com/songgao/water"
)

func NewTun(reqIP, respIP string) *water.Interface {
	inst, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	HandleErr(err)
	inst.Name()
	cmd := exec.Command("ifconfig", inst.Name(), reqIP, respIP, "up")
	err = cmd.Run()
	HandleErr(err)
	fmt.Printf("init tun success name %v reqIP %s respIP %s\n", inst.Name(), reqIP, respIP)
	return inst
}
