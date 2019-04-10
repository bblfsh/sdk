package impl

import (
	"github.com/bblfsh/sdk/v3/driver/native"
	"github.com/bblfsh/sdk/v3/driver/server"
)

func init() {
	// Can be overridden to link a native driver into a Go driver server.
	server.DefaultDriver = native.NewDriver(native.UTF8)
}
