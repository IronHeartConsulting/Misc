//
//   Thrermocoupler amplifier
//

package main

import (
	"flag"
	"log"
	"regexp"
	"strings"
	"strconv"
	"time"
	"os"
	"errors"

	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
	"github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/viper"
)

var ourVersion string = "V0.4"

//  regexp indexes into submatches
const (
		probeNumIndex	= 1
		readingIndex	= 2
)
var TCampReading = regexp.MustCompile(`^#(\d):\s(\d+\.\d+)\sdegC Int:\s\d+\.\d+\sdegC`)
// var TCampReading = regexp.MustCompile(`^#(\d):\s(\d+\.\d+)\sdegC`)

var devName = flag.String("dev","UNKN","device name to connect with")
var dbgLvl = flag.Uint("debug",0,"debug level (0-3)")

type TCamp struct {
	serialPort	serial.Port
	serialNum	string
}

//  TODO  convert to map, indexed by probe number from amp 
type TCProbe struct {
	name	string
	ampPortNum	int
//	lastReading float32
}

var TCbyNum map[int]TCProbe

var TCs [8]TCProbe

var TCamp_1 TCamp
var influxHostName string
var HTTPClient client.Client
var configTitle string
var MyDB		string
var username	string
var password	string
var testName	string


// config structures
type DatabaseConfig struct {
    Host string `mapstructure:"server"`
    User string `mapstructure:"username"`
    Pass string `mapstructure:"password"`
    DBName string `mapstructure:"DBName"`
}

type probes struct {
    Name    string `mapstructure:"name"`
    Num     int  `mapstructure:"num"`
}

type Config struct {
    Db DatabaseConfig `mapstructure:"database"`
    Probe []probes
}

var c Config


// initalize and open serial port to amp
func initTCamp(devName *string) {

	var GHE_Index = 0
	log.Println("+++init TCamp reader")
	ports, err := enumerator.GetDetailedPortsList()
    if err != nil {
        log.Fatal(err)
    }
    if len(ports) == 0 {
        log.Println("No serial ports found!")
        return
    }

    for i, port := range ports {
        log.Printf("Found port: %s %d\n", port.Name, i)
        if port.IsUSB {
            log.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
            log.Printf("   USB serial %s\n", port.SerialNumber)
        }
        if *devName == port.Name {
            log.Println("***matched***")
            GHE_Index = i
        }
    }
	if GHE_Index == 0 {
		log.Fatal(" no matching port ")
	}

    mode := &serial.Mode{
        BaudRate: 115200,
        Parity:   serial.NoParity,
        DataBits: 8,
        StopBits: serial.OneStopBit,
    }
    port, err := serial.Open(ports[GHE_Index].Name, mode)
    if err != nil {
        log.Fatal(err)
    }

	TCamp_1.serialPort = port
    TCamp_1.serialNum  = ports[GHE_Index].SerialNumber

	// report what we've learned
	log.Printf("port:%s",ports[GHE_Index].Name)
	log.Printf("USB ID:%s:%s",ports[GHE_Index].VID,ports[GHE_Index].PID)
	log.Printf("USB SN:%s",TCamp_1.serialNum)
	log.Println("--TComp init complete")
}

// use influxDB to store TC readings
func initDB() {
    var err error

    // Create a new HTTPClient
    HTTPClient, err = client.NewHTTPClient(client.HTTPConfig{
        Addr:     "http://"+c.Db.Host+":8086",
        Username: c.Db.User,
        Password: c.Db.Pass,
    })
    if err != nil {
        log.Fatal(err)
    }

}

