// multi channel MAX31856 TC amp.   One SPI bus, for /CS
//
//   board is playwithfusion SEN-30008 screw terminal

package main

import (
	// "context"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mvpninjas/go-bitflag"
	"github.com/spf13/viper"
	max31856 "github.com/the-sibyl/goMAX31856"
)

var ourVersion string = "V0.5"

//  TODO  convert to map, indexed by probe number from amp
type TCProbe struct {
	name        string
	ampPortNum  int
	cs_name     string
	active      bool
	channel     max31856.MAX31856
	TCtype      string
	lastReading float32 // used for display
}

var TCbyNum map[int]TCProbe

var TCs [4]TCProbe

var dbClient influxdb2.Client
var writeAPIx api.WriteAPIBlocking

var configTitle string
var MyDB string
var testName string

// var ch0 max31856.MAX31856

// config structures
type DatabaseConfig struct {
	Host      string `mapstructure:"server"`
	User      string `mapstructure:"username"`
	Pass      string `mapstructure:"password"`
	DBName    string `mapstructure:"DBName"`
	AuthTKN   string `mapstructure:"authToken"`
	DelayTime int    `mapstructure:"delayTime"`
	ServerURL string `mapstructure:"serverURL"`
	ORGName   string `mapstructure:"org"`
}

type probes struct {
	Name   string `mapstructure:"name"`
	Num    int    `mapstructure:"num"`
	CS     string `mapstructure:"cs"`
	Active bool   `mapstructure:"active"`
	TCtype string `mapstructure:"type"`
}

type Config struct {
	Db    DatabaseConfig `mapstructure:"database"`
	Probe []probes
}

var c Config

// use influxDB to store TC readings
func initDB() {
	// create new client with default option for server URL authenticate by token
	dbClient = influxdb2.NewClient(c.Db.ServerURL, c.Db.AuthTKN)

	// writeAPI expects bucket nam - we use DBName for bucket name
	writeAPIx = dbClient.WriteAPIBlocking(c.Db.ORGName, c.Db.DBName)

}

func initTC() {
	const (
		CNVAVG_16 = 8 << 3
		TC_K      = 3
		TC_J      = 2
	)
	var spiClockSpeed int64 = 100000
	// var ch0 max31856.MAX31856
	var err error
	// devPathCh0 := "/dev/spidev0.0"
	var cr1 bitflag.Flag
	devPathChRoot := "/dev/spidev0."
	timeoutPeriod := time.Second
	fmt.Println("starting TC init")
	for i, _ := range TCs {
		log.Printf("index:%d probe name:%s cs value:%s TC Type:%s\n", i, TCs[i].name, TCs[i].cs_name, TCs[i].TCtype)
		if !TCs[i].active {
			log.Printf("not active, skipping\n")
			continue
		}
		devPathCh := devPathChRoot + TCs[i].cs_name
		TCs[i].channel, err = max31856.Setup(devPathCh, spiClockSpeed, 0, timeoutPeriod)
		if err != nil {
			log.Printf("TC failed setup:%s %s\n", TCs[i].name, err)
			TCs[i].active = false
		} else {
			log.Printf("spi channel init complete for:%s dev name:%s\n", TCs[i].name, devPathCh)
			// set probe type
			bitflag, _ := TCs[i].channel.GetFlags(max31856.CR1_RD)
			log.Printf("config reg 1:%#[1]x\n", bitflag)
			cr1.Set(CNVAVG_16)
			switch TCs[i].TCtype {
			case "k":
				cr1.Set(TC_K)
			case "j":
				cr1.Set(TC_J)
			default:
			}
			TCs[i].channel.SetFlags(max31856.CR1_WR, cr1)
			bitflag, _ = TCs[i].channel.GetFlags(max31856.CR1_RD)
			log.Printf("config reg 1:%#[1]x\n", bitflag)
		}

	}
}

//  insert probe # and readings into DB:q
func insertReading(reading float32, probe int, probeName string) error {

	var tags map[string]string
	var fields map[string]interface{}

	tags = map[string]string{
		"probe_name": probeName,
		"test_name":  testName,
	}
	fields = map[string]interface{}{
		"temp":  reading,
		"probe": probe,
	}

	// log.Printf("tags:%v\n", tags)
	// log.Printf("fields%v\n", fields)

	// Create a point and add to batch
	//    tags = map[string]string{"location": rcvSensorM.Location,
	//                  "DSType": rcvSensorM.DataStreamType,
	//  }
	//
	//    fields = map[string]interface{}{
	//        "temp":   rcvSensorM.Temp,
	//        "hum":    rcvSensorM.RH,
	//    }

	pt := influxdb2.NewPoint("TC_reading", tags, fields, time.Now())
	// spew.Dump(pt)

	// Write the batch
	err := writeAPIx.WritePoint(context.Background(), pt)
	return err

}

//  load probe names from a TOML table
func loadProbeConfig() (err error) {

	if !viper.IsSet("probe") {
		log.Printf("no TCs configured")
		return (errors.New("no probes configured"))
	}
	TCbyNum = make(map[int]TCProbe)
	// enumerate over all the probes configured
	for i, p := range c.Probe {
		log.Printf("probe:%d, name:%s entry:%d", p.Num, p.Name, i)
		TCs[i].ampPortNum = p.Num
		TCs[i].name = p.Name
		TCs[i].cs_name = p.CS
		TCs[i].active = p.Active
		if p.TCtype != "" {
			TCs[i].TCtype = p.TCtype
		} else {
			TCs[i].TCtype = "k"
		}
	}
	return (nil)
}

// retrieve general config informoation
func loadConfig() (err error) {

	viper.SetConfigName("TCProbe")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Config file error:%S", err)
		return (err)
	}
	configTitle = viper.GetString("title")
	log.Printf("%s", configTitle)
	testName = viper.GetString("test_name")

	if err := viper.Unmarshal(&c); err != nil {
		log.Printf("couldn't read config: %s", err)
	}
	log.Printf("config:%+v\n", c)

	return (nil)

}

func main() {
	var err error
	var tempC float32

	err = loadConfig()
	if err != nil {
		os.Exit(3)
	}
	err = loadProbeConfig()
	if err != nil {
		os.Exit(3)
	}
	initDB()
	// spew.Dump(writeAPIx)
	log.Printf("*** MC_MAX31586 Reader %s starting ***", ourVersion)
	initTC()
	for {
		for i, _ := range TCs {
			if !TCs[i].active {
				fmt.Printf(" --.- ")
				continue
			}
			fault, _ := TCs[i].channel.CheckForFaults()
			if !fault {
				tempC, _ = TCs[i].channel.GetTempOnce()
				TCs[i].lastReading = tempC
				err = insertReading(tempC, i, TCs[i].name)
				// log.Printf("TC reading:%f probe name:%s index:%d", tempC, TCs[i].name, i)
				fmt.Printf(" %.1f ", tempC)
				if err != nil {
					log.Printf("DB insert failure:%s\n", err)
				}
			} else {
				faultErr, _ := TCs[i].channel.GetFlags(max31856.SR_RD)
				TCs[i].active = false
				log.Printf("TC fault:%#[1]x channel disabled\n", faultErr)
			}
		}
		fmt.Printf("\r")
		time.Sleep(time.Duration(c.Db.DelayTime) * time.Millisecond)
	}
}
