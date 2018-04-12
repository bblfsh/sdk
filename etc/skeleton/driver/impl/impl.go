package impl

import "gopkg.in/bblfsh/sdk.v1/sdk/driver"

func init() {
	// Can be overridden to link a native driver into a Go driver server.
	driver.DefaultDriver = driver.NewExecDriver()
}
