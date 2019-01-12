package modbus

import (
	"net"
	"time"
)

func NewASCIIClient(device string) *ClientHandler {
	return NewASCIIClient2(device, 115200, 8, "N", 1, time.Second*5)
}

func NewASCIIClient2(device string, baudRate int, dataBits int, parity string, stopBits int, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &ASCIIPackager{},
		Transporter: NewSerialTransport(device, baudRate, dataBits, parity, stopBits, readTimeout),
		Timeout:     readTimeout,
	}
}

func NewASCIIOverTCPClient(address string, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &ASCIIPackager{},
		Transporter: NewTCPAddrTransport(address, readTimeout),
		Timeout:     readTimeout,
	}
}

func NewASCIIOverTCPClient2(conn net.Conn, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &ASCIIPackager{},
		Transporter: NewTCPConnTransport(conn),
		Timeout:     readTimeout,
	}
}
