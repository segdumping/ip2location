package ip2location

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	sep = "|"
	SuperBlockLength = 8
)

//ip2location csv文件压缩
//首部预留 8 bytes 的SUPER BLOCK
//依据每一条记录的起始ip, 结束ip和数据，生成一个index block， 前四个字节存储起始ip, 中间四个字节存储结束ip, 后四个字节存储已经计算出的数据地址，并暂存到INDEX区
//当 INDEX 索引区和 DATA 数据区确定下来之后，再把 INDEX 的起始位置存储到 SUPER BLOCK 的前四个字节，结束位置存储到 SUPER BLOCK 的后四个字节
func Compress(sourcePath, destPath string) error {
	inFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var datas []byte
	var dataIndex = SuperBlockLength
	var lineIndex int
	var indexs []byte
	dataExist := make(map[string]int)
	rd := bufio.NewReader(inFile)

	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		lineIndex = dataIndex
		start, end, info, err := parseLine(line)
		if err != nil {
			return err
		}

		var found bool
		if v, ok := dataExist[info]; ok {
			lineIndex = v
			found = true
		}

		block := makeIndexBlock(start, end, lineIndex, info)
		indexs = append(indexs, block...)
		if !found {
			dataExist[info] = dataIndex
			dataIndex += len([]byte(info))
			datas = append(datas, []byte(info)...)
		}
	}

	dataLen := len(datas)
	indexLen := len(indexs)
	superBlock := makeSuperBlock(SuperBlockLength+dataLen, SuperBlockLength+dataLen+indexLen)
	outFile.Write(superBlock)
	outFile.Write(datas)
	outFile.Write(indexs)

	return nil
}

func makeSuperBlock(startIndex, endIndex int) []byte {
	block := make([]byte, 8)
	write4Byte(block, startIndex)
	write4Byte(block[4:], endIndex)

	return block
}

func makeIndexBlock(startIP, endIP, index int, data string) []byte {
	block := make([]byte, 12)
	write4Byte(block, startIP)
	write4Byte(block[4:], endIP)
	write3Byte(block[8:], index)
	dataLen := len(data)
	write1Byte(block[11:], dataLen)

	return block
}

func write4Byte(b []byte, v int) {
	_ = b[3]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func write3Byte(b []byte, v int) {
	_ = b[2]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

func write1Byte(b []byte, v int) {
	_ = b[0]
	b[0] = byte(v)
}

func parseLine(line string) (ipStart, ipEnd int, data string, err error) {
	cutset := "\""
	var buffer bytes.Buffer
	ipInfo := strings.Split(line, ",")
	buffer.WriteString(strings.Trim(ipInfo[3], cutset))
	buffer.WriteString(sep)
	buffer.WriteString(strings.Trim(ipInfo[4], cutset))
	data = buffer.String()

	ipStr := strings.Trim(ipInfo[0], cutset)
	if ipStart, err = strconv.Atoi(ipStr); err != nil {
		return
	}

	ipStr = strings.Trim(ipInfo[1], cutset)
	if ipEnd, err = strconv.Atoi(ipStr); err != nil {
		return
	}

	return
}
