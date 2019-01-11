go modbus [![Build Status](https://travis-ci.org/xft/modbus.svg?branch=master)](https://travis-ci.org/xft/modbus) [![GoDoc](https://godoc.org/github.com/xft/modbus?status.svg)](https://godoc.org/github.com/xft/modbus)
=========
Fault-tolerant, fail-fast implementation of Modbus protocol in Go.

Supported functions
-------------------
Bit access:
*   Read Discrete Inputs
*   Read Coils
*   Write Single Coil
*   Write Multiple Coils

16-bit access:
*   Read Input Registers
*   Read Holding Registers
*   Write Single Register
*   Write Multiple Registers
*   Read/Write Multiple Registers
*   Mask Write Register
*   Read FIFO Queue

Supported formats
-----------------
*   Serial (RTU)

Usage
-----
Basic usage:
```go
// Modbus RTU
// Default configuration is 19200, 8, 1, even
client = modbus.RTUClient("/dev/ttyS0")
results, err = client.ReadCoils(2, 1)
```

References
----------
-   [Modbus Specifications and Implementation Guides](http://www.modbus.org/specs.php)
