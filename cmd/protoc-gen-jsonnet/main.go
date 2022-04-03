/*
   Copyright 2022 Splunk Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"cd.splunkdev.com/kanantheswaran/protobuf-jsonnet/internal/codegen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func run() error {
	if len(os.Args) > 1 {
		return fmt.Errorf("unknown argument %q (this program should be run by protoc, not directly)", os.Args[1])
	}
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(in, req); err != nil {
		return err
	}
	cg := codegen.NewCodeGenerator(codegen.Options{})
	res, err := cg.Generate(req)
	if err != nil {
		return err
	}
	return writeResponse(res)
}

func writeResponse(res *pluginpb.CodeGeneratorResponse) error {
	out, err := proto.Marshal(res)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(out); err != nil {
		return err
	}
	return nil
}

func writeError(err error) error {
	res := &pluginpb.CodeGeneratorResponse{
		Error: proto.String(err.Error()),
	}
	return writeResponse(res)
}

func main() {
	if err := run(); err != nil {
		if err2 := writeError(err); err2 != nil {
			log.Fatalln(err)
		}
	}
}
