package duck

import (
	"math/rand/v2"
	"sync"
	"time"
)

// Mutex is the mutex with synchronized map, it allows reducing unnecessary locks among different keys.
// This implementation comes from the mapmutex package, I simply copied it here instead of adding it as a dependency.
// See the code here: https://github.com/EagleChen/mapmutex/blob/master/mutex.go
type Mutex struct {
	locks     map[interface{}]interface{}
	m         *sync.Mutex
	maxRetry  int
	maxDelay  float64 // in nanosend
	baseDelay float64 // in nanosecond
	factor    float64
	jitter    float64
}

// TryLock tries to acquire the lock.
func (m *Mutex) TryLock(key interface{}) bool {
	for i := range m.maxRetry {
		m.m.Lock()
		if _, ok := m.locks[key]; ok { // if locked
			m.m.Unlock()
			time.Sleep(m.backoff(i))
		} else { // if unlock, lockit
			m.locks[key] = struct{}{}
			m.m.Unlock()
			return true
		}
	}

	return false
}

// Unlock unlocks for the key
// please call Unlock only after having acquired the lock.
func (m *Mutex) Unlock(key interface{}) {
	m.m.Lock()
	delete(m.locks, key)
	m.m.Unlock()
}

func (m *Mutex) backoff(retries int) time.Duration {
	if retries == 0 {
		return time.Duration(m.baseDelay) * time.Nanosecond
	}
	backoff, maxDelay := m.baseDelay, m.maxDelay
	for backoff < maxDelay && retries > 0 {
		backoff *= m.factor
		retries--
	}
	if backoff > maxDelay {
		backoff = maxDelay
	}
	backoff *= 1 + m.jitter*(rand.Float64()*2-1) //nolint:gosec
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff) * time.Nanosecond
}

// NewMapMutex returns a mapmutex with default configs.
func NewMapMutex() *Mutex {
	return &Mutex{
		locks:     make(map[interface{}]interface{}),
		m:         &sync.Mutex{},
		maxRetry:  200,
		maxDelay:  100000000, // 0.1 second
		baseDelay: 10,        // 10 nanosecond
		factor:    1.1,
		jitter:    0.2,
	}
}

// NewCustomizedMapMutex returns a customized mapmutex.
func NewCustomizedMapMutex(mRetry int, mDelay, bDelay, factor, jitter float64) *Mutex {
	return &Mutex{
		locks:     make(map[interface{}]interface{}),
		m:         &sync.Mutex{},
		maxRetry:  mRetry,
		maxDelay:  mDelay,
		baseDelay: bDelay,
		factor:    factor,
		jitter:    jitter,
	}
}

var databaseLocks = NewMapMutex()

func LockDatabase(path string) {
	for !databaseLocks.TryLock(path) {
	}
}

func UnlockDatabase(path string) {
	databaseLocks.Unlock(path)
}
