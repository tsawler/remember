package remember

import (
	"github.com/alicebob/miniredis/v2"
	"os"
	"testing"
)

var testRedis *miniredis.Miniredis
var testCache *Cache

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	testRedis = s

	ops := Options{
		Server: testRedis.Host(),
		Port:   testRedis.Port(),
		Prefix: "test_cache",
	}
	testCache = New(ops)

	os.Exit(m.Run())
}
