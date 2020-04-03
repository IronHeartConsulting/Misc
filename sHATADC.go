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

const i2cAddr = 0x41

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {


//	cmd line flags
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

}
