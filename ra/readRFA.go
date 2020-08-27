package main

import (
    "bytes"
    "log"
	"os"
	"flag"
	"time"
    "net/http"
	"io/ioutil"
	"encoding/xml"
	"strconv"

	"github.com/spf13/viper"
	"github.com/influxdata/influxdb1-client/v2"
)


var ourVersion string = "V0.2"

var dbgLvl = flag.Uint("debug",0,"debug level (0-3)")


var influxHostName string
var inflClient	client.Client
var configTitle string
var MyDB        string
var username    string
var password    string
var testName    string
var rfaReq		*http.Request
var	rfaClient	*http.Client


// config structures
type DatabaseConfig struct {
    Host string `mapstructure:"server"`
    User string `mapstructure:"username"`
    Pass string `mapstructure:"password"`
    DBName string `mapstructure:"DBName"`
}

type Config struct {
    Db DatabaseConfig `mapstructure:"database"`
}

var c Config

func parseMeterXML( bodyText []byte ) (string, float32, float32, float32) {

	type Result struct {
		// XMLName	xml.Name	`xml:"DeviceDetails"`
		XMLName	xml.Name	`xml:"Device"`
		HWaddr	string	`xml:"DeviceDetails>HardwareAddress"`
		VarValues	[]string	`xml:"Components>Component>Variables>Variable>Value"`
	}
	v := Result{}
	err := xml.Unmarshal(bodyText, &v)
	if err != nil {
		log.Printf("xml unmashall error: %v\n", err)
		return "", 0.0,  0.0,  0.0
	}
	// log.Printf("HW address %s\n",v.HWaddr)
	// log.Printf("Values: %v\n",v.VarValues)
	var0, _  := strconv.ParseFloat(v.VarValues[0], 32)
	var1, _  := strconv.ParseFloat(v.VarValues[1], 32)
	var2, _  := strconv.ParseFloat(v.VarValues[2], 32)
	return v.HWaddr, float32(var0), float32(var1),  float32(var2)
}

// use influxDB to store readings
func initDB() {
    var err error

    // Create a new influx HTTPClient
    inflClient, err = client.NewHTTPClient(client.HTTPConfig{
        Addr:     "http://"+c.Db.Host+":8086",
        Username: c.Db.User,
        Password: c.Db.Pass,
    })
    if err != nil {
        log.Fatal(err)
    }

}

// retrieve general config informoation
func loadConfig() (err error){

    viper.SetConfigName("readRFA")
    viper.AddConfigPath(".")
    err = viper.ReadInConfig()
    if err != nil {
        log.Printf("Config file error:%S",err)
        return(err)
    }
    configTitle = viper.GetString("title")
    log.Printf("%s",configTitle)
    testName = viper.GetString("test_name")
//  MyDB = viper.GetString("DBName")
//  influxHostName = viper.GetString("server")
//  username = viper.GetString("username")
//  password = viper.GetString("password")

    if err := viper.Unmarshal(&c); err != nil {
        log.Printf("couldn't read config: %s", err)
    }

    return (nil)

}

func readLoop() {

	var instantDemand float32
	var sumDelivered, sumReceived float32
	var HWAddr string
	var bodyText []byte
	var err error

	for {
		bodyText, err = readUnit()
		if err != nil {
			log.Printf("readUnit returned err:%v\n",err)
			return
		}
		HWAddr, instantDemand, sumDelivered, sumReceived = parseMeterXML(bodyText)
		err = nil
		err = insertReading(HWAddr, instantDemand, sumDelivered, sumReceived)
		if err != nil {
			log.Println("insertReaading returned err%s",err)
			// log.Println(HWAddr)
			return
		}
		log.Printf("instant demand:%f  sum total delivered:%f received:%f Used:%f\n",
					instantDemand,sumDelivered,sumReceived,sumDelivered-sumReceived)
		time.Sleep(5 * time.Second)
	}
}

func readUnit() ([]byte, error) {

    body := `<Command>
    <Name>device_query</Name>
    <DeviceDetails>
        <HardwareAddress>0x0013500200a196eb </HardwareAddress>
    </DeviceDetails>
    <Components>
        <Component>
            <Name>Main</Name>
                <Variables>
                    <Variable>
                        <Name>zigbee:InstantaneousDemand</Name>
                    </Variable>
                    <Variable>
                        <Name>zigbee:CurrentSummationDelivered</Name>
                    </Variable>
                    <Variable>
                        <Name>zigbee:CurrentSummationReceived</Name>
                    </Variable>
                </Variables>
        </Component>
    </Components>
</Command>`

    rfaClient := &http.Client{}
	defer rfaClient.CloseIdleConnections()
    // build a new request, but not doing the POST yet
    rfaReq, err := http.NewRequest("POST", "http://192.168.21.127/cgi-bin/post_manager/", bytes.NewBuffer([]byte(body)))
    if err != nil {
        log.Printf("New Request rtn err:%v\n",err)
		return nil,  err
    }
    rfaReq.Header.Add("Content-Type", "text/xml; charset=utf-8")
	rfaReq.SetBasicAuth("006d60","4b0190726b0bbe37")
    // now POST it
    resp, err := rfaClient.Do(rfaReq)
    if err != nil {
        log.Printf("readUnit: Do rtn err:%v\n",err)
		return nil, err
    }
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	return bodyText , err
}

func insertReading(HWAddr string, instantDemand float32, sumDelivered float32, sumReceived float32) (err error) {

	var tags map[string]string
	var fields map[string]interface{}

	tags = map[string]string {
		"meterAddr": HWAddr,
	}

	fields = map[string]interface{} {
		"demand": instantDemand,
		"delivered" : sumDelivered,
		"received" : sumReceived,
	}
	// log.Println("debug return")

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: c.Db.DBName,
		Precision: "s",
	})
	if err != nil {
		log.Printf("New Batch Points err:%v",err)
		return err
	}


    pt, err := client.NewPoint("meter_reading", tags, fields, time.Now())
    if err != nil {
        log.Printf("NewPoint retn err:%V\n",err)
		return err
    }

    bp.AddPoint(pt)

	// *** return nil

    // Write the batch
    if err := inflClient.Write(bp); err != nil {
        log.Printf("Write BP rtn err:%V\n",err)
		return err
    }

	return nil

}

func main() {
	var err error
	var HWaddr string
	var instantDemand, sumDelivered, sumReceived float32
	var bodyText []byte

	flag.Parse()
	log.SetFlags(0)
	err = loadConfig()
	if err != nil {
		os.Exit(3)
	}
	initDB()
	log.Printf("*** readFA: read rainforeest automation meter interface. Version%s ***",ourVersion)
	bodyText, err  = readUnit()
	if err != nil {
		log.Printf("readUnit err return:%v\n",err)
		return
	}
	HWaddr, instantDemand, sumDelivered, sumReceived = parseMeterXML(bodyText)
	log.Printf("meter HW address:%s\n",HWaddr)
	log.Printf("instant demand:%f  sum total delivered:%f received:%f\n",instantDemand,sumDelivered,sumReceived)

	readLoop()
}

