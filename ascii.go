package modbus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	asciiStart   = ":"
	asciiEnd     = "\r\n"
	asciiMinSize = 3
	asciiMaxSize = 513

	hexTable = "0123456789ABCDEF"
)

// ASCIIPackager implements Packager interface.
type ASCIIPackager struct {
}

// Encode encodes PDU in a ASCII frame:
//  Start           : 1 char
//  Address         : 2 chars
//  Function        : 2 chars
//  Data            : 0 up to 2x252 chars
//  LRC             : 2 chars
//  End             : 2 chars
func (ascii *ASCIIPackager) Encode(slaveID byte, pdu *ProtocolDataUnit) (adu []byte, err error) {
	var buf bytes.Buffer

	if _, err = buf.WriteString(asciiStart); err != nil {
		return
	}
	if err = writeHex(&buf, []byte{slaveID, pdu.FunctionCode}); err != nil {
		return
	}
	if err = writeHex(&buf, pdu.Data); err != nil {
		return
	}
	// Exclude the beginning colon and terminating CRLF pair characters
	var lrc lrc
	lrc.reset()
	lrc.pushByte(slaveID).pushByte(pdu.FunctionCode).pushBytes(pdu.Data)
	if err = writeHex(&buf, []byte{lrc.value()}); err != nil {
		return
	}
	if _, err = buf.WriteString(asciiEnd); err != nil {
		return
	}
	adu = buf.Bytes()
	return
}

// Verify verifies response length, frame boundary and slave id.
func (ascii *ASCIIPackager) Verify(aduRequest []byte, aduResponse []byte) (err error) {
	length := len(aduResponse)
	// Minimum size (including address, function and LRC)
	if length < asciiMinSize+6 {
		err = fmt.Errorf("modbus: response length '%v' does not meet minimum '%v'", length, 9)
		return
	}
	// Length excluding colon must be an even number
	if length%2 != 1 {
		err = fmt.Errorf("modbus: response length '%v' is not an even number", length-1)
		return
	}
	// First char must be a colon
	str := string(aduResponse[0:len(asciiStart)])
	if str != asciiStart {
		err = fmt.Errorf("modbus: response frame '%v'... is not started with '%v'", str, asciiStart)
		return
	}
	// 2 last chars must be \r\n
	str = string(aduResponse[len(aduResponse)-len(asciiEnd):])
	if str != asciiEnd {
		err = fmt.Errorf("modbus: response frame ...'%v' is not ended with '%v'", str, asciiEnd)
		return
	}
	// Slave id
	responseVal, err := readHex(aduResponse[1:])
	if err != nil {
		return
	}
	requestVal, err := readHex(aduRequest[1:])
	if err != nil {
		return
	}
	if responseVal != requestVal {
		err = fmt.Errorf("modbus: response slave id '%v' does not match request '%v'", responseVal, requestVal)
		return
	}
	return
}

// Decode extracts PDU from ASCII frame and verify LRC.
func (ascii *ASCIIPackager) Decode(adu []byte) (pdu *ProtocolDataUnit, err error) {
	pdu = &ProtocolDataUnit{}
	// Slave address
	address, err := readHex(adu[1:])
	if err != nil {
		return
	}
	// Function code
	if pdu.FunctionCode, err = readHex(adu[3:]); err != nil {
		return
	}
	// Data
	dataEnd := len(adu) - 4
	data := adu[5:dataEnd]
	pdu.Data = make([]byte, hex.DecodedLen(len(data)))
	if _, err = hex.Decode(pdu.Data, data); err != nil {
		return
	}
	// LRC
	lrcVal, err := readHex(adu[dataEnd:])
	if err != nil {
		return
	}
	// Calculate checksum
	var lrc lrc
	lrc.reset()
	lrc.pushByte(address).pushByte(pdu.FunctionCode).pushBytes(pdu.Data)
	if lrcVal != lrc.value() {
		err = fmt.Errorf("modbus: response lrc '%v' does not match expected '%v'", lrcVal, lrc.value())
		return
	}
	return
}

func (ascii *ASCIIPackager) transceive(transporter Transporter, logger Logger, aduRequest []byte, timeout time.Duration) (aduResponse []byte, err error) {
	// Make sure port is connected
	if err = transporter.Connect(); err != nil {
		return
	}

	// Set write and read timeout
	if timeout > 0 {
		if err = transporter.SetReadTimeout(timeout); err != nil {
			return
		}
	}

	// Send the request
	log(logger, "modbus: sending %q\n", aduRequest)
	if _, err = transporter.Write(aduRequest); err != nil {
		return
	}

	// Get the response
	var n int
	var data [asciiMaxSize]byte
	length := 0
	for {
		if n, err = transporter.Read(data[length:]); err != nil {
			return
		}
		length += n
		if length >= asciiMaxSize || n == 0 {
			break
		}
		// Expect end of frame in the data received
		if length > asciiMinSize {
			if string(data[length-len(asciiEnd):length]) == asciiEnd {
				break
			}
		}
	}
	aduResponse = data[:length]
	log(logger, "modbus: received %q\n", aduResponse)
	return
}

// writeHex encodes byte to string in hexadecimal, e.g. 0xA5 => "A5"
// (encoding/hex only supports lowercase string).
func writeHex(buf *bytes.Buffer, value []byte) (err error) {
	var str [2]byte
	for _, v := range value {
		str[0] = hexTable[v>>4]
		str[1] = hexTable[v&0x0F]

		if _, err = buf.Write(str[:]); err != nil {
			return
		}
	}
	return
}

// readHex decodes hexa string to byte, e.g. "8C" => 0x8C.
func readHex(data []byte) (value byte, err error) {
	var dst [1]byte
	if _, err = hex.Decode(dst[:], data[0:2]); err != nil {
		return
	}
	value = dst[0]
	return
}
