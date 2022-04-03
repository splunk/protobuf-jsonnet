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
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func ensurePackage(root map[string]interface{}, elems []string) {
	if len(elems) == 0 {
		return
	}
	elem := elems[0]
	var rest []string
	rest = append(rest, elems[1:]...)
	var val map[string]interface{}
	v, ok := root[elem]
	if !ok {
		val = map[string]interface{}{}
		root[elem] = val
	} else {
		val = v.(map[string]interface{})
	}
	ensurePackage(val, rest)
}

func findPackage(root map[string]interface{}, name string) map[string]interface{} {
	if name != "" {
		elems := strings.Split(name, ".")
		for _, elem := range elems {
			root = root[elem].(map[string]interface{})
		}
	}
	return root
}

func render(root map[string]interface{}) string {
	var b bytes.Buffer
	b.WriteString("{\n")
	var keys []string
	for k := range root {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := root[k]
		b.WriteString(k)
		b.WriteString(": ")
		if sub, ok := v.(map[string]interface{}); ok {
			b.WriteString(render(sub))
		} else {
			b.WriteString(v.(string))
		}
		b.WriteString(",\n")
	}
	b.WriteString("}")
	return b.String()
}

func (c *CodeGenerator) makePackageMap() map[string]interface{} {
	root := map[string]interface{}{}
	pkgs := map[string]bool{}
	for _, v := range c.TypeMap {
		pkgs[v.Package()] = true
	}

	for pkg := range pkgs {
		if pkg != "" {
			ensurePackage(root, strings.Split(pkg, "."))
		}
	}
	return root
}

func (c *CodeGenerator) generateTypes() *pluginpb.CodeGeneratorResponse_File {
	root := c.makePackageMap()
	for _, v := range c.TypeMap {
		if !v.IsTopLevel() {
			continue
		}
		entry := findPackage(root, v.Package())
		entry[v.Name()] = fmt.Sprintf("(import 'pkg/%s.libsonnet').definition", filePathForType(v))
	}
	out := render(root)
	content, err := formatJsonnet(out, formatter.DefaultOptions())
	if err != nil {
		panic(err)
	}
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(typesFile),
		Content: proto.String(content),
	}

}
