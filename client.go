package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"
)

type Logger interface {
	Output(calldepth int, s string) error
}

type ClientHandler struct {
	Packager    Packager
	Transporter Transporter
	SlaveID     byte
	Timeout     time.Duration
	Logger      Logger
	mu          sync.Mutex
}

func (c *ClientHandler) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Transporter.Connect()
}

func (c *ClientHandler) SetLogger(l Logger) {
	c.Logger = l
}

func (c *ClientHandler) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Transporter.Close()
}

func (c *ClientHandler) SetSlaveID(slaveID byte) Client {
	c.SlaveID = slaveID
	return c
}

// Request:
//  Function code         : 1 byte (0x02)
//  Starting address      : 2 bytes
//  Quantity of inputs    : 2 bytes
// Response:
//  Function code         : 1 byte (0x02)
//  Byte count            : 1 byte
//  Input status          : N* bytes (=N or N+1)
func (c *ClientHandler) ReadDiscreteInputs(address, quantity uint16) (inputs []bool, err error) {
	if quantity < 1 || quantity > 2000 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v'", quantity, 1, 2000)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadDiscreteInputs,
		Data:         dataBlock(address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	byteCount := int(response.Data[0])
	length := len(response.Data) - 1
	if byteCount != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, byteCount)
		return
	}

	result := response.Data[1:]
	inputs = make([]bool, byteCount*8)
	for i := 0; i < byteCount; i++ {
		for j := 0; j < 8; j++ {
			inputs[i*8+j] = (result[i] & (1 << uint(j))) != 0
		}
	}
	return inputs[0:quantity], nil
}

// Request:
//  Function code         : 1 byte (0x01)
//  Starting address      : 2 bytes
//  Quantity of coils     : 2 bytes
// Response:
//  Function code         : 1 byte (0x01)
//  Byte count            : 1 byte
//  Coil status           : N* bytes (=N or N+1)
func (c *ClientHandler) ReadCoils(address, quantity uint16) (coils []bool, err error) {
	if quantity < 1 || quantity > 2000 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v'", quantity, 1, 2000)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadCoils,
		Data:         dataBlock(address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	byteCount := int(response.Data[0])
	length := len(response.Data) - 1
	if byteCount != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, byteCount)
		return
	}

	coilStates := response.Data[1:]
	coils = make([]bool, byteCount*8)
	for i := 0; i < byteCount; i++ {
		for j := 0; j < 8; j++ {
			coils[i*8+j] = (coilStates[i] & (1 << uint(j))) != 0
		}
	}
	return coils[0:quantity], nil
}

