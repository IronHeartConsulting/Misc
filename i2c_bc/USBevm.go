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
const CMD1 byte = 0x08
const DATA1 byte = 0x09
const verReg byte = 0x0f
const devCap byte = 0x0d
const intEvent byte = 0x14
const INT_CLEAR byte = 0x18
const statusReg byte = 0x1a
const powerPathStatus byte = 0x26
const portCtl byte = 0x29
const bootStatus byte = 0x2D
const buildDesc byte = 0x2e
const devInfo byte = 0x2f

const powerConnStatus byte = 0x3f
const pdStatus byte = 0x40
const typeCState byte = 0x69
const VMINSYSreg = 0x6b0001

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

	fmt.Println("----boot status-----0x2d")
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

	fmt.Println("----Device Capabiliites----0x0d")
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

	fmt.Println("----status----0x1a")
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

// decode power path status
func decodePPS(buf []byte) {
	fmt.Printf("-----power path status-----0x26:%x\n", buf)
}

// decode port control - mostly enables to the PD controller
func decodePC(buf []byte) {

	chargerDetectEnable := (buf[4] & 0xc0) >> 6
	chargerAdEnable := (buf[4] & 0x1c) >> 2
	res15kPres := buf[4] & 0x01
	typeCI := buf[1] & 0x03

	fmt.Println("----Port Control----0x29")
	switch chargerDetectEnable {
	case 0:
		fmt.Println("don't detect Legacy chargers")
	case 1:
		fmt.Println("detect BC 1.2 chargers")
	case 2:
		fmt.Println("reserved")
	case 3:
		fmt.Println("Detect BC 1.2 and Legacy")
	}
	fmt.Printf("chargeAd:%d\n", chargerAdEnable)
	fmt.Printf("ressitor 15K presence detect:%d\n", res15kPres)
	fmt.Printf("Type C current:%d\n", typeCI)
}

// decode power connection status (0x3f)
func decodePCS(buf []byte) {

	fmt.Printf("----power connection status------0x3f:%x\n", buf)

	chargerAd := buf[2] & 0x03
	chargerDetect := (buf[1] & 0xf0) >> 4
	typeCI := (buf[1] & 0x0c) >> 2
	sourceSink := (buf[1] & 0x02) >> 1
	poowerConn := buf[1] & 0x01

	if poowerConn == 1 {
		fmt.Println("power connection present")
	} else {
		fmt.Println("no power connection")
		// exit - rest of bits are meaninigless
		return
	}

	switch chargerAd {
	case 0:
		fmt.Println("charger advertises disabled or not run")
	case 1:
		fmt.Println("charger ad in process")
	case 2:
		fmt.Println("charger ad complete")
	case 3:
		fmt.Println("Reserved")
	}
	fmt.Printf("charger detection:")
	switch chargerDetect {
	case 0:
		fmt.Println("diabled or not run")
	case 1:
		fmt.Println("in progress")
	case 2:
		fmt.Println("cmpl, non detected")
	case 3:
		fmt.Println("SDP")
	case 4:
		fmt.Println("BC 1.2 CDP")
	case 5:
		fmt.Println("BC 1.2 DCP")
	case 6:
		fmt.Println("Divider1 DCP")
	case 7:
		fmt.Println("Divider2 DCP")
	case 8:
		fmt.Println("Divider3 DCP")
	case 9:
		fmt.Println("1.2-V DCP")
	default:
		fmt.Println("Reserved")
	}
	fmt.Printf("Type C current:")
	switch typeCI {
	case 0:
		fmt.Println("USB default")
	case 1:
		fmt.Println("1.5 A")
	case 2:
		fmt.Println("3.0 A")
	case 3:
		fmt.Println("PD contracted current")
	}
	if sourceSink == 1 { // we are sink
		fmt.Println("power sink")
	} else {
		fmt.Println("power source")
	}
}

// decodePDS - PD status 0x40
func decodePDS(buf []byte) {

	fmt.Println("----PD status-----0x40")
	dataResetDetails := (buf[4] & 0x70) >> 4
	errRcvyDetails := (buf[4] & 0x0f) | buf[3]&0xc0>>6
	hardResetDetails := (buf[3] & 0x3f) >> 4
	softResetDetails := (buf[2] & 0x1f) >> 4
	currentPDrole := (buf[1] & 0x40) >> 6
	portType := buf[1] & 0x30 >> 4
	CCPullUp := (buf[1] & 0x0c) >> 2

	fmt.Printf("data reset detail:%d\n", dataResetDetails)
	fmt.Printf("error reecovery details:%d\n", errRcvyDetails)
	fmt.Printf("hard reset details:%d\n", hardResetDetails)
	fmt.Printf("soft reset details:%d\n", softResetDetails)
	if currentPDrole == 0 {
		fmt.Println("present PD role: sink")
	} else {
		fmt.Println("present PD Role: source")
	}
	fmt.Printf("present type-C power role:")
	switch portType {
	case 0:
		fmt.Println("sink/source")
	case 1:
		fmt.Println("sink")
	case 2:
		fmt.Println("source")
	case 3:
		fmt.Println("source/sink")
	}
	fmt.Printf("CC Pull-UP value:")
	switch CCPullUp {
	case 0:
		fmt.Println("no CC Pull up detected")
	case 1:
		fmt.Println("USB default current")
	case 2:
		fmt.Println("1.5 A (sinkTXNG)")
	case 3:
		fmt.Println("3.0A (sinkTXOK)")
	}
}

