//  serialMdmCtl - read and control the DTR and RT lines of a serial port
package main

import (
	"fmt"
	"flag"
	"os"

	// "github.com/pkg/term"
	// ser "github.com/chimera/go-inside/rs232"
	unix "golang.org/x/sys/unix"
)

var devName = flag.String("dev", "???", "device name/path to control")
var verbose = flag.Bool("v", false,  "if set, put out more info")
var pause = flag.Bool("p", false,  "if set, pause before exiting")
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


func main()  {
	var err error
	var cnt	int
	var mdmSig string
	var mdmSet string
	var quitQ	string

	flag.Parse()

    fd, err = unix.Open(*devName,unix.O_RDWR | unix.O_NOCTTY, unix.S_IRUSR)
    if err != nil {
        fmt.Printf("%s\n",err)
        return
    }
	defer unix.Close(fd)

	if *stateOnly {
		mdmCtls, err := unix.IoctlGetInt(fd, unix.TIOCMGET)
		if err != nil {
			fmt.Printf("ioctl get failed:%s fd:%d\n",err,fd)
		} else {
			printInfo(mdmCtls)
		}
		return
	}

	for {
		fmt.Printf("enter:signal  state>")
		cnt, err = fmt.Scanln(&mdmSig, &mdmSet)
		if err != nil {
			fmt.Printf("input scanln failed:%s\n",err)
			os.Exit(1)
		}
		if cnt != 2 {
			fmt.Printf("input arg count wrong:%d\n",cnt)
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("open failed:%s\n",err)
			return
		}
		mdmCtls, err := unix.IoctlGetInt(fd, unix.TIOCMGET)
		if err != nil {
			fmt.Printf("ioctl get failed:%s fd:%d\n",err,fd)
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
		default:
			fmt.Printf("unknwon modem signal:%s\n",mdmSig)
			continue
		}
		err = unix.IoctlSetPointerInt(fd, unix.TIOCMSET, mdmCtls)
		if err != nil {
			fmt.Printf("ioctl set failed:%s\n",err)
		}
		if *verbose {
			printInfo(mdmCtls)
		}
		if *pause {
			fmt.Printf("Paused - Q to quit, C to continue:")
			cnt, err = fmt.Scanln(&quitQ)
			if err != nil {
				fmt.Printf("input scanln failed:%s\n",err)
				os.Exit(1)
			}
			if quitQ == "Q" {
				break;
			} else {
				continue
			}
		}
	}
	fmt.Println("bye")
}
