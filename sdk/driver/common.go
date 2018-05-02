// Package driver contains all the logic to build a driver.
package driver

import (
	"gopkg.in/bblfsh/sdk.v2/manifest"
)

var (
	// DriverBinary default location of the driver binary. Should not
	// override this variable unless you know what are you doing.
	DriverBinary = "/opt/driver/bin/driver"
	// ManifestLocation location of the manifest file. Should not override
	// this variable unless you know what are you doing.
	ManifestLocation = "/opt/driver/etc/" + manifest.Filename
)

var DefaultDriver = NewExecDriver()

// Run is a common main function used as an entry point for drivers.
// It panics in case of an error.
func Run(t Transforms) {
	RunNative(DefaultDriver, t)
}

// RunNative is like Run but allows to provide a custom driver native driver implementation.
func RunNative(d BaseDriver, t Transforms) {
	dr, err := NewDriverFrom(d, t)
	if err != nil {
		panic(err)
	}
	s := NewServer(dr)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
