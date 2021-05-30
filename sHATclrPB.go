//
//	  I2C test for secureHAT  V2
//		loop on clear PB latch
//
//

package main

import (
	"flag"
	"time"
	"os"
	"bufio"
	"fmt"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
)

const i2cAddr = 0x20
// 0 = on ; 1 = off
const LED_RED			= 0x80
const LED_GRN			= 0x40
const LED_BLU			= 0x20
// 0 = off ; 1 = on
const VIBR				= 0x10
// pulse 1 to reset PB pressed SR latch
const PBReset			= 0x02
const PB				= 0x04

const peIODIR			= 0x00
const peIPOL			= 0x01
const peGPINTEN			= 0x02
const peDEFVAL			= 0x03
const peINTCON			= 0x04
const peIOCON			= 0x05
const peGPPU			= 0x06
const peINTF			= 0x07
const peINTCAP			= 0x08
const peGPIO			= 0x09
const peOLAT			= 0x0A

var shadowPEGPIO		byte
var GPIOButton			byte

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

var i2cHandle	*i2c.I2C

// init PE
func initPE() {

	var i2cBuf []byte
	var ioCount int
	var err error

	i2cBuf = make ([]byte, 2)
	lg.Info("---init PE---")

	i2cBuf = i2cBuf[:2]
	i2cBuf[0] = peIODIR
	i2cBuf[1] = 0x04
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)

	i2cBuf[0] = peIPOL
	i2cBuf[1] = 0x00
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)

	i2cBuf[0] = peGPINTEN
	i2cBuf[1] = 0x04
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)

	i2cBuf[0] = peINTCON
	i2cBuf[1] = 0x04
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)

	i2cBuf[0] = peIOCON
	i2cBuf[1] = 0x22
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)
	shadowPEGPIO = 0xe0  // all three LEDS off, vibrator off, SR latch not pulsed

}


// read GPIO into shadow register
//   only one bit we care about  - GP2 - pushbutton
func readGPIO() {

	var i2cBuf []byte
	var ioCount int
	var err error

	lg.Info("---read GPIO reg---")
	i2cBuf = make ([]byte, 2)
	i2cBuf = i2cBuf[:1]
	i2cBuf[0] = peGPIO
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)

	// read GPIO reg
	ioCount, err = i2cHandle.ReadBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("read count:%d bytes",ioCount)
	// ignore all the output bits
	GPIOButton = PB & i2cBuf[0]
}

// write back shadow GPIO to PE
func writeGPIO() {

	var i2cBuf []byte
	var ioCount int
	var err error

	i2cBuf = make ([]byte, 2)
	lg.Info("---write back GPIO---")

	i2cBuf = i2cBuf[:2]
	i2cBuf[0] = peGPIO
	i2cBuf[1] = shadowPEGPIO
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)
}


// turn off all LEDs
func LEDaOFF()  {
	shadowPEGPIO = shadowPEGPIO | (LED_GRN | LED_RED | LED_BLU)
	writeGPIO()
}

// turn off LED
func LEDoff(LED byte) {
	// set bit in GPIO to 1
	shadowPEGPIO = shadowPEGPIO | LED
	writeGPIO()
}


// turn on LED
func LEDon(LED byte) {
	// set bit in GPIO to 0
	shadowPEGPIO = shadowPEGPIO & ^LED
	writeGPIO()
}

// turn on vibrator
func vibrON(vibr byte)  {
	shadowPEGPIO = shadowPEGPIO | vibr
	writeGPIO()
}


// turn off vibrator
func vibrOFF(vibr byte)  {
	shadowPEGPIO = shadowPEGPIO & ^vibr
	writeGPIO()
}


//  clear PB latch - pulse briefly
func clearPBLatch(pb byte)  {
	shadowPEGPIO = shadowPEGPIO | pb
	writeGPIO()
	time.Sleep(5* time.Millisecond)
	shadowPEGPIO = shadowPEGPIO & ^pb
	writeGPIO()
}


//    ***  main  ***
//    
func main() {


//	cmd line flags

	debugFlag := flag.Bool("debug", false, "control debugging output")
	flag.Parse()

	lg.Info("secureHAT V2 I2C test V0.2")

	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(i2cAddr, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	i2cHandle = i2c

	if (*debugFlag) {
		logger.ChangePackageLogLevel("i2c", logger.DebugLevel)
	} else {
		logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	}

	initPE()
	readGPIO()
	writeGPIO()
	LEDaOFF()
	vibrOFF(VIBR)
	// need to clear latch after boot..make sure hte SR latch is in the set state
	clearPBLatch(PBReset)
	// loop on input, and clear PB latch
	lg.Info("starting PB latch clear loop")
	rdrsin := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("clear PB latch? ")
		rdrsin.ReadString('\n')
		readGPIO()
		if GPIOButton > 0 {
			lg.Info("PB active")
		}
		clearPBLatch(PBReset)
	}
}
