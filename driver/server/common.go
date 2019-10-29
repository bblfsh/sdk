package server

import (
	"runtime"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/driver/native"
)

// ManifestLocation location of the manifest file. Should not override
// this variable unless you know what are you doing.
var ManifestLocation = driver.ManifestLocation

// Run is a common main function used as an entry point for drivers.
// It panics in case of an error.
func Run(t driver.Transforms) {
	n := runtime.NumCPU()
	ch := make(chan driver.Native, n)
	for i := 0; i < n; i++ {
		ch <- native.NewDriver("")
	}
	RunNative(ch, t)
}

// RunNative is like Run but allows to provide a custom driver native driver implementation.
func RunNative(ch chan driver.Native, t driver.Transforms) {
	m, err := manifest.Load(ManifestLocation)
	if err != nil {
		panic(err)
	}
	dr, err := driver.NewDriverFrom(ch, m, t)
	if err != nil {
		panic(err)
	}
	s := NewServer(dr)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
