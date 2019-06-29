package ip2location

import "testing"

func TestCompress(t *testing.T) {
	ipSource := "./data/IP2LOCATION-LITE-DB11.CSV"
	ipDest := "./data/ip2location.bin"
	err := Compress(ipSource, ipDest)
	if err != nil {
		t.Error(err)
	}
}