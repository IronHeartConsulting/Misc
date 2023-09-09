//
//   USBevm
//
//		I2C access to the salve port on the TI USB charger EVM
//		commands nad registers are documented in TPS25750 TRM
//

package main

import (
	"flag"
	"fmt"

	i2c_mod "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
)

const modeReg byte = 0x03
const typeReg byte = 0x04
const verReg byte = 0x0f
const devCap byte = 0x0d

var i2c *i2c_mod.I2C

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func fetchReg(addr byte, count int) {
	var err error
	var buf []byte

	buf, count, err = i2c.ReadRegBytes(addr, count+1)
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}
	fmt.Printf("reg:%x readcount:%d buf:%s\n", addr, count, buf)

}

func fetchRegv(addr byte, count int) {
	var err error
	var buf []byte

	buf, count, err = i2c.ReadRegBytes(addr, count+1)
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}
	fmt.Printf("reg:%x readcount:%d buf:%v\n", addr, count, buf)

}

func main() {

	var err error

	//	cmd line flags
	debugFlag := flag.Bool("debug", false, "control debugging output")
	flag.Parse()

	lg.Info("I2C command to charger EVM V0.1")

	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x21.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err = i2c_mod.NewI2C(0x21, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	if *debugFlag {
		logger.ChangePackageLogLevel("i2c", logger.DebugLevel)
		logger.ChangePackageLogLevel("bsbmp", logger.DebugLevel)
	} else {
		logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
		logger.ChangePackageLogLevel("bsbmp", logger.InfoLevel)
	}

	fetchReg(modeReg, 4)
	fetchReg(typeReg, 4)
	fetchRegv(verReg, 4)
	fetchRegv(devCap, 4)

}
