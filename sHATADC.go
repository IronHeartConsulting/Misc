//
//    read ADC values from secureHAT
//
//

package main

import (
	"flag"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
)

const i2cAddr = 0x32

var i2cCmdReadADC []byte

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {


//	cmd line flags
	var i2cBuf []byte
	var ioCount int

	i2cBuf = make ([]byte, 8)
	debugFlag := flag.Bool("debug", false, "control debugging output")
	flag.Parse()

	lg.Info("secureHAT Read ADC value  V0.1")

	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(i2cAddr, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	if (*debugFlag) {
		logger.ChangePackageLogLevel("i2c", logger.DebugLevel)
	} else {
		logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	}

	i2cCmdReadADC = make([]byte,1)
	i2cCmdReadADC[0] = 4
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2c.WriteBytes(i2cCmdReadADC)
	// read eight bytes -
	ioCount, err = i2c.ReadBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("read count:%d bytes read:%v",ioCount, i2cBuf)
	vin_v := uint16(i2cBuf[0]) << 8 | uint16(i2cBuf[1])
	lg.Infof("Vin:%d",vin_v)
}
