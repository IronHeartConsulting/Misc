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

// rawP -mode server | client -ServerIP <IP addr> -cnt <cnt of blocks to send? -bsize <size of each block>

func encodeUint(x uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, x)
	return buf[bits.LeadingZeros64(x)>>3:]
}

func main() {

	flag.Parse()

	protocol := "61"
	if *mode == "server" {
		fmt.Printf("server mode\n")
		// netaddr, _ := net.ResolveIPAddr("ip4", "127.0.0.1")
		// conn, _ := net.ListenIP("ip4:"+protocol, netaddr)
		priorBlkNum := -1
		conn, _ := net.ListenIP("ip4:"+protocol, nil)
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				fmt.Println("server read error", err)
				os.Exit(4)
			}
			fmt.Printf("Read len:%d\n% X\n", n, buf[0:20])
			blknum := binary.LittleEndian.Uint16(buf[0:2])
			if priorBlkNum+1 != int(blknum) {
				fmt.Printf("out of seq. Expected:%d\n", priorBlkNum+1)
			}
			fmt.Printf("seq num:%d\n", blknum)
			priorBlkNum = priorBlkNum + 1
			ieeeChecksum := crc32.ChecksumIEEE(buf[6 : n-1])
			fmt.Printf("crc32: % X\n", ieeeChecksum)
			bufCkSum := binary.BigEndian.Uint32(buf[2:6])
			fmt.Printf("check sum from client side:% X\n", bufCkSum)
			if bufCkSum != ieeeChecksum {
				fmt.Printf("client checksum differs from server computed\n")
			}
		}
	} else { //assume client mode
		fmt.Printf("client mode. IP:%s\n", *serIP)
		conn, err := net.Dial("ip4:"+protocol, *serIP)
		if err != nil {
			panic(err)
		}
		buf := make([]byte, *blksize)
		rand.Seed(time.Now().UnixNano())
		// pattern := []byte{0xde, 0xad, 0xbe, 0xef}

		for blockNum := 0; blockNum < *blkCnt; blockNum++ {
			blknum := make([]byte, 2)
			binary.LittleEndian.PutUint16(blknum, uint16(blockNum))
			rand.Read(buf)
			copy(buf, blknum)
			ieeeChecksum := crc32.ChecksumIEEE(buf[6 : *blksize-1])
			crcByte := make([]byte, 4)
			binary.BigEndian.PutUint32(crcByte, ieeeChecksum)
			copy(buf[2:6], crcByte)
			conn.Write(buf)
			fmt.Printf("block interation:%d block number:%v block number type:%T\n", blockNum, blknum, blknum)
			fmt.Printf("crc32:% X type:%T\n", ieeeChecksum, ieeeChecksum)
			fmt.Printf("Len buf:%d\n", len(buf))
		}

	}
}
