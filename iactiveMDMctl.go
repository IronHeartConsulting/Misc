//  interactive model control
//  serialMdmCtl - read and control the DTR and RT lines of a serial port
package main

import (
	"fmt"
	"flag"

	unix "golang.org/x/sys/unix"
	gp "github.com/c-bata/go-prompt"
)

var devName = flag.String("dev", "???", "device name/path to control")
var verbose = flag.Bool("v", false,  "if set, put out more info")
var pause = flag.Bool("p", false,  "if set, pause beofre exiting")
var stateOnly = flag.Bool("s", false,  "if set just print the state, and exit")
var fd	int

func printInfo( mdmInfo int) {

	fmt.Printf("modem info:%02x DTR:",mdmInfo)
	if (mdmInfo & unix.TIOCM_DTR)  > 0 {
		fmt.Printf("on   RTS:")
	} else {
		fmt.Printf("Off   RTS:")
	}
	if (mdmInfo & unix.TIOCM_RTS)  > 0 {
		fmt.Printf("on\n")
	} else {
		fmt.Printf("Off\n")
	}
	// fmt.Printf("device:%s signal:%s state:%s\n",*devName, mdmSig, mdmSet)
	// fmt.Printf("DTR:%t RTS:%t\n",stateDTR, stateRTS)
}

func completer(d gp.Document) []gp.Suggest {
	s := []gp.Suggest {
		{Text: "DTR". Description: "DTR mdoem control signal"},
		{Text: "RTS". Description: "RTS mdoem control signal"},
	}
	return gp.FilterHasPrefix(s, d.GetWordBefoeCursor(), true)
}

func main()  {
	var err error

	flag.Parse()
	// if flag.NArg() != 2 {
// 		flag.PrintDefaults()
// 		fmt.Println("signal = DTR|RTS state = on|off")
// 		return
// 	}
	// mdmSig := flag.Arg(0)
	// mdmSet := flag.Arg(1)

	fmt.Println("Select signal")
	t := gp.Input("> ", completer)
	fmt.Println("signal selected:" + t)
	os.Exit(0)

	fd, err = unix.Open(*devName,unix.O_RDWR | unix.O_NOCTTY, unix.S_IRUSR)
	if err != nil {
		fmt.Printf("%s\n",err)
		return
	}
	defer unix.Close(fd)
	mdmCtls, err := unix.IoctlGetInt(fd, unix.TIOCMGET)
	if err != nil {
		fmt.Printf("ioctl get failed:%s\n",err)
		return
	}
	if (*verbose  || *stateOnly ) {
		printInfo(mdmCtls)
	}
	if *stateOnly {
		return
	}
	switch mdmSig {
	case "DTR":
		switch mdmSet {
		case "on":
			mdmCtls |= unix.TIOCM_DTR
			// p.SetDTR(true)
		case "off":
			mdmCtls &= ^(unix.TIOCM_DTR)
			// p.SetDTR(false)
		}
	case "RTS":
		switch mdmSet {
		case "on":
			mdmCtls |= unix.TIOCM_RTS
			// p.SetRTS(true)
		case "off":
			mdmCtls &= ^(unix.TIOCM_RTS)
			// p.SetRTS(false)
		}
	}
	err = unix.IoctlSetPointerInt(fd, unix.TIOCMSET, mdmCtls)
	if err != nil {
		fmt.Printf("ioctl set failed:%s\n",err)
	}
	if *verbose {
		printInfo(mdmCtls)
	}
	if *pause {
		fmt.Println("Paused - Press Enter to resume")
		fmt.Scanln()
	}
}
