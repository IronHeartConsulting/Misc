package main

import (
	"fmt"
	"net"
	"regexp"
)

// var IFregexp = regexp.MustCompile(`IF(\d.)     (\+|\-)(\d.)(\d)(\d)(\d)\d\d\d\d\d ;`)
var IFregexp = regexp.MustCompile(`IF(\d{11})\s{5}([\+\-])(\d{4})(\d)(\d)\s0{2}(\d)(\d)0(\d)(\d)(\d)(\d)1 ;`)

func parseIF(buf []byte, n int) {
	fmt.Printf("from K4 CAT: %s\n", buf[:n])
	// fmt.Printf("%q\n", IFregexp.FindAllSubmatch(buf[:n], -1))
	matches := IFregexp.FindAllSubmatch(buf[:n], -1)
	// fmt.Printf("%T\n", matches)
	// spew.Dump(matches)
	printM(matches)

}

func printM(matches [][][]byte) {
	FA := matches[0][1]
	fmt.Printf("freq:%s\n", FA)
	signOffset := matches[0][2]
	fmt.Printf("RIT/XIT sign:%s\n", string(signOffset))
	offsetHz := matches[0][3]
	fmt.Printf("offset (Hz):%s\n", offsetHz)
	ritStatus := matches[0][4]
	xitStatus := matches[0][5]
	radioTx := matches[0][6]
	radioMode := matches[0][7]
	scanning := matches[0][8]
	split := matches[0][9]
	rspFormat := matches[0][10]
	fmt.Printf("ritStatus:       %s\n", ritStatus)
	fmt.Printf("xitStatus:       %s\n", xitStatus)
	fmt.Printf("radio TX:        %s\n", radioTx)
	fmt.Printf("Mode:            %s\n", radioMode)
	fmt.Printf("scanning:        %s\n", scanning)
	fmt.Printf("split:           %s\n", split)
	fmt.Printf("RESP format:     %s\n", rspFormat)
}

func main() {

	fmt.Printf("regexp:%s\n", IFregexp)

	conn, err := net.Dial("tcp", "192.168.21.206:9200")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buf := make([]byte, 1024)

	FAreq := []byte("FA;")
	_, err = conn.Write(FAreq)
	if err != nil {
		panic(err)
	}

	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("from K4 CAT: %s\n", buf[:n])

	K41Mode := []byte("K41;")
	_, err = conn.Write(K41Mode)
	if err != nil {
		panic(err)
	}

	// enabled Ai1 -> VFO, RIT
	AI1req := []byte("AI1;")
	_, err = conn.Write(AI1req)
	if err != nil {
		panic(err)
	}

	for {
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		parseIF(buf, n)
	}

}
