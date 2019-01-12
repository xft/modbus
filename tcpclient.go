package modbus

import (
	"net"
	"time"
)

func NewTCPClient(address string) *ClientHandler {
	return &ClientHandler{
		Packager:    &TCPPackager{},
		Transporter: NewTCPAddrTransport(address, time.Second*10),
		Timeout:     time.Second * 10,
	}
}

func NewTCPClient2(conn net.Conn) *ClientHandler {
	return &ClientHandler{
		Packager:    &TCPPackager{},
		Transporter: NewTCPConnTransport(conn),
		Timeout:     time.Second * 10,
	}
}
