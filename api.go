package modbus

type Client interface {
	// SetLogger assigns the logger to use
	//
	// The logger parameter is an interface that requires the following
	// method to be implemented (such as the the stdlib log.Logger):
	//
	//    Output(calldepth int, s string)
	//
	SetLogger(logger Logger)

	Close() error

	SetSlaveID(slaveID byte) Client

	// Bit access

	// ReadDiscreteInputs reads from 1 to 2000 contiguous status of
	// discrete inputs in a remote device and returns input status.
	// Function Code 2
	ReadDiscreteInputs(address, quantity uint16) (inputs []bool, err error)
	// ReadCoils reads from 1 to 2000 contiguous status of coils in a
	// remote device and returns coil status.
	// Function Code 1
	ReadCoils(address, quantity uint16) (coils []bool, err error)
	// WriteSingleCoil write a single output to either ON or OFF in a
	// remote device and returns output value.
	// Function Code 5
	WriteSingleCoil(address uint16, coil bool) (err error)
	// WriteMultipleCoils forces each coil in a sequence of coils to either
	// ON or OFF in a remote device and returns quantity of outputs.
	// Function Code 15
	WriteMultipleCoils(address uint16, coils []bool) (err error)

	// 16-bit access

	// ReadHoldingRegisters reads the contents of a contiguous block of
	// holding registers in a remote device and returns register value.
	// Function Code 3
	ReadHoldingRegisters(address, quantity uint16) (readRegisters []uint16, err error)
	// ReadInputRegisters reads from 1 to 125 contiguous input registers in
	// a remote device and returns input registers.
	// Function Code 4
	ReadInputRegisters(address, quantity uint16) (readRegisters []uint16, err error)
	// WriteSingleRegister writes a single holding register in a remote
	// device and returns register value.
	// Function Code 6
	WriteSingleRegister(address, value uint16) (err error)
	// WriteMultipleRegisters writes a block of contiguous registers
	// (1 to 123 registers) in a remote device and returns quantity of registers.
	// Function Code 23
	WriteMultipleRegisters(address uint16, values []uint16) (err error)
	// ReadWriteMultipleRegisters performs a combination of one read
	// operation and one write operation. It returns read registers value.
	// Function Code 23
	ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (readRegisters []uint16, err error)
	// MaskWriteRegister modify the contents of a specified holding
	// register using a combination of an AND mask, an OR mask, and the
	// register's current contents. The function returns
	// AND-mask and OR-mask.
	// Function Code 22
	MaskWriteRegister(address, andMask, orMask uint16) (err error)
	//ReadFIFOQueue reads the contents of a First-In-First-Out (FIFO) queue
	// of register in a remote device and returns FIFO value register.
	// Function Code 24
	ReadFIFOQueue(address uint16) (fifoValues []uint16, err error)

	// Abstract Objects

	// Discrete input
	DiscreteInput(address uint16) DiscreteInput
	// Coil
	Coil(address uint16) Coil
	// Input Register
	InputRegister(address uint16) InputRegister
	// Input Registers
	InputRegisters(address, count uint16) InputRegisters
	// Holding Register
	HoldingRegister(address uint16) HoldingRegister
	// Holding Registers
	HoldingRegisters(address, count uint16) HoldingRegisters
}

type DiscreteInput interface {
	Test() (bool, error)
}

type Coil interface {
	DiscreteInput
	Set() error
	Clear() error
	Toggle() error
}

type InputRegister interface {
	Read() (uint16, error)
}

type InputRegisters interface {
	Read() ([]uint16, error)
	ReadString() (string, error)
}

type HoldingRegister interface {
	InputRegister
	Write(uint16) error
}

type HoldingRegisters interface {
	InputRegisters
	Write([]uint16) error
	WriteString(s string) error
}