func readLoop() {

    buff := make([]byte,4096)
    cumRecved := 0

    for {
        n, err := TCamp_1.serialPort.Read(buff[cumRecved:cap(buff)])
        if err != nil {
            log.Fatal(err)
            break
        }
        if n == 0 {
        log.Println("\nEOF")
            break
        }
		// find all substrings delimited by \r\n
		if *dbgLvl > 0 { log.Printf("n:%d cum:%d",n,cumRecved) }
		// log.Printf("%X\n", string(buff[cumRecved:n+cumRecved]))
		if *dbgLvl >3 { log.Printf("%s\n", string(buff[cumRecved:n+cumRecved])) }
		lines := strings.Split(string(buff[cumRecved:n+cumRecved]), "\r\n")
		if *dbgLvl >1 { log.Printf("line coount:%d",len(lines)) }
		// log.Printf("%#v",lines)
		for i := range  lines {
			if *dbgLvl >2 { log.Printf("line:%s",lines[i]) }
			matched := TCampReading.Match([]byte(lines[i]))
			if *dbgLvl >3 { log.Println(matched) }
			if matched {
				var submatches [][]byte
				submatches = TCampReading.FindSubmatch([]byte(lines[i]))
				probe, _  := strconv.Atoi(string(submatches[probeNumIndex]))
				reading, _  := strconv.ParseFloat(string(submatches[readingIndex]), 32)
				_, probeCFGed := TCbyNum[probe]
				probeName	  := TCbyNum[probe].name
				if probeCFGed {
					// log.Printf("probe:%d reading:%f name:%s", probe, reading, probeName)
					// insert the probe number and the reading for that probe into the readings DB 
					insertReading(float32(reading), probe, probeName)
				} else {
					log.Printf("probe:%d reading:%f", probe, reading)
					log.Println("not confgured")
				}
			}
		}
//		break
	}
}

//  insert probe # and readings into DB
func insertReading( reading float32 , probe int, probeName string) {

	var tags map[string]string
	var fields map[string]interface{}

	tags = map[string]string{
		"probe_name": probeName,
		"test_name":  testName,
	}
	fields = map[string]interface{} {
		"temp": reading,
		"probe": probe,
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
        Database:  c.Db.DBName,
        Precision: "s",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a point and add to batch
//    tags = map[string]string{"location": rcvSensorM.Location,
//                  "DSType": rcvSensorM.DataStreamType,
//  }
//
//    fields = map[string]interface{}{
//        "temp":   rcvSensorM.Temp,
//        "hum":    rcvSensorM.RH,
//    }

    pt, err := client.NewPoint("TC_reading", tags, fields, time.Now())
    if err != nil {
        log.Fatal(err)
    }
    bp.AddPoint(pt)

    // Write the batch
    if err := HTTPClient.Write(bp); err != nil {
        log.Fatal(err)
    }

	// log.Println(". point written")

}

//  load probe names from a TOML table
func loadProbeConfig() (err error){

	if viper.IsSet("probe") == false {
		log.Printf("no TCs configured")
		return(errors.New("no probes configured"))
	}
	TCbyNum = make(map[int]TCProbe)
    // enumerate over all the probes configured
    for i, p := range c.Probe {
        log.Printf("probe:%d, name:%s entry:%d",p.Num,p.Name,i)
		t := TCbyNum[p.Num]
		t.name = p.Name
		t.ampPortNum = p.Num
		TCbyNum[p.Num] = t
		// TCbyNum[p.Num].name = p.Name
    }
	return(nil)
}

// retrieve general config informoation
func loadConfig() (err error){

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Config file error:%S",err)
		return(err)
	}
	configTitle = viper.GetString("title")
	log.Printf("%s",configTitle)
	testName = viper.GetString("test_name")
//	MyDB = viper.GetString("DBName")
//	influxHostName = viper.GetString("server")
//	username = viper.GetString("username")
//	password = viper.GetString("password")

    if err := viper.Unmarshal(&c); err != nil {
        log.Printf("couldn't read config: %s", err)
    }

	return (nil)

}

//  print various config items
func printConfig() {

    log.Printf("%s",configTitle)
    log.Printf("Test Name:%s  DB Name:%s Host Name:%s  User ID:%s Pass:%s",testName,c.Db.DBName,c.Db.Host,c.Db.User,c.Db.Pass)
}

func main() {
	var err error

	flag.Parse()
	log.SetFlags(0)
	err = loadConfig()
	if err != nil {
		os.Exit(3)
	}
	err = loadProbeConfig()
	if err != nil {
		os.Exit(3)
	}
	initDB()
	log.Printf("*** TCamp Reader %s starting ***",ourVersion)
	initTCamp(devName)
	readLoop()
}
