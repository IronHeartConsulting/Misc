package main

import (
    "fmt"
    "net"
    "os"
	"encoding/hex"
	"encoding/binary"
	"time"
	"flag"
	"strconv"

	"github.com/spf13/viper"
	"github.com/influxdata/influxdb1-client/v2"
)


var ourVersion string = "V0.3"
var dbgLvl uint

var flagDbgLvl = flag.Uint("debug",0,"debug level (0-3)")
var cfgPath = flag.String("config",".","path to config file")


var inflClient  client.Client
var configTitle string
var configDbgLvl uint
var configFile = "ginlong"   // default.  cmdline --config overrides
var	la net.TCPAddr


// config structures
type DatabaseConfig struct {
    Host string `mapstructure:"server"`
    User string `mapstructure:"username"`
    Pass string `mapstructure:"password"`
    DBName string `mapstructure:"DBName"`
}

type IPConfig struct {
	Tcpport string `mapstructure:"tcpport"`
}

type Config struct {
    Db DatabaseConfig `mapstructure:"database"`
	IPinfo IPConfig `mapstructure:"IPinfo"`
}

var c Config

const (
//**    CONN_HOST = "localhost"
    CONN_HOST = "0.0.0.0"
//    CONN_PORT = 8510
    CONN_TYPE = "tcp"
)

const	timeLayout = time.RFC850

type byteField struct {
	fName	string
	fStart	int				// offset to start of field
	fLen	int				// count of bytes in field
	fType	int				//
	fDiv	int				// scale factor, or divisor
	fID		int				// field ID
}

const (
	tInteger = iota
	tFloat
	tString
)

var fPkt []byteField
var	mapBF	map[int]byteField
var recvBuf []byte
var recvLen	int

const (
	header	= "685951b0"
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

	recvBuf = make([]byte, 1024)
	// init the packet field array
	fPkt = []byteField {
		byteField {
		fName:	"header",
		fStart:	0,
		fLen:	4,
		fType:	tInteger,
		fDiv:	1,
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
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_temp,
		},
		byteField {
		fName:	"Voltage DC1",
		fStart:	33,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_vdc1,
		},
		byteField {
		fName:	"Voltage DC2",
		fStart:	35,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_vdc2,
		},
		byteField {
		fName:	"Amps DC1",
		fStart:	39,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_DCamps1,
		},
		byteField {
		fName:	"Amps DC2",
		fStart:	41,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_DCamps2,
		},
		byteField {
		fName:	"AC amps",
		fStart:	45,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_ACamps,
		},
		byteField {
		fName:	"AC voltage",
		fStart:	51,
		fLen:	2,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_ACvolt,
		},
		byteField {
		fName:	"Frequency (Hz)",
		fStart:	57,
		fLen:	2,
		fType:	tInteger,
		fDiv:	100,
		fID:	fID_freq,
		},
		byteField {
		fName:	"Watts",
		fStart:	59,
		fLen:	2,
		fType:	tInteger,
		fDiv:	1,
		fID:	fID_watts,
		},
		byteField {
		fName:	"kWh yesterday",
		fStart:	67,
		fLen:	2,
		fType:	tInteger,
		fDiv:	100,
		fID:	fID_kWhYD,
		},
		byteField {
		fName:	"kWh today",
		fStart:	69,
		fLen:	2,
		fType:	tInteger,
		fDiv:	100,
		fID:	fID_kWhDY,
		},
		byteField {
		fName:	"kWh total",
		fStart:	71,
		fLen:	4,
		fType:	tInteger,
		fDiv:	10,
		fID:	fID_kWhtot,
		},
		byteField {
		fName:	"kWh this month",
		fStart:	87,
		fLen:	2,
		fType:	tInteger,
		fDiv:	1,
		fID:	fID_kWhmth,
		},
		byteField {
		fName:	"kWh last month",
		fStart:	91,
		fLen:	2,
		fType:	tInteger,
		fDiv:	1,
		fID:	fID_kWhlm,
		},
	}
	mapBF = make(map[int]byteField)
	// iterate over the array just built...
	for i, bF := range fPkt {
		mapBF[bF.fID] = bF
		if dbgLvl >= 5 {fmt.Printf("index:%d elem:%v\n",i, bF) }
	}
	// echo the map 
	if dbgLvl >=5 {
		for bfID, bF := range mapBF {
			fmt.Printf("ID:%v BF%v\n",bfID,bF)
		}
	}

}

// use influxDB to store readings
func initDB() error {
    var err error

    // Create a new influx HTTPClient
    inflClient, err = client.NewHTTPClient(client.HTTPConfig{
        Addr:     "http://"+c.Db.Host+":8086",
        Username: c.Db.User,
        Password: c.Db.Pass,
    })
    if err != nil {
        fmt.Printf("ginlong: init DB failed%s\n",err)
		return err
    }
	return nil

}


