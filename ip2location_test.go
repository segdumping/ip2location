package ip2location

import "testing"

func BenchmarkIp2Location_MemorySearch(b *testing.B) {
	location, err := New("./data/ip2location.bin")
	if err != nil {
		b.Error(err)
	}

	for i := 0; i <b.N; i++ {
		location.MemorySearch("127.0.0.1")
	}
}
