package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/xft/modbus"
)

func TestASCIIClient(t *testing.T) {
	cli := modbus.NewASCIIClient("/dev/tty.modbus.ascii")
	cli.SetSlaveID(1)
	cli.SetLogger(logFunc(func(calldepth int, s string) error {
		fmt.Println(s)
		return nil
	}))

	_, err := cli.HoldingRegister(0x100).Read()
	if err != nil {
		t.Fatal(err)
	}

	err = cli.SetSlaveID(2).Coil(0x0001).Toggle()
	if err != nil {
		t.Fatal(err)
	}

	clientTestAll(t, cli)
}

func TestASCIIOverTCPClient(t *testing.T) {
	cli := modbus.NewASCIIOverTCPClient("127.0.0.1:5021", time.Second*5)
	cli.SetSlaveID(1)
	cli.SetLogger(logFunc(func(calldepth int, s string) error {
		fmt.Println(s)
		return nil
	}))

	clientTestAll(t, cli)
}
