// package drwmutex provides a DRWMutex, a distributed RWMutex for use when
// there are many readers spread across many cores, and relatively few cores.
// DRWMutex is meant as an almost drop-in replacement for sync.RWMutex.
package drwmutex

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// cpus maps (non-consecutive) CPUID values to integer indices.
var cpus map[uint64]int

// init will construct the cpus map so that CPUIDs can be looked up to
// determine a particular core's lock index.
func init() {
	start := time.Now()
	cpus = map_cpus()
	fmt.Fprintf(os.Stderr, "%d/%d cpus found in %v: %v\n", len(cpus), runtime.NumCPU(), time.Now().Sub(start), cpus)
}

type DRWMutex []sync.RWMutex

// New returns a new, unlocked, distributed RWMutex.
func New() DRWMutex {
	return make(DRWMutex, runtime.GOMAXPROCS(0))
}

// Lock takes out an exclusive writer lock similar to sync.Mutex.Lock.
// A writer lock also excludes all readers.
func (mx DRWMutex) Lock() {
	for core := range mx {
		mx[core].Lock()
	}
}

// Unlock releases an exclusive writer lock similar to sync.Mutex.Unlock.
func (mx DRWMutex) Unlock() {
	for core := range mx {
		mx[core].Unlock()
	}
}

// RLocker returns a sync.Locker presenting Lock() and Unlock() methods that
// take and release a non-exclusive *reader* lock. Note that this call may be
// relatively slow, depending on the underlying system architechture, and so
// its result should be cached if possible.
func (mx DRWMutex) RLocker() sync.Locker {
	return mx[cpus[cpu()]].RLocker()
}

// RLock takes out a non-exclusive reader lock, and returns the lock that was
// taken so that it can later be released.
func (mx DRWMutex) RLock() (l sync.Locker) {
	l = mx[cpus[cpu()]].RLocker()
	l.Lock()
	return
}
