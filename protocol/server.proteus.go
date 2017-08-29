// Copyright 2017 Sourced Technologies SL
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package protocol

import (
	"golang.org/x/net/context"
)

type protocolServiceServer struct {
}

func NewProtocolServiceServer() *protocolServiceServer {
	return &protocolServiceServer{}
}
func (s *protocolServiceServer) Parse(ctx context.Context, in *ParseRequest) (result *ParseResponse, err error) {
	result = new(ParseResponse)
	result = Parse(in)
	return
}
func (s *protocolServiceServer) Version(ctx context.Context, in *VersionRequest) (result *VersionResponse, err error) {
	result = new(VersionResponse)
	result = Version(in)
	return
}
