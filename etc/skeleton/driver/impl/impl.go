package impl

import "gopkg.in/bblfsh/sdk.v2/sdk/driver"

func init() {
	// Can be overridden to link a native driver into a Go driver server.
	driver.DefaultDriver = driver.NewExecDriver()
}
