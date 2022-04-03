//go:build !release
// +build !release

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

package testutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type ProtocConfig struct {
	Files        []string
	IncludePaths []string
	Parameter    string
}

type CodeGenerator interface {
	Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
}

// Request returns a file descriptor set for the supplied files using the protoc command.
func Request(t *testing.T, cfg ProtocConfig) *pluginpb.CodeGeneratorRequest {
	executable, err := exec.LookPath("protoc")
	require.NoError(t, err)

	tmpFile, err := ioutil.TempFile("", "proto-test")
	require.NoError(t, err)
	tmpName := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpName) }()

	var args []string
	for _, p := range cfg.IncludePaths {
		abs, err := filepath.Abs(p)
		require.NoError(t, err)
		args = append(args, "-I", abs)
	}
	args = append(args, fmt.Sprintf("--descriptor_set_out=%s", tmpName))
	args = append(args, cfg.Files...)
	t.Log(executable, args)
	cmd := exec.Command(executable, args...)
	var buf bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v\nstderr: %s)", err, buf.String())
	}
	require.NoError(t, err)

	b, err := ioutil.ReadFile(tmpName)
	require.NoError(t, err)
	var desc descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(b, &desc)
	require.NoError(t, err)
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: cfg.Files,
		Parameter:      proto.String(cfg.Parameter),
		ProtoFile:      desc.GetFile(),
		CompilerVersion: &pluginpb.Version{
			Major: proto.Int32(3),
			Minor: proto.Int32(17),
			Patch: proto.Int32(3),
		},
	}
}

// GenerateCode generates code in a subdirectory of the test dir and returns the directory name.
func GenerateCode(t *testing.T, generator CodeGenerator, req *pluginpb.CodeGeneratorRequest, testDir string) string {
	targetDir := filepath.Join(testDir, ".gen")
	err := os.RemoveAll(targetDir)
	require.NoError(t, err)
	err = os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)
	res, err := generator.Generate(req)
	require.NoError(t, err)
	for _, outFile := range res.GetFile() {
		name := outFile.GetName()
		dir := filepath.Dir(name)
		err = os.MkdirAll(filepath.Join(targetDir, dir), 0755)
		require.NoError(t, err)
		err = ioutil.WriteFile(filepath.Join(targetDir, name), []byte(outFile.GetContent()), 0644)
		require.NoError(t, err)
	}
	return targetDir
}
