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

package codegen

import (
	_ "embed"

	"cd.splunkdev.com/kanantheswaran/protobuf-jsonnet/internal/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	typesFile              = "types.libsonnet"
	docIndexFile           = "index.html"
	pkgPath                = "pkg"
	docPath                = "doc"
	validatorsFile         = pkgPath + "/validators.libsonnet"
	generatorJsonnetFile   = pkgPath + "/generator.libsonnet"
	constraintsJsonnetFile = pkgPath + "/field-constraints.libsonnet"
	dispatchJsonnetFile    = pkgPath + "/dispatch.libsonnet"
	wellKnownJsonnetFile   = pkgPath + "/well-known.libsonnet"
	stylesFile             = docPath + "/styles.css"
)

// Options are code generator Options.
type Options struct {
}

// CodeGenerator generates the jsonnet code for a set of messages and enums.
type CodeGenerator struct {
	Options
	TypeMap map[string]model.Type
	files   []*pluginpb.CodeGeneratorResponse_File
}

// NewCodeGenerator returns a code generator.
func NewCodeGenerator(opts Options) *CodeGenerator {
	return &CodeGenerator{
		Options: opts,
	}
}

func (c *CodeGenerator) Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	c.TypeMap = model.Load(&descriptorpb.FileDescriptorSet{File: req.GetProtoFile()})

	tlMap := c.TypeLinkMap()

	// generate stuff
	for _, v := range c.TypeMap {
		var f *pluginpb.CodeGeneratorResponse_File
		{
			switch {
			case v.GetEnum() != nil:
				f = c.generateEnum(v.GetEnum())
			case v.GetMessage() != nil:
				f = c.generateMessage(v.GetMessage())
			}
			c.files = append(c.files, f)
		}
		{
			switch {
			case v.GetEnum() != nil:
				f = c.generateEnumDocs(v.GetEnum(), tlMap)
			case v.GetMessage() != nil:
				f = c.generateMessageDocs(v.GetMessage(), tlMap)
			}
			if f != nil {
				c.files = append(c.files, f)
			}
		}
	}

	c.files = append(c.files, c.generateValidator())
	c.files = append(c.files, c.generateTypes())
	c.files = append(c.files, c.generateDocIndex(tlMap))

	c.files = append(c.files, c.staticFiles()...)
	return &pluginpb.CodeGeneratorResponse{
		File: c.files,
	}, nil
}

type link struct {
	Target string
}

type typeLinkMap struct {
	Map map[string]string
}

func (tlm *typeLinkMap) Link(name string) *link {
	if ret, ok := tlm.Map[name]; ok {
		return &link{Target: ret}
	}
	return nil
}

func (c *CodeGenerator) TypeLinkMap() *typeLinkMap {
	ret := map[string]string{}
	for _, v := range c.TypeMap {
		ret[v.QualifiedName()] = filePathForType(v)
	}
	return &typeLinkMap{Map: ret}
}

//go:embed static/well-known.libsonnet
var wellKnownJsonnet string

//go:embed static/dispatch.libsonnet
var dispatchJsonnet string

//go:embed static/styles.css
var stylesCSS string

//go:embed static/generator.libsonnet
var generatorJsonnet string

//go:embed static/field-constraints.libsonnet
var constraintsJsonnet string

func (c *CodeGenerator) staticFiles() []*pluginpb.CodeGeneratorResponse_File {
	return []*pluginpb.CodeGeneratorResponse_File{
		{
			Name:    proto.String(wellKnownJsonnetFile),
			Content: proto.String(wellKnownJsonnet),
		},
		{
			Name:    proto.String(dispatchJsonnetFile),
			Content: proto.String(dispatchJsonnet),
		},
		{
			Name:    proto.String(generatorJsonnetFile),
			Content: proto.String(generatorJsonnet),
		},
		{
			Name:    proto.String(constraintsJsonnetFile),
			Content: proto.String(constraintsJsonnet),
		},
		{
			Name:    proto.String(stylesFile),
			Content: proto.String(stylesCSS),
		},
	}
}
