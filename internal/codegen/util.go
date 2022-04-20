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
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"text/template"

	"github.com/google/go-jsonnet/formatter"
	"github.com/iancoleman/strcase"
	"github.com/splunk/protobuf-jsonnet/internal/model"
)

const (
	defaultPackage = "_default"
)

// toJSON implements a json encoding function for use in text templates.
func toJSON(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// toTerseJSON implements a terse json encoding function for use in text templates.
func toTerseJSON(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func templateFor(str string) *template.Template {
	return template.Must(
		template.New("jsonnet").
			Funcs(template.FuncMap{
				"json":            toJSON,
				"terseJson":       toTerseJSON,
				"fileNameForType": fileNameForType,
				"filePathForType": filePathForType,
			}).
			Parse(str),
	)
}

var root = htmlTemplate.New("__root__")

func htmlTemplateFor(name string, str string) *htmlTemplate.Template {
	return htmlTemplate.Must(
		root.New(name).
			Funcs(htmlTemplate.FuncMap{
				"json":            toJSON,
				"terseJson":       toTerseJSON,
				"fileNameForType": fileNameForType,
				"filePathForType": filePathForType,
				"headerValues": func(title, stylesPath string) map[string]interface{} {
					return map[string]interface{}{
						"Title":      title,
						"StylesPath": stylesPath,
					}
				},
			}).
			Parse(str),
	)
}

func formatJsonnet(s string, opts formatter.Options) (string, error) {
	content, err := formatter.Format("generated.jsonnet", s, opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, s)
		return "", err
	}
	return content, nil
}

// generateJsonnet generates code using the supplied template and data, ensures it is valid jsonnet
// by running the jsonnet formatter on it.
func generateJsonnet(t *template.Template, data interface{}) (string, error) {
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		return "", err
	}
	return formatJsonnet(b.String(), formatter.DefaultOptions())
}

func mustGenerateJsonnet(t *template.Template, data interface{}) string {
	ret, err := generateJsonnet(t, data)
	if err != nil {
		panic(err)
	}
	return ret
}

func fileNameForType(t model.Type) string {
	return strcase.ToKebab(t.NestedName())
}

func filePathForType(t model.Type) string {
	p := t.Package()
	if p == "" {
		p = defaultPackage
	}
	return fmt.Sprintf("%s/%s", p, fileNameForType(t))
}
