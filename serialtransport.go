package modbus

import (
	"fmt"
	"io"
	"time"

	"github.com/xft/serial"
)

type serialPort struct {
	// Serial port configuration.
	serial.Config

	port *serial.Port
}

func NewSerialTransport(device string, baudRate int, dataBits int, parity string, stopBits int, timeout time.Duration) *serialPort {
	port := &serialPort{
		Config: serial.Config{
			Device:      device,
			BaudRate:    baudRate,
			ReadTimeout: timeout,
		},
	}
	port.DataBits = byte(dataBits)
	if len(parity) == 0 {
		port.Parity = serial.ParityNone
	} else {
		port.Parity = serial.Parity(parity[0])
	}
	port.StopBits = serial.StopBits(stopBits)
	return port
}

func (s *serialPort) Connect() (err error) {
	if s.port == nil {
		s.port, err = serial.Open(&s.Config)
	}
	return err
}

type serialReadTimeoutErr struct {
	device string
	errStr string
}

func (e *serialReadTimeoutErr) Error() string {
	return fmt.Sprintf("read serial %s: %s", e.device, e.errStr)
}

func (e *serialReadTimeoutErr) Timeout() bool {
	return true
}

func (s *serialPort) Read(b []byte) (n int, err error) {
	n, err = s.port.Read(b)
	if s.ReadTimeout > 0 && ((err == nil && n == 0) || (err != nil && err == io.EOF)) {
		err = &serialReadTimeoutErr{device: s.Device, errStr: "i/o timeout"}
	}
	return
}

func (s *serialPort) Write(b []byte) (n int, err error) {
	return s.port.Write(b)
}

func (s *serialPort) Close() (err error) {
	if s.port != nil {
		err = s.port.Close()
		s.port = nil
	}
	return err
}

func (s *serialPort) SetReadTimeout(timeout time.Duration) (err error) {
	if s.ReadTimeout != timeout {
		if s.port != nil {
			err = s.Close()
			if err != nil {
				return
			}
			saved := s.ReadTimeout
			s.ReadTimeout = timeout
			err = s.Connect()
			if err != nil {
				s.ReadTimeout = saved
			}
		}
	}
	return nil
}

func (s *serialPort) Flush() (err error) {
	return s.port.Flush()
}
