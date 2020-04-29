// Copyright 2020 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reflect

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cilium/hubble/pkg/defaults"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

var (
	serverURL string
	output    string
)

// New reflect command.
func New() *cobra.Command {
	reflectCmd := &cobra.Command{
		Use:   "reflect",
		Short: "Let Hubble reflect upon itself",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReflect()
		},
	}
	reflectCmd.Flags().StringVarP(&serverURL, "server", "", defaults.GetDefaultSocketPath(), "URL to connect to server")
	reflectCmd.Flags().StringVarP(&output, "output", "o", "compact", "\"compact\" or \"json\"")
	return reflectCmd
}

func runReflect() (err error) {
	conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), defaults.DefaultRequestTimeout)
	defer cancel()

	client, err := rpb.NewServerReflectionClient(conn).ServerReflectionInfo(ctx)
	if err != nil {
		return err
	}
	req := rpb.ServerReflectionRequest{
		Host:           serverURL,
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{},
	}
	if err := client.Send(&req); err != nil {
		return err
	}
	res, err := client.Recv()
	if err != nil {
		return err
	}
	services, ok := res.GetMessageResponse().(*rpb.ServerReflectionResponse_ListServicesResponse)
	if !ok {
		return fmt.Errorf("unexpected response: %v", res)
	}
	visited := make(map[string]bool)
	for _, svc := range services.ListServicesResponse.Service {
		if svc.Name == "grpc.reflection.v1alpha.ServerReflection" || svc.Name == "grpc.health.v1.Health" {
			continue
		}
		err = client.Send(&rpb.ServerReflectionRequest{
			Host: serverURL,
			MessageRequest: &rpb.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: svc.Name,
			},
		})
		res, err := client.Recv()
		if err != nil {
			return err
		}
		files, ok := res.GetMessageResponse().(*rpb.ServerReflectionResponse_FileDescriptorResponse)
		if !ok {
			return fmt.Errorf("unexpected response: %v", res)
		}
		if err := handleDescriptorResponse(visited, client, files); err != nil {
			return err
		}

	}
	return nil
}

func handleDescriptorResponse(
	visited map[string]bool,
	client rpb.ServerReflection_ServerReflectionInfoClient,
	resp *rpb.ServerReflectionResponse_FileDescriptorResponse) error {
	for _, r := range resp.FileDescriptorResponse.FileDescriptorProto {
		desc := descriptor.FileDescriptorProto{}
		if err := proto.Unmarshal(r, &desc); err != nil {
			return err
		}
		if !visited[desc.GetName()] {
			visited[desc.GetName()] = true
			if err := ppDesc(&desc); err != nil {
				return err
			}
		}
		for _, dep := range desc.Dependency {
			if err := resolveFile(visited, client, dep); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveFile(
	visited map[string]bool,
	client rpb.ServerReflection_ServerReflectionInfoClient, filename string) error {
	req := rpb.ServerReflectionRequest{
		Host:           serverURL,
		MessageRequest: &rpb.ServerReflectionRequest_FileByFilename{FileByFilename: filename},
	}
	if err := client.Send(&req); err != nil {
		return err
	}
	res, err := client.Recv()
	if err != nil {
		return err
	}
	files, ok := res.GetMessageResponse().(*rpb.ServerReflectionResponse_FileDescriptorResponse)
	if !ok {
		return fmt.Errorf("unexpected response: %v", res.GetMessageResponse())
	}
	return handleDescriptorResponse(visited, client, files)
}

func getParamType(typeName string, streaming bool) string {
	if streaming {
		return fmt.Sprintf("stream %s", typeName)
	} else {
		return typeName
	}
}

func ppDesc(desc *descriptor.FileDescriptorProto) error {
	if output == "json" {
		return json.NewEncoder(os.Stdout).Encode(desc)
	}
	for _, s := range desc.GetService() {
		for _, val := range s.GetMethod() {
			fmt.Printf("service rpc .%s.%s(%s) returns (%s) {}\n", desc.GetPackage(), val.GetName(), getParamType(val.GetInputType(), val.GetClientStreaming()), getParamType(val.GetOutputType(), val.GetServerStreaming()))
		}
	}
	for _, e := range desc.GetEnumType() {
		for _, val := range e.GetValue() {
			fmt.Printf("enum .%s.%s %s = %d;\n", desc.GetPackage(), e.GetName(), val.GetName(), val.GetNumber())
		}
	}
	for _, f := range desc.GetMessageType() {
		for _, field := range f.GetField() {
			if field.GetTypeName() != "" {
				if field.OneofIndex != nil {
					oneOf := f.OneofDecl[field.GetOneofIndex()]
					fmt.Printf("message .%s.%s oneOf %s %s %s = %d;\n", desc.GetPackage(), f.GetName(), oneOf.GetName(), field.GetTypeName(), field.GetName(), field.GetNumber())
				} else {
					fmt.Printf("message .%s.%s %s %s = %d;\n", desc.GetPackage(), f.GetName(), field.GetTypeName(), field.GetName(), field.GetNumber())
				}
			} else {
				typeString := strings.ToLower(strings.Split(field.GetType().String(), "_")[1])
				fmt.Printf("message .%s.%s %s %s = %d;\n", desc.GetPackage(), f.GetName(), typeString, field.GetName(), field.GetNumber())
			}
		}
	}
	return nil
}
