package main

import (
    "fmt"
    "net"
    "os"
	"encoding/hex"
	"time"
)

const (
//**    CONN_HOST = "localhost"
    CONN_HOST = "0.0.0.0"
    CONN_PORT = 8510
    CONN_TYPE = "tcp"
)

type byteField struct {
	fName	string
	fStart	int				// offset to start of field
	fLen	int				// count of bytes in field
	fType	int				//
	fID		int				// field ID
}

const (
	tInteger = iota
	tInteger10				// in tenths
	tInteger100				// in hundereths
	tFloat
	tString
)

var fPkt []byteField

const (
	header	= "685951b0"
	sn		= 15
	temp	= 31
	vdc1	= 33
	vdc2	= 35
	DCamps1	= 39
	DCamps2	= 41
	ACamps	= 45
	ACVolt	= 51
	freq	= 57
	watts	= 59
	kwh_yd	= 67
	kwhDY	= 69
	kWhtot	= 71
	kWhmth	= 87
	kWhlm	= 91
)

const (
	fID_header	= iota
	fID_sn
	fID_temp
	fID_vdc1
	fID_vdc2
	fID_DCamps1
	fID_DCamps2
	fID_ACamps
	fID_ACvolt
	fID_freq
	fID_watts
	fID_kWhYD
	fID_kWhDY
	fID_kWhtot
	fID_kWhmth
	fID_kWhlm
)
var dumpFile *os.File

func init() {

	// init the packet field array
	fPkt = []byteField {
		byteField {
		fName:	"header",
		fStart:	0,
		fLen:	4,
		fType:	tInteger,
		fID:	fID_header,
		},
		byteField {
		fName:	"S/N",
		fStart:	15,
		fLen:	17,
		fType:	tString,
		fID:	fID_sn,
		},
		byteField {
		fName:	"Temperature",
		fStart:	31,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_temp,
		},
		byteField {
		fName:	"Voltage DC1",
		fStart:	33,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_vdc1,
		},
		byteField {
		fName:	"Voltage DC2",
		fStart:	35,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_vdc2,
		},
		byteField {
		fName:	"Amps DC1",
		fStart:	39,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_DCamps1,
		},
		byteField {
		fName:	"Amps DC2",
		fStart:	41,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_DCamps2,
		},
		byteField {
		fName:	"AC amps",
		fStart:	45,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_ACamps,
		},
		byteField {
		fName:	"AC voltage",
		fStart:	51,
		fLen:	2,
		fType:	tInteger10,
		fID:	fID_ACvolt,
		},
		byteField {
		fName:	"Frequency (Hz)",
		fStart:	57,
		fLen:	2,
		fType:	tInteger100,
		fID:	fID_freq,
		},
		byteField {
		fName:	"Watts so far today",
		fStart:	59,
		fLen:	2,
		fType:	tInteger100,
		fID:	fID_watts,
		},
		byteField {
		fName:	"kWh yesterday",
		fStart:	67,
		fLen:	2,
		fType:	tInteger100,
		fID:	fID_kWhYD,
		},
		byteField {
		fName:	"kWh today",
		fStart:	69,
		fLen:	2,
		fType:	tInteger100,
		fID:	fID_kWhDY,
		},
		byteField {
		fName:	"kWh total",
		fStart:	71,
		fLen:	4,
		fType:	tInteger10,
		fID:	fID_kWhtot,
		},
		byteField {
		fName:	"kWh this month",
		fStart:	87,
		fLen:	2,
		fType:	tInteger,
		fID:	fID_kWhmth,
		},
		byteField {
		fName:	"kWh last month",
		fStart:	91,
		fLen:	2,
		fType:	tInteger,
		fID:	fID_kWhlm,
		},
	}
	// iterate over the arary just built...
	for i, bF := range fPkt {
		fmt.Printf("index:%d elem:%v\n",i, bF)
	}

}

func main() {
	var err error
	var	la net.TCPAddr

	la.IP = net.IPv4(0, 0, 0, 0)
	la.Port = CONN_PORT

	dumpFile, err = os.OpenFile("inverter_dump", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("File does not exist or cannto be created")
		os.Exit(1)
	}
	defer dumpFile.Close()

    // Listen for incoming connections.
    l, err := net.ListenTCP(CONN_TYPE, &la)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes.
    defer l.Close()
    // **** fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	// fmt.Printf("interface:%v\n",l.Addr)
    for {
        // Listen for an incoming connection.
        conn, err := l.AcceptTCP()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
		ra := conn.RemoteAddr().(*net.TCPAddr)
		fmt.Printf("client addr:%v\n",ra)
        // Handle connections in a new goroutine.
        go handleRequest(conn)
    }
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
	fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("msg from inverter, len:",reqLen)
	fmt.Printf("%s\n",time.Now())
	fmt.Println(hex.Dump(buf[0:reqLen]))
	_, err = fmt.Fprintf(dumpFile, string(buf[0:reqLen]))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
	conn.Close()
}
