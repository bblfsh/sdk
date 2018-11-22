package cmd

// Contains CLI helpers, shared between bblfshd and drivers.

import (
	"flag"
	"fmt"

	"google.golang.org/grpc"
)

const (
	// MaxMsgSizeCLIName is name of the CLI flag to set max msg size.
	maxMsgSizeCLIName = "grpc-max-message-size"
	// MaxMsgSizeCLIDesc is description for the CLI flag to set max msg size.
	maxMsgSizeCLIDesc = "max. message size to send/receive to/from clients (in MB)"

	// DefaulGRPCMaxSendRecvMsgSizeMB is maximum msg size for gRPC in MB.
	DefaulGRPCMaxSendRecvMsgSizeMB = 100
	maxMsgSizeCapMB                = 2048
)

// GRPCSizeOptions returns a slice of gRPC server options with the max
// message size the server can send/receive set.
// If a >2GB value is requested: the maximum size limit is capped
// at 100 MB and an error is returned.
// It is intended to be shared by gRPC in bblfshd Server and Drivers.
func GRPCSizeOptions(maxMessageSizeMB int) ([]grpc.ServerOption, error) {
	var err error
	sizeMB := maxMessageSizeMB
	if sizeMB >= maxMsgSizeCapMB || sizeMB <= 0 {
		err = fmt.Errorf("%s=%d is too big (limit is %dMB), using %d instead",
			maxMsgSizeCLIName, sizeMB, maxMsgSizeCapMB-1, DefaulGRPCMaxSendRecvMsgSizeMB)
		sizeMB = DefaulGRPCMaxSendRecvMsgSizeMB
	}

	sizeBytes := sizeMB * 1024 * 1024
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(sizeBytes),
		grpc.MaxSendMsgSize(sizeBytes),
	}, err
}

// MaxSendRecvMsgSizeMB sets the CLI configuation flag for max
// gRPC send/recive msg size.
func MaxSendRecvMsgSizeMB(fs *flag.FlagSet) *int {
	return fs.Int(maxMsgSizeCLIName, DefaulGRPCMaxSendRecvMsgSizeMB, maxMsgSizeCLIDesc)
}
