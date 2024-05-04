package remember

import (
	"github.com/alicebob/miniredis/v2"
	"log"
	"os"
	"testing"
)

var testRedis *miniredis.Miniredis
var testRedisCache, testBadgerCache CacheInterface

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	testRedis = s
	defer s.Close()

	ops := Options{
		Server: testRedis.Host(),
		Port:   testRedis.Port(),
		Prefix: "test_cache",
	}
	testRedisCache, _ = New("redis", ops)
	defer testRedisCache.Close()

	testBadgerCache, _ = New("badger", Options{BadgerPath: "./testdata/badger"})
	defer testBadgerCache.Close()

	m.Run()
	cleanup()

	os.Exit(0)
}

func cleanup() {
	err := os.RemoveAll("./testdata/badger")
	if err != nil {
		log.Println("ERROR", err)
	}
}
