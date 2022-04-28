// multi channel MAX31856 TC amp.   One SPI bus, for /CS
//
//   board is playwithfusion SEN-30008 screw terminal

package main

import (
	// "context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/davecgh/go-spew/spew"
)

var ourVersion string = "V0.2"

//  TODO  convert to map, indexed by probe number from amp
type TCProbe struct {
	name       string
	ampPortNum int
	cs_name    string
	//	lastReading float32
}

var TCbyNum map[int]TCProbe

var TCs [4]TCProbe

// var TCamp_1 TCamp

var influxHostName string

var dbClient influxdb2.Client
// var writeAPIx  dbClient.WriteAPI
var writeAPIx  int
var configTitle string
var MyDB string
var username string
var password string
var testName string

// config structures
type DatabaseConfig struct {
	Host   string `mapstructure:"server"`
	User   string `mapstructure:"username"`
	Pass   string `mapstructure:"password"`
	DBName string `mapstructure:"DBName"`
}

type probes struct {
	Name string `mapstructure:"name"`
	Num  int    `mapstructure:"num"`
	CS   string `mapstructure:"cs"`
}

type Config struct {
	Db    DatabaseConfig `mapstructure:"database"`
	Probe []probes
}

var c Config

// use influxDB to store TC readings
func initDB() {
	const token = "sHMevZdU_7FoHVAnnt9jtCSrLqlwgvBautxWa8S-63cUyqsNsAdegQ8VFgNwhmBXl5MXwAo-q8iIipn824w5kg=="
	var xxx dbClient.WriteAPIBlocking

	// create new client with default option for server URL authenticate by token
	dbClient = influxdb2.NewClient("https://us-central1-1.gcp.cloud2.influxdata.com", token)
	writeAPIx := dbClient.WriteAPIBlocking("kevin.rowett@xconn-tech.com", "TC")
	log.Println("--- start of writeAPI Dump ---")
	spew.Dump(writeAPIx)
	log.Println("--- end of writeAPI Dump ---")
	log.Printf("var:%v\n",writeAPIx)
}

//  insert probe # and readings into DB
func insertReading(reading float32, probe int, probeName string) {

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

	log.Printf("tags:%v\n", tags)
	log.Printf("fields%v\n", fields)

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
	spew.Dump(pt)

	// Write the batch
	// if err := writeAPIx.WritePoint(context.Background(), pt); err != nil {
	//	log.Fatal(err)
	// }

	// log.Println(". point written")

}

//  load probe names from a TOML table
func loadProbeConfig() (err error) {

	if viper.IsSet("probe") == false {
		log.Printf("no TCs configured")
		return (errors.New("no probes configured"))
	}
	TCbyNum = make(map[int]TCProbe)
	// enumerate over all the probes configured
	for i, p := range c.Probe {
		log.Printf("probe:%d, name:%s entry:%d", p.Num, p.Name, i)
		t := TCbyNum[p.Num]
		t.name = p.Name
		t.ampPortNum = p.Num
		t.cs_name = p.CS
		TCbyNum[p.Num] = t
		// TCbyNum[p.Num].name = p.Name
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
	//	MyDB = viper.GetString("DBName")
	//	influxHostName = viper.GetString("server")
	//	username = viper.GetString("username")
	//	password = viper.GetString("password")

	if err := viper.Unmarshal(&c); err != nil {
		log.Printf("couldn't read config: %s", err)
	}

	return (nil)

}

func main() {
	var err error

	err = loadConfig()
	if err != nil {
		os.Exit(3)
	}
	err = loadProbeConfig()
	if err != nil {
		os.Exit(3)
	}
	initDB()
	log.Printf("*** MC_MAX31586 Reader %s starting ***", ourVersion)
}
