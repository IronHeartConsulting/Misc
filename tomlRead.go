//
//   test reading and parsing TOML config file with viper
//

package main

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/viper"
)

var ourVersion string = "V0.3"

type TCProbe struct {
	name	string
	ampPortNum	int
	lastReading float32
}

var TCs [8]TCProbe

var configTitle string
var testName	string

type DatabaseConfig struct {
	Host string `mapstructure:"server"`
	User string `mapstructure:"username"`
	Pass string `mapstructure:"password"`
	DBName string `mapstructure:"DBName"`
}

type probes struct {
	Name	string `mapstructure:"name"`
	Num		int  `mapstructure:"num"`
}

type Config struct {
	Db DatabaseConfig `mapstructure:"database"`
	Probe []probes
}

var c Config

//  load probe names from a TOML table
func loadProbeConfig() (err error){

	if viper.IsSet("probe") == false {
		log.Printf("no TCs configured")
		return(nil)
	}
	// enumerate over all the probes configured
	for i, p := range c.Probe {
		log.Printf("probe:%d, name:%s entry:%d",p.Num,p.Name,i)
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
	testName = viper.GetString("test_name")
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

	log.Printf("*** Viper Config read and parse %s***",ourVersion)
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
	printConfig()
}
