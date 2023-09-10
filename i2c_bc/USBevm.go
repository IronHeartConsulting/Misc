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
const CUSTUSE byte = 0x06
const verReg byte = 0x0f
const devCap byte = 0x0d
const statusReg byte = 0x1a
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

// decodeDC - deocde device capabiliites
func decodeDC(buf []byte) {

	I2Cmlvl := (buf[1] & 0x80) >> 7
	BC1p2Spted := (buf[1] & 0x60) >> 5
	USBPDCap := (buf[1] & 0x04) >> 2
	powerRole := buf[1] & 0x03

	fmt.Println("Device Capabiliites")
	if I2Cmlvl == 1 { // 3.3V pull up voltage
		fmt.Printf("I2C master pullup voltage:3.3V\n")
	} else {
		fmt.Printf("I2C master pullup voltage:1.8V or 3.3V\n")
	}
	switch BC1p2Spted {
	case 0:
		fmt.Printf(" BC1.2 Not Supported\n")
	case 1:
		fmt.Printf("BC1.2 Only source supported\n")
	case 2, 3:
		fmt.Printf("(BC 1.2) Reserved\n")
	}
	if USBPDCap == 0 {
		fmt.Println("USB PD supported")
	} else {
		fmt.Println("USB PD Not supported")
	}
	switch powerRole {
	case 0:
		fmt.Println("Power role (DRP) sink and src")
	case 1:
		fmt.Println("Power Role: src only")
	case 2:
		fmt.Println("power role: undefined")
	case 3:
		fmt.Println("power role: src only")
	}
	fmt.Println()
}

// decodeST - deocde status register
func decodeST(buf []byte) {
	bist := (buf[4] & 0x04) >> 2
	legacy := buf[4] & 0x03
	USBHost := (buf[3] & 0xc0) >> 6
	VbusStatus := (buf[3] & 0x30) >> 4
	dataRole := (buf[1] & 0x40) >> 6
	portRole := (buf[1] & 0x20) >> 5
	plugOrient := (buf[1] & 0x10) >> 4

	if bist == 1 {
		fmt.Println("BIST in progress")
	}
	switch legacy {
	case 0:
		fmt.Println("not in Legacy (non PD) mode")
	case 1:
		fmt.Println("Legacy sink")
	case 2:
		fmt.Println("legacy source")
	case 3:
		fmt.Println("need dead battery cleared")
	}
	switch USBHost {
	case 0:
		fmt.Println("no host present")
	case 1:
		fmt.Println("Port partner is PD - no USB")
	case 2:
		fmt.Println("port partner - non PD")
	case 3:
		fmt.Printf("Host present and PD")
	}
	fmt.Printf("VBUS Status:")
	switch VbusStatus {
	case 0:
		fmt.Printf("vSafe0V\n")
	case 1:
		fmt.Printf("vSafe5V\n")
	case 2:
		fmt.Printf("within limits\n")
	case 3:
		fmt.Printf("not within known range\n")
	}
	if dataRole == 1 {
		fmt.Println("data role: DFP")
	} else {
		fmt.Println("data role:UFP")
	}
	if portRole == 1 {
		fmt.Println("PD controller is SOURCE")
	} else {
		fmt.Println("PD controller is SINK")
	}
	if plugOrient == 1 {
		fmt.Println("plug - upside down orientation")
	} else {
		fmt.Println("plug - normal orientation")
	}
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
	var limitReg bool = true

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
	fmt.Println()
	if string(buf[1:5]) == "APP " {
		limitReg = false
	}

	buf = fetchReg(typeReg, 4)
	fmt.Printf("type:%s\n", buf)
	buf = fetchReg(verReg, 4)
	// fmt.Printf("ver string:%x\n", buf)
	fmt.Printf("ver: %02X%02X.%02X.%02X\n", buf[4], buf[3], buf[2], buf[1])
	buf = fetchReg(devCap, 4)
	fmt.Printf("device capabilities:%x\n", buf)
	decodeDC(buf)
	buf = fetchReg(bootStatus, 5)
	// fmt.Printf("boot status:%x\n", buf[1:6])
	decodeBS(buf)
	buf = fetchReg(devInfo, 40)
	fmt.Printf("device info:%s\n", buf)
	if limitReg == true { // can't read any more registers, not in APP mode
		fmt.Println("limited register access")
		return
	}
	buf = fetchReg(CUSTUSE, 8)
	fmt.Printf("Customer string:%s\n", buf)
	buf = fetchReg(statusReg, 5)
	decodeST(buf)
}
