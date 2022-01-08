package cache

import (
	cac "github.com/patrickmn/go-cache"
	"strconv"
	"testing"
	"time"
)

var (
	cMap         *cacheMap
	cSyncMap     *cacheSyncMap
	sessionCache *cac.Cache
)

func TestMain(m *testing.M) {
	cMap = New(time.Minute)
	cSyncMap = NewS(time.Minute)
	sessionCache = cac.New(10*time.Minute, 20*time.Minute)
	m.Run()
}

//
//func TestOpenCache(t *testing.T) {
//	cSyncMap.Set(1, 2, 0)
//	go func() {
//		time.Sleep(time.Second)
//		cSyncMap.Set(2, 2, 0)
//	}()
//	go func() {
//		time.Sleep(time.Second * 4)
//		t.Log(cSyncMap.Get(1))
//		t.Log(cSyncMap.Get(2))
//	}()
//	go func() {
//		time.Sleep(time.Second * 5)
//		t.Log(cSyncMap.Get(1))
//		t.Log(cSyncMap.Get(2))
//	}()
//	time.Sleep(time.Second * 10)
//	t.Log(cSyncMap.Get(1))
//	t.Log(cSyncMap.Get(2))
//}

//func TestOpenBenchCache(t *testing.T) {
//	for i := 0; i < 1000000; i++ {
//		cSyncMap.Set(i, i, 0)
//	}
//	t.Log("start sleep")
//	time.Sleep(time.Second * 10)
//	t.Log(cSyncMap.Get(1))
//}

//
//func TestCacheMap(t *testing.T) {
//	cMap.Set("123", "test cache")
//	vi, ok := cMap.Get("123")
//	if !ok {
//		t.Error("map cache get is not exists")
//		t.FailNow()
//	}
//	v, ok := vi.(string)
//	if !ok {
//		t.Error("type of data abnormal")
//		t.FailNow()
//	}
//	if v != "test cache" {
//		t.Error("value is not equal")
//		t.FailNow()
//	}
//}
//
//func TestSyncCacheMap(t *testing.T) {
//	cSyncMap.Set("123", "test cache")
//	vi, ok := cSyncMap.Get("123")
//	if !ok {
//		t.Error("sync map cache get is not exists")
//		t.FailNow()
//	}
//	v, ok := vi.(string)
//	if !ok {
//		t.Error("type of data abnormal")
//		t.FailNow()
//	}
//	if v != "test cache" {
//		t.Error("value is not equal")
//		t.FailNow()
//	}
//}

func BenchmarkCacheMapSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cMap.Set(strconv.Itoa(i), i, DefaultExpiration)
	}
}

func BenchmarkCacheMapGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cMap.Get(strconv.Itoa(i))
	}
}

func BenchmarkCacheSyncMapSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cSyncMap.Set(i, i, DefaultExpiration)
	}
}

func BenchmarkCacheSyncMapGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//if i%20 == 0 {
		//	cSyncMap.Set(i, i, 0)
		//} else {
		cSyncMap.Get(i)
		//}
	}
}

func BenchmarkOpenCacheSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sessionCache.Set(strconv.Itoa(i), i, DefaultExpiration)
	}
}

func BenchmarkOpenCacheGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//if i%20 == 0 {
		//	//cSyncMap.Set(i, i, 0)
		//	sessionCache.Set(strconv.Itoa(i), i, 0)
		//} else {
		sessionCache.Get(strconv.Itoa(i))
		//}

	}
}

//
//func BenchmarkSyncMapGet(b *testing.B) {
//	sm := sync.Map{}
//	for i := 0; i < b.N; i++ {
//		sm.Load(i)
//	}
//}
