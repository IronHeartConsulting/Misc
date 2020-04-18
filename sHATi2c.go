//
//	  I2C test for secureHAT
//    read ADC values from secureHAT
//    read power control variables
//    write power control variables
//
//

package main

import (
	"flag"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
)

const i2cAddr = 0x32
const i2cCmdReadADC		= 4
const i2cCmdReadPwr		= 10
const i2cCmdWritePwr	= 11
const i2cCmdSetLEDs		= 7
const i2cCmdSetVib		= 8
const i2cCmdReadEEPROM	= 5
const i2cCmdWriteEEPROM = 6

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

//  read the ADC values
//
func readADC(i2cHandle *i2c.I2C ) {
	var i2cBuf []byte
	var ioCount int
	var err error

	i2cBuf = make ([]byte, 8)

	lg.Info("---reading ADC values---")

	i2cBuf = i2cBuf[:1]
	i2cBuf[0] = i2cCmdReadADC
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	// read ADC values
	i2cBuf = i2cBuf[:8]
	ioCount, err = i2cHandle.ReadBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("read count:%d bytes read:%v",ioCount, i2cBuf)
	vin_v := uint16(i2cBuf[0]) << 8 | uint16(i2cBuf[1])
	lg.Infof("Vin:%d",vin_v)
}


//  read the power variables
//
func readPwr(i2cHandle *i2c.I2C ) {
	var i2cBuf []byte
	var ioCount int
	var err error

	i2cBuf = make ([]byte, 8)
	lg.Info("---reading Power control variables---")

	i2cBuf = i2cBuf[:1]
	i2cBuf[0] = i2cCmdReadPwr
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	// read power variables
	i2cBuf = i2cBuf[:7]
	ioCount, err = i2cHandle.ReadBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("read count:%d bytes read:%v",ioCount, i2cBuf)
	Vcutoff_v := uint16(i2cBuf[0]) << 8 | uint16(i2cBuf[1])
	delayRpi	:= uint8(i2cBuf[2])
	hdUSBOff	:= uint8(i2cBuf[3])
	hdUSBOn	:= uint8(i2cBuf[4])
	PwrOffMode	:= uint8(i2cBuf[5])
	PwrState	:= uint8(i2cBuf[6])
	lg.Infof("Batt cutoff                           :%d",Vcutoff_v)
	lg.Infof("delay on battery start - RPi turn off :%d",delayRpi)
	lg.Infof("USB on->off hold down delay           :%d",hdUSBOff)
	lg.Infof("USB off->on hold down delay           :%d",hdUSBOn)
	switch PwrOffMode {
		case 1:
			lg.Info("power off mode - timer")
		case 2:
			lg.Info("power off mode - low battery")
		default:
			lg.Infof("power off mode - unknown :%d",PwrOffMode)
	}
	switch PwrState {
		case 1:
			lg.Info("power state - Normal")
		case 2:
			lg.Info("power state - on Battery")
		case 3:
			lg.Info("power state - Battery below cutoff")
		default:
			lg.Infof("power control state - unknown %d",PwrState)
	}
}

//  write the power variables
//
func writePwr(i2cHandle *i2c.I2C ) {
	var i2cBuf []byte
	var ioCount int
	var err error

	i2cBuf = make ([]byte, 8)
	lg.Info("---writing Power control variables---")

	i2cBuf = i2cBuf[:7]
	i2cBuf[0] = i2cCmdWritePwr
	i2cBuf[1] = 0
	i2cBuf[2] = 25
	i2cBuf[3] = 200;  // delay power of when going to battery
	i2cBuf[4] = 8;  // USB on to off hold down timer
	i2cBuf[5] = 3;  // USB off to on hold down timer
	i2cBuf[6] = 2;  // power off - low battery
	// send command to secureHAT to read back the ADC value
	ioCount, err = i2cHandle.WriteBytes(i2cBuf)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("wrote count:%d bytes",ioCount)
}

func main() {


//	cmd line flags

	debugFlag := flag.Bool("debug", false, "control debugging output")
	flag.Parse()

	lg.Info("secureHAT Read ADC value  V0.2")

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

	readADC(i2c)
	readPwr(i2c)
	writePwr(i2c)
	readPwr(i2c)
}
