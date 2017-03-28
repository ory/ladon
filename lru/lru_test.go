package lru

import (
	"github.com/hashicorp/golang-lru"
	"regexp"
	"sync"
	"testing"
)

func BenchmarkLadonLRU(b *testing.B) {
	p, _ := regexp.Compile("^(foo|bar)$")
	c, _ := lru.New(1024)

	for z := 0; z < b.N; z++ {
		c.Add(z, p)
		o, _ := c.Get(z)
		if r, ok := o.(*regexp.Regexp); ok {
			r.MatchString("foo")
		}
	}
}

func BenchmarkLadonSimple(b *testing.B) {
	var lock sync.RWMutex

	p, _ := regexp.Compile("^(foo|bar)$")
	cache := map[int]interface{}{}

	for z := 0; z < b.N; z++ {
		lock.Lock()
		cache[z] = p
		lock.Unlock()
		lock.RLock()
		o := cache[z]
		lock.RUnlock()
		if r, ok := o.(*regexp.Regexp); ok {
			r.MatchString("foo")
		}
	}
}

func BenchmarkLadonNoTypeCast(b *testing.B) {
	var lock sync.RWMutex
	p, _ := regexp.Compile("^(foo|bar)$")
	cache := map[int]*regexp.Regexp{}

	for z := 0; z < b.N; z++ {
		lock.Lock()
		cache[z] = p
		lock.Unlock()
		lock.RLock()
		r := cache[z]
		lock.RUnlock()
		r.MatchString("foo")
	}
}

func BenchmarkLadonNoTypeCastNoLockJustYolo(b *testing.B) {
	p, _ := regexp.Compile("^(foo|bar)$")
	cache := map[int]*regexp.Regexp{}

	for z := 0; z < b.N; z++ {
		cache[z] = p
		r := cache[z]
		r.MatchString("foo")
	}
}
