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
	"encoding/base64"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"github.com/splunk/protobuf-jsonnet/internal/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var _ = htmlTemplateFor("header", `
<html lang="en">
<head>
<link rel="stylesheet" href="{{.StylesPath}}/styles.css">
<title>{{.Title}}</title>
</head>
<body>
{{ if ne .StylesPath "doc" }}
<div class='crumb'>
	<a href="../../index.html">Home</a>
</div>
{{end}}
<h1>{{.Title}}</h1>
`)

var _ = htmlTemplateFor("footer", `
</body>
</html>
`)

func generateFile(t *template.Template, data interface{}) (string, error) {
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

func mustGenerateFile(t *template.Template, data interface{}) string {
	ret, err := generateFile(t, data)
	if err != nil {
		panic(err)
	}
	return ret
}

var enumDocTemplate = htmlTemplateFor("enum", `
{{template "header" (headerValues .Object.QualifiedName "..")}}

<h2>Values</h2>

<dl>
{{ range $k, $v := .Object.ValueMap }}
<dt>{{$k}}</dt><dd>{{$v}}</dd>
{{ end }}
</dl>

<h2>Example</h2>

<pre class='example'>
local types = import 'types.libsonnet';
types.{{.Object.QualifiedName}}.{{.Object.NameForFirstValue}}
</pre>

{{template "footer"}}
`)

type enumTemplateData struct {
	TypeLinkMap *typeLinkMap
	Object      model.Type
}

func (c *CodeGenerator) generateEnumDocs(e *model.Enum, typeLinks *typeLinkMap) *pluginpb.CodeGeneratorResponse_File {
	content := mustGenerateFile(enumDocTemplate, enumTemplateData{
		TypeLinkMap: typeLinks,
		Object:      e,
	})
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(docPath + "/" + filePathForType(e) + ".html"),
		Content: proto.String(content),
	}
}

func (c *CodeGenerator) fieldExample(fld *model.Field) string {
	switch {
	case fld.TypeName() == "bool",
		fld.TypeName() == "google.protobuf.BoolValue":
		return "false"
	case fld.TypeName() == "string",
		fld.TypeName() == "google.protobuf.StringValue":
		return `'string'`
	case fld.TypeName() == "bytes",
		fld.TypeName() == "google.protobuf.BytesValue":
		return "'" + base64.StdEncoding.EncodeToString([]byte("string")) + "'"
	case fld.TypeName() == "google.protobuf.DoubleValue",
		fld.TypeName() == "google.protobuf.Int32Value",
		fld.TypeName() == "google.protobuf.Int64Value",
		fld.TypeName() == "google.protobuf.UInt32Value",
		fld.TypeName() == "google.protobuf.UInt64Value":
		return "1"
	case fld.FieldType() == model.FieldTypeMessage:
		return fmt.Sprintf("_m_(types.%s)", fld.TypeName())
	case fld.FieldType() == model.FieldTypeEnum:
		t := c.TypeMap[fld.TypeName()]
		e := t.GetEnum()
		return fmt.Sprintf("_e_(types.%s.%s)", fld.TypeName(), e.NameForFirstValue())
	default:
		return "1"
	}
}

var linkRegex = regexp.MustCompile(`_([me])_\((.+?)\)`)

func (c *CodeGenerator) messageExampleHTML(m *model.Message, tlm *typeLinkMap) (string, error) {
	var example bytes.Buffer
	example.WriteString(fmt.Sprintf("local types = import '%s';\n\n", typesFile))
	example.WriteString(fmt.Sprintf("types.%s.", m.QualifiedName()))
	for _, field := range m.Fields() {
		example.WriteString("\n    ")
		example.WriteString(field.SetterName())
		example.WriteString("(")
		if field.IsList() {
			example.WriteString("[")
		} else if field.IsMap() {
			example.WriteString("{ 'key': ")
		}
		example.WriteString(c.fieldExample(field))
		if field.IsList() {
			example.WriteString("]")
		} else if field.IsMap() {
			example.WriteString(" }")
		}
		example.WriteString(").")
	}
	example.WriteString("\n_validate()")
	opts := formatter.DefaultOptions()
	opts.PadArrays = true
	exampleCode, err := formatJsonnet(example.String(), opts)
	if err != nil {
		return "", err
	}
	exampleCode = linkRegex.ReplaceAllStringFunc(exampleCode, func(matched string) string {
		parts := linkRegex.FindStringSubmatch(matched)
		what := parts[1]
		name := parts[2]
		qName := strings.TrimPrefix(name, "types.")
		if what == "e" {
			pos := strings.LastIndex(qName, ".")
			qName = qName[:pos]
		}
		link := tlm.Link(qName)
		if link == nil {
			return name
		} else {
			return fmt.Sprintf(`<a href="../%s.html">%s</a>`, link.Target, name)
		}
	})
	exampleCode = strings.ReplaceAll(exampleCode, "[", "<span class='coll'>[</span>")
	exampleCode = strings.ReplaceAll(exampleCode, "]", "<span class='coll'>]</span>")
	exampleCode = strings.ReplaceAll(exampleCode, "{", "<span class='coll'>{</span>")
	exampleCode = strings.ReplaceAll(exampleCode, "}", "<span class='coll'>}</span>")
	return exampleCode, nil
}

var messageDocTemplate = htmlTemplateFor("message", `
{{template "header" (headerValues .Object.QualifiedName "..")}}

