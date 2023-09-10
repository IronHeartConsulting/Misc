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
const bootStatus byte = 0x2D
const devInfo byte = 0x2f

var i2c *i2c_mod.I2C

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

// decodeBS - decode boot status register - 5 bytes LE
func decodeBS(buf []byte) {

	var srcName string

	fmt.Printf("           HW version:%x\n", buf[5])
	ptchCfgSrc := (buf[4] & 0xe0) >> 5
	switch ptchCfgSrc {
	case 0:
		srcName = "no config loaded"
	case 1, 2, 3, 4, 7:
		srcName = "Reserved"
	case 5:
		srcName = "EEPROM"
	case 6:
		srcName = "I2C"
	}

	fmt.Printf("  patch config source:%d\n", (buf[4]&0xe0)>>5) // bits 31:29
	fmt.Printf("         patch config:%s\n", srcName)
	fmt.Printf("thermal caused reboot:%d\n", (buf[3]&0x08)>>3) // bit 19
	fmt.Printf("   patch download err:%d\n", (buf[2]&0x04)>>2) // bit 10
	fmt.Printf("       EEPROM Present:%d\n", (buf[1]&0x08)>>3) // bit 3
	fmt.Printf("            dead batt:%d\n", (buf[1]&0x04)>>2) // bit 2
	fmt.Printf(" patch bundle hdr err:%d\n", buf[1]&0x01)      // bit 0

}

func fetchReg(addr byte, count int) []byte {
	var err error
	var buf []byte

	buf, count, err = i2c.ReadRegBytes(addr, count+1)
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}
	// fmt.Printf("reg:%x readcount:%d buf:%s\n", addr, count, buf)
	return buf

}

func fetchRegv(addr byte, count int) {
	var err error
	var buf []byte

	buf, count, err = i2c.ReadRegBytes(addr, count+1)
	if err != nil {
		fmt.Printf("error:%s\n", err)
	}
	fmt.Printf("reg:%x readcount:%d buf:%x\n", addr, count, buf)

}

func main() {

	var err error
	var buf []byte

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

	buf = fetchReg(modeReg, 4)
	fmt.Printf("mode:%s\n", buf)
	buf = fetchReg(typeReg, 4)
	fmt.Printf("type:%s\n", buf)
	buf = fetchReg(verReg, 4)
	fmt.Printf("ver string:%x\n", buf)
	fmt.Printf("ver: %02X%02X.%02X.%02X\n", buf[4], buf[3], buf[2], buf[1])
	buf = fetchReg(devCap, 4)
	fmt.Printf("device capabilities:%x\n", buf)
	buf = fetchReg(bootStatus, 5)
	fmt.Printf("boot status:%x\n", buf[1:6])
	decodeBS(buf)
	buf = fetchReg(devInfo, 40)
	fmt.Printf("device info:%s\n", buf)

}
