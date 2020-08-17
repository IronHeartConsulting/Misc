package main

import (
    "bytes"
    "fmt"
    "net/http"
	"io/ioutil"
	"encoding/xml"
)

func parseMeterXML( bodyText []byte ) (string, string, string, string) {
	type Result struct {
		// XMLName	xml.Name	`xml:"DeviceDetails"`
		XMLName	xml.Name	`xml:"Device"`
		HWaddr	string	`xml:"DeviceDetails>HardwareAddress"`
		VarValues	[]string	`xml:"Components>Component>Variables>Variable>Value"`
	}
	v := Result{}
	err := xml.Unmarshal(bodyText, &v)
	if err != nil {
		fmt.Printf("xml unmashall error: %v\n", err)
		return "", "", "", ""
	}
	fmt.Printf("HW address %s\n",v.HWaddr)
	fmt.Printf("Values: %v\n",v.VarValues)
	return v.HWaddr, v.VarValues[0], v.VarValues[1], v.VarValues[2]
}

func main() {
	var err error
	var HWaddr, instantDemand, sumDelivered, sumReceived string
    // or you can use []byte(`...`) and convert to Buffer later on
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

	// fmt.Println(body)
    client := &http.Client{}
    // build a new request, but not doing the POST yet
    req, err := http.NewRequest("POST", "http://192.168.21.127/cgi-bin/post_manager/", bytes.NewBuffer([]byte(body)))
    if err != nil {
        fmt.Println(err)
    }
    req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.SetBasicAuth("006d60","4b0190726b0bbe37")
    // now POST it
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
    }
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	HWaddr, instantDemand, sumDelivered, sumReceived = parseMeterXML(bodyText)
	fmt.Printf("meter HW address:%s\n",HWaddr)
	fmt.Printf("instant demand:%s  sum total delivered:%s received:%s\n",instantDemand,sumDelivered,sumReceived)
	// s := string(bodyText)
	// 's' is now a string version of an XML response
	// fmt.Println("XML formatted result:")
	// fmt.Println(s)
	// fmt.Println("---end---")
}

