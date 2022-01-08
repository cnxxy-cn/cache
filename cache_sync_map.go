package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// cacheSyncMap 缓存结构体
// 使用sync.Map作为基础，去掉锁争用
// 适合读多写少的场景，如果绝大部分是写入，请使用 cacheMap
type cacheSyncMap struct {
	// m 存储容器
	m sync.Map

	// et 默认过期时间
	et time.Duration

	// cb 删除/清理后回调函数
	cb CleanCallback

	// nc 距离下一次清理周期
	nc int64

	// nt 下一次过期时间
	nt int64

	// nh 异步通知有新的清理时间
	nh chan bool
}

// NewS 创建新的缓存实例
// 该缓存底层采用sync.map实现，不使用锁，适用场景为读多写少
// expire 该缓存实例内数据默认过期时间
func NewS(expire time.Duration) (csm *cacheSyncMap) {
	csm = &cacheSyncMap{
		et: expire,
		cb: nil,
		nh: make(chan bool, 1),
	}
	go csm.clear()
	return
}

// Get 获取一个缓存值
// 第一个返回值为获取到的缓存数据
// 第二个返回值为是否从缓存中获取到值
// 当第二个返回值为 false 时，第一个返回值为 nil
func (csm *cacheSyncMap) Get(k interface{}) (interface{}, bool) {
	val, ok := csm.m.Load(k)
	if ok {
		return val.(item).v, true
	}
	return nil, false
}

// Add 添加数据，只增加不覆盖
// 如果已存在则返回 false， 如果不存在则存储数据并返回 true
func (csm *cacheSyncMap) Add(key, val interface{}, d time.Duration) bool {
	_, ok := csm.m.Load(key)
	if ok {
		return false
	}
	csm.set(key, val, d, false)
	return true
}

// Set 设置数据，如果不存在写入，如果存在覆盖
func (csm *cacheSyncMap) Set(key, val interface{}, d time.Duration) {
	csm.set(key, val, d, false)
}

// Replace 替换数据
// 返回 old 为旧数据，如果不存在则为nil
// loaded 是否找到被替换数据
func (csm *cacheSyncMap) Replace(key, val interface{}, d time.Duration) (old interface{}, loaded bool) {
	return csm.set(key, val, d, true)
}

// set 存储数据基础方法
func (csm *cacheSyncMap) set(key, val interface{}, d time.Duration, replace bool) (old interface{}, loaded bool) {
	it := item{v: val}
	if d == NoExpiration {
		if replace {
			old, loaded = csm.m.LoadOrStore(key, it)
		} else {
			csm.m.Store(key, it)
		}
		return
	}

	nowTime := time.Now().UnixNano()
	var ex int64
	if d == DefaultExpiration {
		ex = int64(csm.et)
	} else {
		ex = int64(d)
	}
	it.t = nowTime + ex

	if replace {
		old, loaded = csm.m.LoadOrStore(key, it)
	} else {
		csm.m.Store(key, it)
	}
	if atomic.LoadInt64(&csm.nc) == 0 || it.t < atomic.LoadInt64(&csm.nt) {
		atomic.StoreInt64(&csm.nc, ex)
		atomic.StoreInt64(&csm.nt, it.t)
		csm.nh <- true
	}
	return
}

// Delete 删除数据
func (csm *cacheSyncMap) Delete(k interface{}) error {
	csm.m.Delete(k)
	return nil
}

// Clean 清空所有数据
func (csm *cacheSyncMap) Clean() error {
	csm.m = sync.Map{}
	atomic.StoreInt64(&csm.nc, 0)
	atomic.StoreInt64(&csm.nt, 0)
	csm.nh <- true
	return nil
}

// ClearCallback 设置删除时回调函数
func (csm *cacheSyncMap) ClearCallback(f CleanCallback) error {
	csm.cb = f
	return nil
}

//Range 遍历缓存中数据
func (csm *cacheSyncMap) Range(f CleanCallback) {
	csm.m.Range(func(k, v interface{}) bool {
		f(k, v.(item))
		return true
	})
}

//clear 周期清理缓存，启动入口
func (csm *cacheSyncMap) clear() {
	for {
		//log.Println("tick csm.nc", csm.nc)
		select {
		case <-time.Tick(time.Duration(csm.nc)):
			csm._clear()
		case <-csm.nh:
			//log.Println("get nh")
		}
	}
}

//_clear 清理缓存方法，删除已过期数据
func (csm *cacheSyncMap) _clear() {
	var min int64
	csm.m.Range(func(k, v interface{}) bool {
		it := v.(item)
		if it.check() {
			csm.m.Delete(k)
		} else {
			if min == 0 || it.t < min {
				min = it.t
			}
		}
		return true
	})
	if min > 0 {
		atomic.StoreInt64(&csm.nc, min-time.Now().UnixNano())
	} else {
		atomic.StoreInt64(&csm.nc, 0)
	}
}
