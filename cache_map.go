package cache

import (
	"sync"
	"time"
)

type cacheMap struct {
	// m 存储容器
	m map[string]item

	// et 默认过期时间
	et time.Duration

	// nc 距离下一次清理周期
	nc int64

	// nt 下一次过期时间
	nt int64

	// cb 删除/清理后回调函数
	cb CleanCallback

	// map 读写锁
	l sync.RWMutex

	// nh 异步通知有新的清理时间
	nh chan bool
}

// New 创建新的缓存实例
// 该缓存底层采用原生map实现，读写锁保证数据安全，适用场景为读写频率相似或写多读少
// expire 该缓存实例内数据默认过期时间
func New(expire time.Duration) (cm *cacheMap) {
	cm = &cacheMap{
		m:  make(map[string]item),
		et: expire,
		cb: nil,
		nh: make(chan bool, 1),
	}
	go cm.clear()
	return
}

// Get 获取一个缓存值
// 第一个返回值为获取到的缓存数据
// 第二个返回值为是否从缓存中获取到值
// 当第二个返回值为 false 时，第一个返回值为 nil
func (cm *cacheMap) Get(key string) (interface{}, bool) {
	cm.l.RLock()
	vt, ok := cm.m[key]
	if !ok {
		cm.l.RUnlock()
		return nil, false
	}
	cm.l.RUnlock()
	return vt.v, true
}

// Set 设置数据，如果不存在写入，如果存在覆盖
func (cm *cacheMap) Set(key string, val interface{}, d time.Duration) {
	cm.set(key, val, d, true)
}

// Replace 替换数据
// 如果已存在相同key数据，则返回旧数据
// 如果是替换旧数据则返回true，否则返回false
func (cm *cacheMap) Replace(key string, val interface{}, d time.Duration) (interface{}, bool) {
	return cm.set(key, val, d, true)
}

// set 存储数据基础方法
func (cm *cacheMap) set(key string, val interface{}, d time.Duration, replace bool) (old interface{}, loaded bool) {
	cm.l.Lock()

	it := item{v: val}
	if d == NoExpiration {
		if replace {
			old, loaded = cm.m[key]
		}
		cm.m[key] = it
		cm.l.Unlock()
		return
	}

	nowTime := time.Now().UnixNano()
	var ex int64
	if d == DefaultExpiration {
		ex = int64(cm.et)
	} else {
		ex = int64(d)
	}
	it.t = nowTime + ex

	if replace {
		old, loaded = cm.m[key]
	}
	cm.m[key] = it
	if cm.nc == 0 || it.t < cm.nt {
		cm.nc = ex
		cm.nt = it.t
		cm.nh <- true
	}
	cm.l.Unlock()
	return
}

// Delete 删除数据
func (cm *cacheMap) Delete(k string) error {
	cm.l.Lock()
	delete(cm.m, k)
	cm.l.Unlock()
	return nil
}

// Clean 清空所有缓存
func (cm *cacheMap) Clean() error {
	cm.l.Lock()
	defer cm.l.Unlock()
	cm.m = make(map[string]item)
	return nil
}

// ClearCallback 设置缓存清理时回调函数
func (cm *cacheMap) ClearCallback(f CleanCallback) error {
	cm.cb = f
	return nil
}

//Range 遍历缓存中数据
func (cm *cacheMap) Range(f CleanCallback) {
	cm.l.RLock()
	for k, v := range cm.m {
		f(k, v.v)
	}
	cm.l.RUnlock()
}

//clear 周期清理缓存，启动入口
func (cm *cacheMap) clear() {
	for {
		if cm.nc == 0 {
			<-cm.nh
		}

		select {
		case <-time.Tick(time.Duration(cm.nc)):
			cm._clear()
		case <-cm.nh:

		}
	}
}

//_clear 清理缓存方法，删除已过期数据
func (cm *cacheMap) _clear() {
	var min int64
	cm.l.Lock()
	for key, val := range cm.m {
		if val.check() {
			delete(cm.m, key)
		} else {
			if min == 0 || val.t < min {
				min = val.t
			}
		}
	}
	if min > 0 {
		cm.nc = min - time.Now().UnixNano()
	} else {
		cm.nc = 0
	}
	cm.l.Unlock()
}
