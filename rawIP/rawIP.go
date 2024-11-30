package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"math/bits"
	"math/rand"
	"net"
	"os"
	"time"
)

var mode = flag.String("mode", "client", "client | server mode")
var serIP = flag.String("serverIP", "127.0.0.1", "ip adddress of server process")
var blkCnt = flag.Int("cnt", 10, "Count of block to send")
var blksize = flag.Int("bsize", 1024, "size of each block")
var debug = flag.Bool("debug", false, "debug true|false")

var protocol = "61"
var clconn net.Conn

// cmd list - reset; print stats; exit

type cmd int

const (
	CmdReset cmd = iota + 1
	CmdPrints
	CmdExit
)

// rawP -mode server | client -ServerIP <IP addr> -cnt <cnt of blocks to send? -bsize <size of each block>

func encodeUint(x uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, x)
	return buf[bits.LeadingZeros64(x)>>3:]
}

func main() {

	flag.Parse()

	fmt.Println("raw IP xfer. Version", "V0.7")
	if *mode == "server" {
		serverMode()
	} else { //assume client mode
		clientMode()
	}
}

func serverMode() {
	fmt.Printf("server mode\n")
	// netaddr, _ := net.ResolveIPAddr("ip4", "127.0.0.1")
	// conn, _ := net.ListenIP("ip4:"+protocol, netaddr)
	priorBlkNum := -1
	blocksRecved := 0
	OOS := 0
	crcErr := 0
	conn, _ := net.ListenIP("ip4:"+protocol, nil)
	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("server read error", err)
			os.Exit(4)
		}
		if *debug {
			fmt.Printf("Read len:%d\n% X\n", n, buf[0:20])
		}
		blknum := binary.LittleEndian.Uint16(buf[0:2])
		if blknum == 0xffff {
			cmd := buf[2]
			switch cmd {
			case 1:
				fmt.Print("Reset")
				priorBlkNum = -1
				blocksRecved = 0
				OOS = 0
				crcErr = 0
			case 2:
				fmt.Println()
				fmt.Printf("block count:        %d\n", blocksRecved)
				fmt.Printf("out of order blocks:%d\n", OOS)
				fmt.Printf("blocks with CRC err:%d\n", crcErr)
			case 3:
				fmt.Println("Exit")
				os.Exit(0)
			}
			continue
		}

		// only count blocks that are data
		fmt.Print(".")
		blocksRecved = blocksRecved + 1

		if priorBlkNum+1 != int(blknum) {
			fmt.Printf("out of seq. Expected:%d\n", priorBlkNum+1)
			OOS = OOS + 1
		}
		if *debug {
			fmt.Printf("seq num:%d\n", blknum)
		}
		priorBlkNum = priorBlkNum + 1
		ieeeChecksum := crc32.ChecksumIEEE(buf[6 : n-1])
		bufCkSum := binary.BigEndian.Uint32(buf[2:6])
		if *debug {
			fmt.Printf("crc32: % X\n", ieeeChecksum)
			fmt.Printf("check sum from client side:% X\n", bufCkSum)
		}
		if bufCkSum != ieeeChecksum {
			fmt.Printf("client checksum differs from server computed\n")
			crcErr = crcErr + 1
		}
	}
}

func clientMode() {
	var err error

	fmt.Printf("client mode. IP:%s\n", *serIP)
	clconn, err = net.Dial("ip4:"+protocol, *serIP)
	if err != nil {
		panic(err)
	}
	sendCMD(CmdReset)
	buf := make([]byte, *blksize)
	rand.Seed(time.Now().UnixNano())

	for blockNum := 0; blockNum < *blkCnt; blockNum++ {
		blknum := make([]byte, 2)
		binary.LittleEndian.PutUint16(blknum, uint16(blockNum))
		rand.Read(buf)
		copy(buf, blknum)
		ieeeChecksum := crc32.ChecksumIEEE(buf[6 : *blksize-1])
		crcByte := make([]byte, 4)
		binary.BigEndian.PutUint32(crcByte, ieeeChecksum)
		copy(buf[2:6], crcByte)
		clconn.Write(buf)
		fmt.Printf("block interation:%d block number:%v block number type:%T\n", blockNum, blknum, blknum)
		fmt.Printf("crc32:% X type:%T\n", ieeeChecksum, ieeeChecksum)
		fmt.Printf("Len buf:%d\n", len(buf))
	}
	sendCMD(CmdPrints)
	sendCMD(CmdExit)

}

func sendCMD(Cmd2send cmd) {
	fmt.Printf("sending cmd:%d\n", Cmd2send)
	CMdBuf := make([]byte, 3)
	CMdBuf[0] = 0xff
	CMdBuf[1] = 0xff
	switch Cmd2send {
	case CmdReset:
		CMdBuf[2] = byte(CmdReset)
	case CmdPrints:
		CMdBuf[2] = byte(CmdPrints)
	case CmdExit:
		CMdBuf[2] = byte(CmdExit)
	}
	clconn.Write(CMdBuf)
}
