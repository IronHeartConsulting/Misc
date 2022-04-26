// multi channel MAX31856 TC amp.   One SPI bus, for /CS
//
//   board is playwithfusion SEN-30008 screw terminal

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/viper"
	max31856 "github.com/the-sibyl/goMAX31856"
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
var HTTPClient client.Client
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
	var err error

	// Create a new HTTPClient
	HTTPClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://" + c.Db.Host + ":8086",
		Username: c.Db.User,
		Password: c.Db.Pass,
	})
	if err != nil {
		log.Fatal(err)
	}

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
	var spiClockSpeed int64 = 100000
	devPathCh0 := "/dev/spidev0.0"
	timeoutPeriod := time.Second

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

	ch0, err := max31856.Setup(devPathCh0, spiClockSpeed, 16, timeoutPeriod)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("??", max31856.CJLF_WR)

	for {
		temperature, _ := ch0.GetTempOnce()
		fmt.Println("temp:", temperature)
		insertReading(float32(temperature), 1, TCbyNum[1].name)
		time.Sleep(500 * time.Millisecond)
	}
}