// Request:
//  Function code         : 1 byte (0x05)
//  Output address        : 2 bytes
//  Output value          : 2 bytes
// Response:
//  Function code         : 1 byte (0x05)
//  Output address        : 2 bytes
//  Output value          : 2 bytes
func (c *ClientHandler) WriteSingleCoil(address uint16, coil bool) (err error) {
	// The requested ON/OFF state can only be 0xFF00 and 0x0000
	var value uint16
	if coil {
		value = 0xFF00
	} else {
		value = 0x0000
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleCoil,
		Data:         dataBlock(address, value),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results := response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if value != respValue {
		err = fmt.Errorf("modbus: response value '%v' does not match request '%v'", respValue, value)
		return
	}
	return
}

// Request:
//  Function code         : 1 byte (0x0F)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
//  Byte count            : 1 byte
//  Outputs value         : N* bytes
// Response:
//  Function code         : 1 byte (0x0F)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
func (c *ClientHandler) WriteMultipleCoils(address uint16, coils []bool) (err error) {
	count := len(coils)
	if count < 1 || count > 1968 {
		err = fmt.Errorf("modbus: quantity '%v' (len(coils)) must be between '%v' and '%v'", count, 1, 1968)
		return
	}
	quantity := uint16(count)

	byteCount := uint((quantity + 7) / 8)
	value := make([]byte, byteCount)
	for i, v := range coils {
		if v {
			value[i>>3] |= 1 << (uint(i) & 7)
		}
	}

	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleCoils,
		Data:         dataBlockSuffix(value, address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results := response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if quantity != respValue {
		err = fmt.Errorf("modbus: response quantity '%v' does not match request '%v'", respValue, quantity)
		return
	}
	return
}

// Request:
//  Function code         : 1 byte (0x03)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
// Response:
//  Function code         : 1 byte (0x03)
//  Byte count            : 1 byte
//  Register value        : Nx2 bytes
func (c *ClientHandler) ReadHoldingRegisters(address, quantity uint16) (values []uint16, err error) {
	if quantity < 1 || quantity > 125 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v'", quantity, 1, 125)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadHoldingRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	values = bytesToWordArray(response.Data[1:])
	return
}

// Request:
//  Function code         : 1 byte (0x04)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
// Response:
//  Function code         : 1 byte (0x04)
//  Byte count            : 1 byte
//  Input registers       : N bytes
func (c *ClientHandler) ReadInputRegisters(address, quantity uint16) (values []uint16, err error) {
	if quantity < 1 || quantity > 125 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v'", quantity, 1, 125)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadInputRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	values = bytesToWordArray(response.Data[1:])
	return
}

// Request:
//  Function code         : 1 byte (0x06)
//  Register address      : 2 bytes
//  Register value        : 2 bytes
// Response:
//  Function code         : 1 byte (0x06)
//  Register address      : 2 bytes
//  Register value        : 2 bytes
func (c *ClientHandler) WriteSingleRegister(address, value uint16) (err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleRegister,
		Data:         dataBlock(address, value),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results := response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if value != respValue {
		err = fmt.Errorf("modbus: response value '%v' does not match request '%v'", respValue, value)
		return
	}
	return
}

// Request:
//  Function code         : 1 byte (0x10)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
//  Byte count            : 1 byte
//  Registers value       : N* bytes
// Response:
//  Function code         : 1 byte (0x10)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
func (c *ClientHandler) WriteMultipleRegisters(address uint16, values []uint16) (err error) {
	count := len(values)
	if count < 1 || count > 123 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v'", count, 1, 123)
		return
	}
	quantity := uint16(count)

	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleRegisters,
		Data:         dataBlockSuffix(dataBlock(values...), address, quantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results := response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if quantity != respValue {
		err = fmt.Errorf("modbus: response quantity '%v' does not match request '%v'", respValue, quantity)
		return
	}
	return
}

// Request:
//  Function code         : 1 byte (0x16)
//  Reference address     : 2 bytes
//  AND-mask              : 2 bytes
//  OR-mask               : 2 bytes
// Response:
//  Function code         : 1 byte (0x16)
//  Reference address     : 2 bytes
//  AND-mask              : 2 bytes
//  OR-mask               : 2 bytes
func (c *ClientHandler) MaskWriteRegister(address, andMask, orMask uint16) (err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeMaskWriteRegister,
		Data:         dataBlock(address, andMask, orMask),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 6 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 6)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	respValue = binary.BigEndian.Uint16(response.Data[2:])
	if andMask != respValue {
		err = fmt.Errorf("modbus: response AND-mask '%v' does not match request '%v'", respValue, andMask)
		return
	}
	respValue = binary.BigEndian.Uint16(response.Data[4:])
	if orMask != respValue {
		err = fmt.Errorf("modbus: response OR-mask '%v' does not match request '%v'", respValue, orMask)
		return
	}
	return
}

// Request:
//  Function code         : 1 byte (0x17)
//  Read starting address : 2 bytes
//  Quantity to read      : 2 bytes
//  Write starting address: 2 bytes
//  Quantity to write     : 2 bytes
//  Write byte count      : 1 byte
//  Write registers value : N* bytes
// Response:
//  Function code         : 1 byte (0x17)
//  Byte count            : 1 byte
//  Read registers value  : Nx2 bytes
func (c *ClientHandler) ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (values []uint16, err error) {
	if readQuantity < 1 || readQuantity > 125 {
		err = fmt.Errorf("modbus: quantity to read '%v' must be between '%v' and '%v'", readQuantity, 1, 125)
		return
	}
	if writeQuantity < 1 || writeQuantity > 121 {
		err = fmt.Errorf("modbus: quantity to write '%v' must be between '%v' and '%v'", writeQuantity, 1, 121)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadWriteMultipleRegisters,
		Data:         dataBlockSuffix(value, readAddress, readQuantity, writeAddress, writeQuantity),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", len(response.Data)-1, count)
		return
	}
	values = bytesToWordArray(response.Data[1:])
	return
}

// Request:
//  Function code         : 1 byte (0x18)
//  FIFO pointer address  : 2 bytes
// Response:
//  Function code         : 1 byte (0x18)
//  Byte count            : 2 bytes
//  FIFO count            : 2 bytes
//  FIFO count            : 2 bytes (<=31)
//  FIFO value register   : Nx2 bytes
func (c *ClientHandler) ReadFIFOQueue(address uint16) (values []uint16, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadFIFOQueue,
		Data:         dataBlock(address),
	}
	response, err := c.transceive(&request)
	if err != nil {
		return
	}
	if len(response.Data) < 4 {
		err = fmt.Errorf("modbus: response data size '%v' is less than expected '%v'", len(response.Data), 4)
		return
	}
	count := int(binary.BigEndian.Uint16(response.Data))
	if count != (len(response.Data) - 1) {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", len(response.Data)-1, count)
		return
	}
	count = int(binary.BigEndian.Uint16(response.Data[2:]))
	if count > 31 {
		err = fmt.Errorf("modbus: fifo count '%v' is greater than expected '%v'", count, 31)
		return
	}
	values = bytesToWordArray(response.Data[4:])
	return
}

