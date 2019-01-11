package modbus

import (
	"io"
	"time"
)

// Transporter specifies the transport layer.
type Transporter interface {
	Connect() error
	io.ReadWriteCloser
	SetReadTimeout(timeout time.Duration) error
	Flush() error
}
