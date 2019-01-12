package test

import (
	"fmt"
	"testing"

	"github.com/xft/modbus"
)

func TestTCPClient(t *testing.T) {
	cli := modbus.NewTCPClient("127.0.0.1:5022")
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
