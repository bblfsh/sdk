package sdk

import (
	"fmt"

	"google.golang.org/grpc"
)

const NativeBin = "/opt/driver/bin/native"
const NativeBinTest = "/opt/driver/src/build/native"

const (
	// MaxMsgSizeCLIName is name of the CLI flag to set max msg size.
	MaxMsgSizeCLIName = "grpc-max-message-size"
	// MaxMsgSizeCLIDesc is description for the CLI flag to set max msg size.
	MaxMsgSizeCLIDesc = "max. message size to send/receive to/from clients (in MB)"

	// DefaulGRPCMaxMsgSizeMb is maximum msg size for gRPC in Mb.
	DefaulGRPCMaxMsgSizeMb = 100
	gRPCMaxMsgSizeCapMb    = 2048
)

//BuildGRPCOptions creats gRPC ServerOption \w maxRecv/Send msg size set.
// Is intended to be shared by gRPC in bblfshd Server and Drivers.
// Sets the hard limit of message size to less than 2GB since
// it may overflow an int value, and it should be big enough.
func BuildGRPCOptions(maxMessageSizeMb int) ([]grpc.ServerOption, error) {
	var err error
	size := maxMessageSizeMb
	if size >= gRPCMaxMsgSizeCapMb {
		err = fmt.Errorf("%s=%d is too big (limit is %dMB), using %d instead",
			MaxMsgSizeCLIName, size, gRPCMaxMsgSizeCapMb-1, DefaulGRPCMaxMsgSizeMb)
		size = DefaulGRPCMaxMsgSizeMb
	}

	size = size * 1024 * 1024

	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(size),
		grpc.MaxSendMsgSize(size),
	}, err
}
