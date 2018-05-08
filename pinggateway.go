package pinggateway

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//ICMP 报文
type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

var icmp ICMP
var mayPoweroff bool = false

//PingGateway loop表示一直ping，返回false 表示网关不通
func PingGateway() bool {
	iaddres, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("PingGateway Error:" + err.Error())
		return false
	}
	var result string
	for _, addre := range iaddres {
		if add, ok := addre.(*net.IPNet); ok == true && add.IP.IsLoopback() == false && add.IP.To4() != nil {

			inetAddSlice := add.IP.Mask(add.Mask)
			inetAddSpiltList := strings.Split(inetAddSlice.String(), ".")
			ipLastValue := inetAddSpiltList[len(inetAddSpiltList)-1]
			ipLastValueInt, err := strconv.Atoi(ipLastValue)
			if err != nil {
				fmt.Println(err)
				return false
			}
			ipLastValueInt++
			ipLastValueInt = 1 //网关不是等于1 么？
			result = inetAddSpiltList[0] + "." + inetAddSpiltList[1] + "." + inetAddSpiltList[2] + "." + strconv.Itoa(ipLastValueInt)
			break
		}
	}
	//gateway
	raddr, _ := net.ResolveIPAddr("ip", result)
	ticker := time.Tick(time.Second * 2 * 5)
	icmp.Type = 8
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 0
	icmp.SequenceNum = 0
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = checkSum(buffer.Bytes())
	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, icmp)
	var netaddr *net.IPConn
	var neterr error
NEXTWAKEUP:
	for {
		if netaddr != nil {
			netaddr.Close()
		}
		netaddr, neterr = net.DialIP("ip4:icmp", &(net.IPAddr{IP: net.ParseIP("0.0.0.0")}), raddr)
		if neterr != nil {
			if checkPoweroff(true) {
				//ubuntu 下是pm-hibernate
				//centos 下是 systemctl hibernate
				exec.Command("systemctl hibernate", "").Run()
				exec.Command("pm-hibernate", "").Run()
				fmt.Println("pm-hibernate1:")
				time.Sleep(time.Minute)
			}
			fmt.Println("neterr:", neterr)
			goto NEXTWAKEUP
		}
		_, writeErr := netaddr.Write(buffer.Bytes())
		if writeErr != nil {
			if checkPoweroff(true) {
				//ubuntu 下是pm-hibernate
				//centos 下是 systemctl hibernate
				exec.Command("systemctl hibernate", "").Run()
				exec.Command("pm-hibernate", "").Run()
				fmt.Println("pm-hibernate2:")
				time.Sleep(time.Minute)
			}
			fmt.Println("writeErr:", writeErr)
			goto NEXTWAKEUP
		}
		netaddr.SetDeadline((time.Now().Add(time.Second * 20)))
		recv := make([]byte, 1024)
		_, readErr := netaddr.Read(recv)
		if readErr != nil {
			if checkPoweroff(true) {
				//ubuntu 下是pm-hibernate
				//centos 下是 systemctl hibernate
				exec.Command("systemctl hibernate", "").Run()
				exec.Command("pm-hibernate", "").Run()
				fmt.Println("pm-hibernate3:")
				//	time.Sleep(time.Minute)
			}
			goto NEXTWAKEUP
		} else {
			checkPoweroff(false)
		}
		<-ticker
	}
}

//连续2次ping不通才算. result 为true就是准备关机
func checkPoweroff(maybe bool) bool {
	if maybe == true {
		if mayPoweroff == true {
			return true
		}
		mayPoweroff = true
	} else {
		mayPoweroff = false
	}
	return false
}
func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)
}