// prep and init Server - incoming TCP connections from ginlong inverter
func initServer() {
	la.IP = net.IPv4(0, 0, 0, 0)

	la.Port, _ = strconv.Atoi(c.IPinfo.Tcpport)
	// la.Port = CONN_PORT
}

// retrieve general config informoation
func loadConfig() (err error){

    viper.SetConfigName(configFile)
    viper.AddConfigPath(".")
    viper.AddConfigPath(*cfgPath)
	viper.SetDefault("debugLevel", 0)
	viper.WatchConfig()
	// viper.OnConfigChange(func(e fsnotify.Event) {
	// 		fmt.Println("Config file changed:", e.Name)
	// 	})
    err = viper.ReadInConfig()
    if err != nil {
        fmt.Printf("Config file error:%s\n",err)
        return(err)
    }
    configTitle = viper.GetString("title")
	configDbgLvl = viper.GetUint("debugLevel")
    fmt.Printf("%s\n",configTitle)
    if err := viper.Unmarshal(&c); err != nil {
        fmt.Printf("couldn't read config:%s\n", err)
    }
	// fmt.Printf("%v\n",c)

	dbgLvl = viper.GetUint("debugLevel")
	// select debug level.  Use cmd line, if it's present, else use value from config file
	flag.Visit( func (isPresentFlag *flag.Flag) {
		fmt.Printf("flags present name:%s\n",isPresentFlag.Name)
		if isPresentFlag.Name == "debug"  {
			dbgLvl = *flagDbgLvl
		}
	})
    return (nil)

}


//  get a value field from the buffer
//  Input: byteField Output:flaot64 value
//	
func  gfFloat(BF byteField) (float64) {
	rtnVal := -1.99
	// BF := mapBF[fID]
	// fmt.Printf("getField:%v\n",BF)
	switch BF.fLen {
	case 2:
		intVal := binary.BigEndian.Uint16(recvBuf[BF.fStart:])
		rtnVal = (float64(intVal))/float64(BF.fDiv)
	case 4:
		intVal := binary.BigEndian.Uint32(recvBuf[BF.fStart:])
		rtnVal = (float64(intVal))/float64(BF.fDiv)
	default:
		if dbgLvl >=1 {fmt.Printf("getField: err - unknown field len:%d\n",BF.fLen) }
	}
	return rtnVal
}

//  get a string field from the buffer
//  Input: byteField Output:
//	
func  gfString(BF byteField) ([]byte) {
	fEnd := BF.fLen + BF.fStart
	rtnval := make([]byte,BF.fLen)
	copy(rtnval, recvBuf[BF.fStart:fEnd])
	return rtnval
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	var value float64
	// var BF byteField
	// printList := [...]int{fID_vdc1, fID_DCamps1, fID_vdc2, fID_DCamps2, fID_ACvolt, fID_ACamps,
	//			fID_freq, fID_watts, fID_kWhDY, fID_temp }
	printList := [...]int{fID_watts, fID_kWhDY, fID_kWhtot, fID_vdc1, fID_DCamps1, fID_vdc2,
			fID_DCamps2, fID_ACvolt, fID_ACamps, fID_freq,fID_kWhYD, fID_kWhmth, fID_kWhlm, fID_temp }

	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(recvBuf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	if dbgLvl >= 1 {fmt.Println("msg from inverter, len:",reqLen) }
	if dbgLvl >= 2 { fmt.Println(hex.Dump(recvBuf[0:reqLen])) }
	_, err = fmt.Fprintf(dumpFile, string(recvBuf[0:reqLen]))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}

	if dbgLvl >= 1 { fmt.Printf("S/N:%s\n",string(gfString(mapBF[fID_sn]))) }
	t:= time.Now()
	fmt.Printf("%s ", t.Format(timeLayout))
	for _, fieldID := range printList {
		BF := mapBF[fieldID]
		value = gfFloat(BF)
		if dbgLvl >= 2 {fmt.Printf("%s:%.2f\n",BF.fName,value) }
		fmt.Printf("%.2f ",value)
	}
	fmt.Printf("\n")
	conn.Close()
}


func main() {
	var err error
	flag.Parse()

	fmt.Printf("*** ginlong: server for ginlong data write. Version:%s ***\n",ourVersion)
	err = loadConfig()
	fmt.Printf("debug level:%d\n",dbgLvl)
	if  err != nil {
		os.Exit(1)
	}

	initDB()
	initServer()

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
    fmt.Println("Listening on " + CONN_HOST + ":" + c.IPinfo.Tcpport)
	// fmt.Printf("interface:%v\n",l.Addr)
    for {
        // Listen for an incoming connection.
        conn, err := l.AcceptTCP()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
		ra := conn.RemoteAddr().(*net.TCPAddr)
		fmt.Printf("client addr:%v  ",ra)
        // Handle connections in a new goroutine.
        go handleRequest(conn)
    }
}
