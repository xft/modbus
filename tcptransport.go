package modbus

import (
	"errors"
	"net"
	"time"
)

type tcpConnCategoryPort struct {
	conn net.Conn
}

func NewTCPConnTransport(conn net.Conn) *tcpConnCategoryPort {
	return &tcpConnCategoryPort{
		conn: conn,
	}
}

func (tcp *tcpConnCategoryPort) Connect() (err error) {
	if tcp.conn == nil {
		err = errors.New("connection was closed, not support to reconnect")
	}
	return err
}

func (tcp *tcpConnCategoryPort) Read(b []byte) (n int, err error) {
	return tcp.conn.Read(b)
}

func (tcp *tcpConnCategoryPort) Write(b []byte) (n int, err error) {
	return tcp.conn.Write(b)
}

func (tcp *tcpConnCategoryPort) Close() (err error) {
	if tcp.conn != nil {
		err = tcp.conn.Close()
		tcp.conn = nil
	}
	return err
}

func (tcp *tcpConnCategoryPort) SetReadTimeout(timeout time.Duration) (err error) {
	return tcp.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (tcp *tcpConnCategoryPort) Flush() (err error) {
	var n int
	b := make([]byte, 1024)
	for {
		if err = tcp.conn.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
			return
		}
		n, err = tcp.conn.Read(b)
		if err != nil {
			if netError, ok := err.(net.Error); ok && netError.Timeout() {
				err = nil
			}
			break
		} else if n == 0 {
			break
		}
	}
	return
}

type tcpAddrCategoryPort struct {
	address        string
	connectTimeout time.Duration
	tcpConnCategoryPort
}

func NewTCPAddrTransport(addr string, connectTimeout time.Duration) *tcpAddrCategoryPort {
	return &tcpAddrCategoryPort{
		address:        addr,
		connectTimeout: connectTimeout,
	}
}

func (tcp *tcpAddrCategoryPort) Connect() (err error) {
	if tcp.conn == nil {
		dialer := net.Dialer{Timeout: tcp.connectTimeout}
		tcp.conn, err = dialer.Dial("tcp", tcp.address)
	}
	return err
}