{{$root := . }}

<h2>Example</h2>
<div class='disclaimer'>
Disclaimer: The example is meant to show what methods are available on the object and does not necessarily constitute working
code.
</div>

<pre class='example'>
{{.Example}}
</pre>

{{with .Object.NestedEnums}}
<h2>Nested Enums</h2>
<ul>
{{range .}}
	{{$message := .}}
	{{with $root.TypeLinkMap.Link .QualifiedName}}
		<li><a href="../{{.Target}}.html">{{$message.QualifiedName}}</a></li>
	{{else}}
		<li>{{.QualifiedName}}</li>
	{{end}}
{{end}}
</ul>
{{end}}

{{with .Object.NestedMessages}}
<h2>Nested Messages</h2>
<ul>
{{range .}}
	{{$message := .}}
	{{with $root.TypeLinkMap.Link .QualifiedName}}
		<li><a href="../{{.Target}}.html">{{$message.QualifiedName}}</a></li>
	{{else}}
		<li>{{.Name}}</li>
	{{end}}
{{end}}
</ul>
{{end}}

{{with .Object.Fields}}
<h2>Fields</h2>
<table class='fields'>
<thead>
	<tr>
		<th>Name</th>
		<th>Type</th>
		<th>One-of group</th>
		<th>Required</th>
		<th>Constraints</th>
	</tr>
</thead>
<tbody>
{{range .}}
	{{$field := .}}
	<tr>
		<td>{{.Name}}</td>
		<td>
			{{with .IsList}}[]{{end}}
			{{with .IsMap}}map[string]{{end}}
			{{with $root.TypeLinkMap.Link .TypeName}}
				<a href="../{{.Target}}.html">{{$field.TypeName}}</a>
			{{else}}
				{{$field.TypeName}}
			{{end}}
		</td>
		<td>{{.OneOfGroup}}</td>
		<td>
			{{with .IsRequired}}yes{{end}}&nbsp;
		</td>
		<td>
			<code>{{with .Constraints}}{{terseJson .}}{{end}}</code>
		</td>
	</tr>
{{end}}
</tbody>
</table>
{{end}}

{{template "footer"}}
`)

type messageTemplateData struct {
	enumTemplateData
	Example template.HTML
}

func (c *CodeGenerator) generateMessageDocs(m *model.Message, typeLinks *typeLinkMap) *pluginpb.CodeGeneratorResponse_File {
	example, err := c.messageExampleHTML(m, typeLinks)
	if err != nil {
		panic(err)
	}
	content := mustGenerateFile(messageDocTemplate, messageTemplateData{
		enumTemplateData: enumTemplateData{
			TypeLinkMap: typeLinks,
			Object:      m,
		},
		Example: template.HTML(example),
	})
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(docPath + "/" + filePathForType(m) + ".html"),
		Content: proto.String(content),
	}
}

var indexTemplate = htmlTemplateFor("index", `
{{template "header" (headerValues "Home" "doc")}}
<ul>
{{range $k, $v := .TypeLinkMap.Map}}
<li><a href="doc/{{$v}}.html">{{$k}}</a></li>
{{end}}
</ul>
{{template "footer"}}
`)

type indexTemplateData struct {
	TypeLinkMap *typeLinkMap
}

func (c *CodeGenerator) generateDocIndex(typeLinks *typeLinkMap) *pluginpb.CodeGeneratorResponse_File {
	content := mustGenerateFile(indexTemplate, indexTemplateData{
		TypeLinkMap: typeLinks,
	})
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(docIndexFile),
		Content: proto.String(content),
	}
}