type ioStyle struct {
	master  Client
	address uint16
	count   uint16
}

type roBit ioStyle
type rwBit ioStyle

type roRegister ioStyle
type roRegisters ioStyle
type rwRegister ioStyle
type rwRegisters ioStyle

func (io *roBit) Test() (result bool, err error) {
	res, err := io.master.ReadDiscreteInputs(io.address, 1)
	if err == nil {
		result = res[0]
	}
	return
}

func (io *rwBit) Test() (result bool, err error) {
	res, err := io.master.ReadCoils(io.address, 1)
	if err == nil {
		result = res[0]
	}
	return
}

func (io *rwBit) Set() (err error) {
	return io.master.WriteSingleCoil(io.address, true)
}

func (io *rwBit) Clear() (err error) {
	return io.master.WriteSingleCoil(io.address, false)
}

func (io *rwBit) Toggle() (err error) {
	state, err := io.Test()
	if err != nil {
		return
	}
	if state {
		return io.Clear()
	}
	return io.Set()
}

func (io *roRegister) Read() (value uint16, err error) {
	res, err := io.master.ReadInputRegisters(io.address, 1)
	if err == nil {
		value = res[0]
	}
	return
}

func (io *rwRegister) Read() (value uint16, err error) {
	res, err := io.master.ReadHoldingRegisters(io.address, 1)
	if err == nil {
		value = res[0]
	}
	return
}

func (io *rwRegister) Write(value uint16) (err error) {
	return io.master.WriteSingleRegister(io.address, value)
}

func (io *roRegisters) Read() (values []uint16, err error) {
	return io.master.ReadInputRegisters(io.address, io.count)
}
func (io *roRegisters) ReadString() (s string, err error) {
	words, err := io.Read()
	if err != nil {
		return
	}
	s = string(filterNullChar(wordsToByteArray(words)))
	return
}

