package ip2location

import (
"errors"
"strconv"
"strings"
"io/ioutil"
)

const (
	IndexBlockLength = 12
)

type Ip2Location struct {
	// super block index info
	firstIndexPtr int64
	lastIndexPtr  int64
	totalBlocks   int64

	// for memory mode only
	// the original db binary string

	dbBinStr []byte
	dbFile   string
}

type IpInfo struct {
	Country  string
	Province string
}

func (ip IpInfo) String() string {
	return ip.Country + "|" + ip.Province
}

func getIpInfo(line []byte) IpInfo {
	lineSlice := strings.Split(string(line), sep)
	ipInfo := IpInfo{}
	length := len(lineSlice)
	if length == 1 {
		ipInfo.Country = lineSlice[0]
	} else if length >= 2 {
		ipInfo.Country = lineSlice[0]
		ipInfo.Province = lineSlice[1]
	}

	return ipInfo
}

func New(path string) (*Ip2Location, error) {
	var err error
	this := &Ip2Location{dbFile: path}
	this.dbBinStr, err = ioutil.ReadFile(this.dbFile)
	if err != nil {
		return nil, err
	}

	this.firstIndexPtr = getLong(this.dbBinStr, 0)
	this.lastIndexPtr = getLong(this.dbBinStr, 4)
	this.totalBlocks = (this.lastIndexPtr-this.firstIndexPtr)/IndexBlockLength + 1

	return this, nil
}

func (this *Ip2Location) MemorySearch(ipStr string) (ipInfo IpInfo, err error) {
	ipInfo = IpInfo{}
	ip, err := ip2long(ipStr)
	if err != nil {
		return ipInfo, err
	}

	h := this.totalBlocks
	var dataPtr, l int64
	for l <= h {
		m := (l + h) >> 1
		p := this.firstIndexPtr + m*IndexBlockLength
		sip := getLong(this.dbBinStr, p)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(this.dbBinStr, p+4)
			if ip > eip {
				l = m + 1
			} else {
				dataPtr = getLong(this.dbBinStr, p+8)
				break
			}
		}
	}
	if dataPtr == 0 {
		return ipInfo, errors.New("not found")
	}

	dataLen := (dataPtr >> 24) & 0xFF
	dataPtr = dataPtr & 0x00FFFFFF
	ipInfo = getIpInfo(this.dbBinStr[(dataPtr):dataPtr+dataLen])
	return ipInfo, nil

}

func getLong(b []byte, offset int64) int64 {
	val := int64(b[offset]) |
		int64(b[offset+1])<<8 |
		int64(b[offset+2])<<16 |
		int64(b[offset+3])<<24

	return val
}

func ip2long(IpStr string) (int64, error) {
	bits := strings.Split(IpStr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip format error")
	}

	var sum int64
	for i, n := range bits {
		bit, _ := strconv.ParseInt(n, 10, 64)
		sum += bit << uint(24-8*i)
	}

	return sum, nil
}
