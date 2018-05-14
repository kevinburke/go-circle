/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package onceflight provides a duplicate function call suppression
// mechanism.
package onceflight

import "sync"

// call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
	// true if call has been completed; guarded by (Once)Group.mu
	complete bool
}

// Group is like singleflight.Group, but caches the results of calls. Only one
// call per key will ever execute; the results of the first call for a given key
// will be cached and returned to every caller from then on.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		if c.complete {
			g.mu.Unlock()
			return c.val, c.err
		}
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	g.mu.Lock()
	c.complete = true
	g.mu.Unlock()
	c.wg.Done()

	return c.val, c.err
}