func (io *rwRegisters) Read() (values []uint16, err error) {
	return io.master.ReadHoldingRegisters(io.address, io.count)
}

func (io *rwRegisters) Write(values []uint16) (err error) {
	if l := len(values); l > int(io.count) {
		return fmt.Errorf("invalid length of words %d", l)
	}
	return io.master.WriteMultipleRegisters(io.address, values)
}

func (c *ClientHandler) DiscreteInput(addr uint16) DiscreteInput {
	return &roBit{master: c, address: addr}
}

func (c *ClientHandler) Coil(addr uint16) Coil {
	return &rwBit{master: c, address: addr}
}

func (c *ClientHandler) InputRegister(addr uint16) InputRegister {
	return &roRegister{master: c, address: addr}
}

func (c *ClientHandler) InputRegisters(addr, count uint16) InputRegisters {
	return &roRegisters{c, addr, count}
}

func (c *ClientHandler) HoldingRegister(addr uint16) HoldingRegister {
	return &rwRegister{master: c, address: addr}
}

func (c *ClientHandler) HoldingRegisters(addr, count uint16) HoldingRegisters {
	return &rwRegisters{c, addr, count}
}

func (io *rwRegisters) ReadString() (s string, err error) {
	words, err := io.Read()
	if err != nil {
		return
	}
	s = string(filterNullChar(wordsToByteArray(words)))
	return
}

func (io *rwRegisters) WriteString(s string) (err error) {
	return io.Write(bytesToWordArray([]byte(s)))
}

// send sends request and checks possible exception in the response.
func (c *ClientHandler) transceive(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	aduRequest, err := c.Packager.Encode(c.SlaveID, request)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	aduResponse, err := c.Packager.transceive(c.Transporter, c.Logger, aduRequest, c.Timeout)
	if err != nil {
		return
	}

	if err = c.Packager.Verify(aduRequest, aduResponse); err != nil {
		return
	}
	response, err = c.Packager.Decode(aduResponse)
	if err != nil {
		return
	}
	// Check correct function code returned (exception)
	if response.FunctionCode != request.FunctionCode {
		err = responseError(response)
		return
	}
	if response.Data == nil || len(response.Data) == 0 {
		// Empty response
		err = fmt.Errorf("modbus: response data is empty")
		return
	}
	return
}

// dataBlock creates a sequence of uint16 data.
func dataBlock(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}

// dataBlockSuffix creates a sequence of uint16 data and append the suffix plus its length.
func dataBlockSuffix(suffix []byte, value ...uint16) []byte {
	length := 2 * len(value)
	data := make([]byte, length+1+len(suffix))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	data[length] = uint8(len(suffix))
	copy(data[length+1:], suffix)
	return data
}

func responseError(response *ProtocolDataUnit) error {
	mbError := &ModbusError{FunctionCode: response.FunctionCode}
	if response.Data != nil && len(response.Data) > 0 {
		mbError.ExceptionCode = response.Data[0]
	}
	return mbError
}

func log(logger Logger, format string, v ...interface{}) {
	if logger != nil {
		logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func bytesToWordArray(bytes []byte) []uint16 {
	l := len(bytes)
	n := int(math.Ceil(float64(l) / 2))
	array := make([]uint16, n)
	for i := 0; i < n; i++ {
		j := i * 2
		if j+2 > l {
			array[i] = uint16(bytes[j])
		} else {
			array[i] = binary.BigEndian.Uint16(bytes[j : j+2])
		}
	}
	return array
}

func wordsToByteArray(words []uint16) []byte {
	array := make([]byte, 2*len(words))
	for i, v := range words {
		binary.BigEndian.PutUint16(array[i*2:], v)
	}
	return array
}

func filterNullChar(a []byte) []byte {
	x := filter(a, func(c byte) bool {
		return c != 0
	})
	return x
}

func filter(s []byte, fn func(byte) bool) []byte {
	var p []byte // == nil
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}
