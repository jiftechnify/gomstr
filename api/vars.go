package main

import (
	"net/http"
	"sync"
)

var (
	varsLock sync.RWMutex
	vars     map[*http.Request]map[string]interface{}
)

// OpenVars creates a slot of the request scoped key-value pairs
func OpenVars(r *http.Request) {
	varsLock.Lock()
	defer varsLock.Unlock()

	if vars == nil {
		vars = map[*http.Request]map[string]interface{}{}
	}
	vars[r] = map[string]interface{}{}
}

// CloseVars deletes the request scoped variables
func CloseVars(r *http.Request) {
	varsLock.Lock()
	defer varsLock.Unlock()

	delete(vars, r)
}

// GetVar acquires the value related to key from the request scoped k-v pairs
func GetVar(r *http.Request, key string) interface{} {
	// concurrent read to the vars is allowed.
	varsLock.RLock()
	defer varsLock.RUnlock()

	return vars[r][key]
}

// SetVar associates the value to the key and puts it in the request scoped environment
func SetVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	defer varsLock.Unlock()

	vars[r][key] = value
}
