// Package driver contains all the logic to build a driver.
package driver

import "gopkg.in/bblfsh/sdk.v1/manifest"

var (
	// DriverBinary default location of the driver binary. Should not
	// override this variable unless you know what are you doing.
	DriverBinary = "/opt/driver/bin/driver"
	// NativeBinary default location of the native driver binary. Should not
	// override this variable unless you know what are you doing.
	NativeBinary = "/opt/driver/bin/native"
	// ManifestLocation location of the manifest file. Should not override
	// this variable unless you know what are you doing.
	ManifestLocation = "/opt/driver/etc/" + manifest.Filename
)
