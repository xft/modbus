package test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/xft/modbus"
)

func assertEquals(t *testing.T, expected, actual interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		// Get file name only
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)", file, line,
			expected, expected, actual, actual)
		t.FailNow()
	}
}

func clientTestReadCoils(t *testing.T, client modbus.Client) {
	// Read discrete outputs 20-38:
	address := uint16(0x0013)
	quantity := uint16(0x0013)
	results, err := client.ReadCoils(address, quantity)
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(t, int(quantity), len(results))
}

func clientTestReadDiscreteInputs(t *testing.T, client modbus.Client) {
	// Read discrete inputs 197-218
	address := uint16(0x00C4)
	quantity := uint16(0x0016)
	results, err := client.ReadDiscreteInputs(address, quantity)
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(t, int(quantity), len(results))
}

func clientTestReadHoldingRegisters(t *testing.T, client modbus.Client) {
	// Read registers 108-110
	address := uint16(0x006B)
	quantity := uint16(0x0003)
	results, err := client.ReadHoldingRegisters(address, quantity)
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(t, int(quantity), len(results))
}

func clientTestReadInputRegisters(t *testing.T, client modbus.Client) {
	// Read input register 9
	address := uint16(0x0008)
	quantity := uint16(0x0001)
	results, err := client.ReadInputRegisters(address, quantity)
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(t, int(quantity), len(results))
}

func clientTestWriteSingleCoil(t *testing.T, client modbus.Client) {
	// Write coil 173 ON
	address := uint16(0x00AC)
	err := client.WriteSingleCoil(address, true)
	if err != nil {
		t.Fatal(err)
	}
}

func clientTestWriteSingleRegister(t *testing.T, client modbus.Client) {
	// Write register 2 to 00 03 hex
	address := uint16(0x0001)
	value := uint16(0x0003)
	err := client.WriteSingleRegister(address, value)
	if err != nil {
		t.Fatal(err)
	}
}

func clientTestWriteMultipleCoils(t *testing.T, client modbus.Client) {
	// Write a series of 10 coils starting at coil 20
	address := uint16(0x0013)
	values := []bool{true, true, false, false, true}
	err := client.WriteMultipleCoils(address, values)
	if err != nil {
		t.Fatal(err)
	}
}

func clientTestWriteMultipleRegisters(t *testing.T, client modbus.Client) {
	// Write two registers starting at 2 to 00 0A and 01 02 hex
	address := uint16(0x0001)
	values := []uint16{0x000A, 0x0102}
	err := client.WriteMultipleRegisters(address, values)
	if err != nil {
		t.Fatal(err)
	}
}

func clientTestMaskWriteRegisters(t *testing.T, client modbus.Client) {
	// Mask write to register 5
	address := uint16(0x0004)
	andMask := uint16(0x00F2)
	orMask := uint16(0x0025)
	err := client.MaskWriteRegister(address, andMask, orMask)
	if err != nil {
		t.Fatal(err)
	}
}

func clientTestReadWriteMultipleRegisters(t *testing.T, client modbus.Client) {
	// read six registers starting at register 4, and to write three registers starting at register 15
	address := uint16(0x0003)
	quantity := uint16(0x0006)
	writeAddress := uint16(0x000E)
	writeQuantity := uint16(0x0003)
	values := []byte{0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF}
	results, err := client.ReadWriteMultipleRegisters(address, quantity, writeAddress, writeQuantity, values)
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(t, int(quantity), len(results))
}

func clientTestReadFIFOQueue(t *testing.T, client modbus.Client) {
	// Read queue starting at the pointer register 1246
	address := uint16(0x04DE)
	results, err := client.ReadFIFOQueue(address)
	// Server not implemented
	if err != nil {
		assertEquals(t, "modbus: exception '1' (illegal function), function '152'", err.Error())
	} else {
		assertEquals(t, 0, len(results))
	}
}

func clientTestAll(t *testing.T, client modbus.Client) {
	clientTestReadCoils(t, client)
	clientTestReadDiscreteInputs(t, client)
	clientTestReadHoldingRegisters(t, client)
	clientTestReadInputRegisters(t, client)
	clientTestWriteSingleCoil(t, client)
	clientTestWriteSingleRegister(t, client)
	clientTestWriteMultipleCoils(t, client)
	clientTestWriteMultipleRegisters(t, client)
	clientTestMaskWriteRegisters(t, client)
	clientTestReadWriteMultipleRegisters(t, client)
	clientTestReadFIFOQueue(t, client)
}
