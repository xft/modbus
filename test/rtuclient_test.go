package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/xft/modbus"
)

type logFunc func(calldepth int, s string) error

func (l logFunc) Output(calldepth int, s string) error {
	return l(calldepth, s)
}

func TestRTUClient(t *testing.T) {
	cli := modbus.NewRTUClient("/dev/tty.modbus.rtu")
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

func TestRTUOverTCPClient(t *testing.T) {
	cli := modbus.NewRTUOverTCPClient("127.0.0.1:5020", time.Second*5)
	cli.SetSlaveID(1)
	cli.SetLogger(logFunc(func(calldepth int, s string) error {
		fmt.Println(s)
		return nil
	}))

	clientTestAll(t, cli)
}