// decodeIntEvt - 0x14 - interrupt event - we don't docde all of the events.
func decodeIntEvt(buf []byte) {

	fmt.Println("----INT Event list---0x14")
	fmt.Printf("interrupt event bits:%x\n", buf)
	I2CMNACKed := (buf[11] & 0x08) >> 3
	CMDcmpl := (buf[4] & 0x40) >> 6
	if I2CMNACKed == 1 {
		fmt.Println("I2C Master cmd NACKed")
	}
	if CMDcmpl == 1 {
		fmt.Println("4CC completed")
	}
}

// clearInts  -- clear Int bits
// expects a 11 byte array.  set a bit to clear that interrupt
func clearInts(buf []byte) {

	fmt.Println("----clear Interrupts----0x18")
	cmdByte := make([]byte, 1)
	cmdByte[0] = INT_CLEAR
	//  cmdByte[1] = 11
	cmdplusbuf := append(cmdByte, buf...)
	wrCount, err := i2c.WriteBytes(cmdplusbuf)
	if err != nil {
		fmt.Printf("err resetting interrupts:% s\n", err)
	}
	fmt.Printf("wrote:%d bytes - INT_CLEAR\n", wrCount)
}

// write a 4CC command
func write4CC(buf []byte) {

	fmt.Println("---write 4CC----0x08")
	cmdByte := make([]byte, 2)
	cmdByte[0] = CMD1 // register
	cmdByte[1] = 0x04 // number of bytes in the data
	cmdplusbuf := append(cmdByte, buf...)
	wrCount, err := i2c.WriteBytes(cmdplusbuf)
	if err != nil {
		fmt.Printf("I2C write failed:%s\n", err)
	}
	fmt.Printf("wrote:%d bytes %s\n", wrCount, buf)
}

// write a 4CC command with data
func write4CCwData(buf []byte, data []byte) {

	var err error
	var wrCount int

	fmt.Println("---write 4CC w/DATA----0x08")

	dataReg := make([]byte, (1 + len(data)))
	dataReg[0] = DATA1
	dataCmd := append(dataReg, data...)
	_, err = i2c.WriteBytes(dataCmd)
	if err != nil {
		fmt.Printf("data reg I2C write failed:%s\n", err)
	}
	fmt.Printf("data buf:%#v\n", data)

	cmdByte := make([]byte, 2)
	cmdByte[0] = CMD1 // register
	cmdByte[1] = 0x04 // number of bytes in the data
	cmdplusbuf := append(cmdByte, buf...)
	wrCount, err = i2c.WriteBytes(cmdplusbuf)
	if err != nil {
		fmt.Printf("I2C write failed:%s\n", err)
	}
	fmt.Printf("wrote:%d bytes %s\n", wrCount, buf)
}

// fill data buffer for 4CC command
func write4CCDataBuf(data []byte) {
	var err error

	fmt.Println("---write 4CC DATA buf----0x09")

	dataReg := make([]byte, (1 + len(data)))
	dataReg[0] = DATA1
	dataCmd := append(dataReg, data...)
	_, err = i2c.WriteBytes(dataCmd)
	if err != nil {
		fmt.Printf("data reg I2C write failed:%s\n", err)
	}
	fmt.Printf("data buf:%#v\n", data)
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

	intBuf := make([]byte, 12)

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

	buf = fetchReg(buildDesc, 49)
	fmt.Printf("build description:%s\n\n", buf)

	buf = fetchReg(statusReg, 5)
	decodeST(buf)

	buf = fetchReg(powerPathStatus, 5)
	decodePPS(buf)

	buf = fetchReg(portCtl, 4)
	decodePC(buf)

	// 0x3F
	buf = fetchReg(powerConnStatus, 2)
	decodePCS(buf)

	// 0x40
	buf = fetchReg(pdStatus, 4)
	decodePDS(buf)

	// 0x69
	buf = fetchReg(typeCState, 4)
	fmt.Printf("Type C state:%x\n", buf)

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// clear interrupts
	buf[11] &= 0x04 // clear I2CMaserNCKed
	buf[4] &= 0x40  // CMDComplete
	clearInts(buf)

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// send 4CC command
	buf = fetchReg(CMD1, 4)
	fmt.Printf("4CC command reg:%s\n", buf)
	write4CC([]byte("DBfg"))

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// read 4CC command register
	buf = fetchReg(CMD1, 4)
	fmt.Printf("4CC command reg:%s\n", buf)

	//send bad command, jsut to make sure it's working
	fmt.Println("*** send bad 4CC cmd *** ")
	// clear interrupts
	intBuf[11] &= 0x04 // clear I2CMaserNCKed
	intBuf[4] &= 0x40  // CMDComplete
	clearInts(intBuf)

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// send bad 4CC command
	buf = fetchReg(CMD1, 4)
	fmt.Printf("4CC command reg:%s\n", buf)
	write4CC([]byte("DBXX"))

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// read 4CC command register
	buf = fetchReg(CMD1, 4)
	fmt.Printf("4CC command reg:%s\n", buf)

	// read from BC  read the VSYSMIN register (0x00) len = 1, BC I2C addr = 0x6b
	// part info (0x48)

	// clear interrupts
	intBuf[11] &= 0x04 // clear I2CMaserNCKed
	intBuf[4] &= 0x40  // CMDComplete
	clearInts(intBuf)

	vminReg := []byte{0x6b, 0x48, 0x01}
	write4CCDataBuf(vminReg)
	// read back what came back in I2C read
	fetchRegv(DATA1, 4)

	write4CC([]byte("I2Cr"))

	// 0x14 - INT_EVENT
	buf = fetchReg(intEvent, 11)
	decodeIntEvt(buf)

	// read 4CC command register
	buf = fetchReg(CMD1, 4)
	fmt.Printf("4CC command reg:%s\n", buf)

	// read back what came back in I2C read
	fetchRegv(DATA1, 4)
}
