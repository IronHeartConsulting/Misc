package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// read eznec far field data file, convert to cvs format

var ourVersion string = "V0.1"
var dbgLvl uint

var flagDbgLvl = flag.Uint("debug", 0, "debug level (0-?)")

const (
	freqIndex      int = 2
	elevAngleIndex int = 5
	AZAngleIndex   int = 0
	VdBIndex       int = 1
	HdBIndex       int = 2
	TotaldBIndex   int = 3
	VPhaseindex    int = 4
	HPhaseIndex    int = 5
)

func main() {
	var err error
	var inFileD, outFileD *os.File
	var inFileS *bufio.Scanner
	var lineCount int = 0
	var splits []string
	var refFreq float64
	var elevAngle float64
	var reSSN *regexp.Regexp
	var AZAngle int
	var VdB, HdB, TotaldB, VPhase, HPhase float64

	reSSN, err = regexp.Compile(`\s*[0-9]*\s*\-?[0-9]+\.[0-9]+`)
	if err != nil {
		fmt.Printf("SSN regexp failed to compile:%s\n", err)
		os.Exit(1)
	}

	flag.Parse()

	fmt.Printf("*** eznec far field file convert to CSV type. Version:%s ***\n", ourVersion)
	fmt.Printf("debug level:%d\n", dbgLvl)

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	inFileName := flag.Arg(0)
	outFileName := flag.Arg(1)
	fmt.Printf(("input file name:%s  Output file name:%s\n"), inFileName, outFileName)
	// read input file, line by line
	inFileD, err = os.Open(inFileName)
	if err != nil {
		fmt.Printf("input file:%s failed open with error:%s\n", inFileName, err)
		os.Exit(1)
	}
	outFileD, err = os.Create(outFileName)
	if err != nil {
		fmt.Printf("output file%s failed to open with error:%s\n", outFileName, err)
		os.Exit(1)
	}
	defer inFileD.Close()
	defer outFileD.Close()
	fmt.Fprintf(outFileD, "Freq (MHz), EL, AZ, V dB, H dB, Total dB, V Phase, H Phase\n")
	inFileS = bufio.NewScanner(inFileD)
	for inFileS.Scan() {
		lineCount++
		splits = strings.Fields(inFileS.Text())
		if len(splits) == 0 {
			continue
		}
		// fmt.Printf("line:%d - %s\n", lineCount, inFileS.Text())
		if splits[0] == "Frequency" {
			refFreq, _ = strconv.ParseFloat(splits[freqIndex], 32)
			fmt.Printf(" (New) reference freq:%f\n", refFreq)
			continue
		}
		if splits[0] == "Azimuth" {
			elevAngle, _ = strconv.ParseFloat(splits[elevAngleIndex], 32)
			fmt.Printf("(New) elevation Angle:%f\n", elevAngle)
			// fmt.Printf("Freq (MHz), EL, AZ, V dB, H dB, Total dB, V Phase, H Phase\n")
			continue
		}
		if splits[0] == "Deg" {
			continue
		}
		listSSN := reSSN.FindAll([]byte(inFileS.Text()), -1)
		if (listSSN != nil) && (len(listSSN) == 5) {
			// fmt.Printf("line:%d - %s\n", lineCount, inFileS.Text())
			// fmt.Printf("splits:%v  ", splits)
			// fmt.Printf("SSN list:%q len:%d  ", listSSN, len(listSSN))
			AZAngle, _ = strconv.Atoi(splits[0])
			VdB, _ = strconv.ParseFloat(splits[VdBIndex], 32)
			HdB, _ = strconv.ParseFloat(splits[HdBIndex], 32)
			TotaldB, _ = strconv.ParseFloat(splits[TotaldBIndex], 32)
			VPhase, _ = strconv.ParseFloat(splits[VPhaseindex], 32)
			HPhase, _ = strconv.ParseFloat(splits[HPhaseIndex], 32)
			// fmt.Printf("%f, %0.2f, %d, %0.2f, %0.2f, %0.2f, %0.2f, %0.2f\n", refFreq, elevAngle, AZAngle, VdB, HdB, TotaldB, VPhase, HPhase)
			fmt.Fprintf(outFileD, "%f, %0.2f, %d, %0.2f, %0.2f, %0.2f, %0.2f, %0.2f\n", refFreq, elevAngle, AZAngle, VdB, HdB, TotaldB, VPhase, HPhase)
		} else {
			fmt.Println(" ")
		}
	}
	fmt.Printf("total lines:%d\n", lineCount)
}
