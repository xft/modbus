package modbus

import (
	"net"
	"time"
)

func NewRTUClient(device string) *ClientHandler {
	return NewRTUClient2(device, 115200, 8, "N", 1, time.Second*5)
}

func NewRTUClient2(device string, baudRate int, dataBits int, parity string, stopBits int, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &RTUPackager{},
		Transporter: NewSerialTransport(device, baudRate, dataBits, parity, stopBits, readTimeout),
		Timeout:     readTimeout,
	}
}

func NewRTUOverTCPClient(address string, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &RTUPackager{},
		Transporter: NewTCPAddrTransport(address, readTimeout),
		Timeout:     readTimeout,
	}
}

func NewRTUOverTCPClient2(conn net.Conn, readTimeout time.Duration) *ClientHandler {
	return &ClientHandler{
		Packager:    &RTUPackager{},
		Transporter: NewTCPConnTransport(conn),
		Timeout:     readTimeout,
	}
}
