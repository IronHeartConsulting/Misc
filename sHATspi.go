//
//   sHATspi - inerface to eeram via SPI bus
//
package main

import (
	"fmt"
	"log"
	"encoding/binary"
	"bufio"
	"os"
	"strings"
	"io/ioutil"

	"periph.io/x/conn/physic"
	"periph.io/x/conn/spi"
	"periph.io/x/conn/spi/spireg"
	"periph.io/x/host"

	"github.com/davecgh/go-spew/spew"
)

var c	spi.Conn
const	eeramSize		= 32*1024

//  read and print status register
func readSR() {
	var write	[]byte
	var read	[]byte
	var err		error

	write = []byte{0x05, 0x00}
	read = make([]byte, len(write))
	if err = c.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	// Use read.
	fmt.Printf("SR:%02X\n", read[1:])
}

//   write status register
func writeSR(sr byte) {
	var write	[]byte
	var err		error

	write = []byte{0x01, 0x00}
	write[1] = sr
	if err = c.Tx(write,nil); err != nil {
		log.Fatal(err)
	}
	spew.Dump(write, sr)
}
// write enable latch
func WEL() {
	var write	[]byte
	var err		error
	// enable writes to SRAM
	write = []byte{0x06}
	if err = c.Tx(write, nil); err != nil {
		log.Fatal(err)
	}
}


// write disable (turn off write enanble latch)
func WRDI() {
	var write	[]byte
	var err		error
	// enable writes to SRAM
	write = []byte{0x04}
	if err = c.Tx(write, nil); err != nil {
		log.Fatal(err)
	}
}


// write to SRAM - input:address, byte to write
func wrSRAM(addr uint16, data byte) (error){
	var write	[]byte
	var err		error

	addrBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(addrBuf, addr)
	// spew.Dump(addr, addrBuf)
	if addr > eeramSize -1 { // beyond limit of array size
		fmt.Printf("addr:%d larger than array:%d\n",addr, eeramSize)
		// TODO return more meaningfull error
		return nil
	}
	write = make([]byte, 4)
	write[0] = 0x02
	write[1] = addrBuf[0]
	write[2] = addrBuf[1]
	write[3] = data
	if err = c.Tx(write, nil); err != nil {
		log.Fatal(err)
	}
	return err
}

// read SRAM - input: address, return: byte of data at that address
func rdSRAM(addr uint16) (byte) {
	var write	[]byte
	var read	[]byte
	var err		error

	addrBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(addrBuf, addr)
	// spew.Dump(addr, addrBuf)
	if addr > eeramSize -1 { // beyond limit of array size
		fmt.Printf("addr:%d larger than array:%d\n",addr, eeramSize)
		return 0
	}
	write = []byte{0x03, 0x00, 0x00, 0x00}
	write[1] = addrBuf[0]
	write[2] = addrBuf[1]
	read = make([]byte, len(write))
	if err = c.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	// Use read.
	// data = read[1:]
	// spew.Dump(read)
	data := read[len(read)-1]
	// spew.Dump(data)
	return data
}

// write a []byte to the SRAM
func wrBufSRAM(addr uint16, dataBuf []byte) {
	fmt.Printf("write %d bytes to %d\n",len(dataBuf),addr)
	// we usee continuous writes and increment past page boundary.
	//  if the address and the file size is beyond the end of the array
	//  the array will wrap to address 0.
	//  we never accept a buffer larger than the array, so we won't over-write our own buffer
	// enable PRO mode (page roll over mode)
	WEL()
	writeSR(0x20)
	readSR()
}


func main() {

	var	err		error
	var addr	uint16
	var data	byte

	// Make sure periph is initialized.
	if _, err = host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use spireg SPI port registry to find the first available SPI bus.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	c, err = p.Connect(physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(c)

	readSR()
	// read test locations, before any write
	fmt.Printf("--- prior to test writes ---\n")
	addr = 0
	data = rdSRAM(addr)
	fmt.Printf("read 0x%02x:0x%02X\n",addr,data)
	addr = 1 * 1024
	data = rdSRAM(addr)
	fmt.Printf("read 0x%02x:0x%02X\n",addr,data)
	fmt.Printf("---\n")

	WEL()
	readSR()
	addr = 1 * 1024
	wrSRAM(addr, 0xad)
	readSR()
	data = rdSRAM(addr)
	fmt.Printf("read 0x%02x:0x%02X\n",addr,data)
	addr = 0
	data = rdSRAM(addr)
	fmt.Printf("read 0x%02x:0x%02X\n",addr,data)

	// now, lets get a file name
	rdrsin := bufio.NewReader(os.Stdin)
	fmt.Printf("file name?:")
	fName, _ := rdrsin.ReadString('\n')
	fName = strings.TrimSuffix(fName, "\n")
	fInfo, err := os.Stat(fName)
	if err != nil {
		fmt.Printf("file error:%s\n",err)
		return
	}
	fSize := fInfo.Size()
	if fSize > eeramSize {
		fmt.Printf("file size:%d larger than EERAM capacity\n",fSize)
		return
	}
	fmt.Printf("file to be read:%s\n",fName)
	fileBuf, err := ioutil.ReadFile(fName)
	if err != nil {
		fmt.Printf("file read error:%s\n",err)
		return
	}
	fmt.Printf("file buffer len:%d\n",len(fileBuf))
	addr = 0
	wrBufSRAM(addr, fileBuf)
}
