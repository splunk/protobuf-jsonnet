package codegen_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/formatter"
	"github.com/splunk/protobuf-jsonnet/internal/codegen"
	"github.com/splunk/protobuf-jsonnet/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Test struct {
	Name    string      `json:"name"`
	Summary string      `json:"summary,omitempty"`
	File    string      `json:"file,omitempty"`
	Code    string      `json:"code,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Err     string      `json:"err,omitempty"`
}

type Suite struct {
	VM              string   `json:"vm"`
	IncludeValidate bool     `json:"includeValidate,omitempty"`
	ProtoFiles      []string `json:"protoFiles,omitempty"`
}

type testRunner struct {
	t    *testing.T
	test Test
	dir  string
	vm   func(code, logicalFile string) (string, error)
}

func mustFormatJsonnet(t *testing.T, s string) string {
	content, err := formatter.Format("generated.jsonnet", s, formatter.DefaultOptions())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, s)
	}
	require.NoError(t, err)
	return content
}

func (r *testRunner) run() {
	code := r.test.Code
	if code == "" {
		b, err := os.ReadFile(filepath.Join(r.dir, r.test.File))
		require.NoError(r.t, err)
		code = string(b)
	}
	code = mustFormatJsonnet(r.t, code)
	if os.Getenv("VERBOSE") == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "CODE:\n%s\n", code)
	}
	res, err := r.vm(code, "<"+r.test.Name+">")
	if os.Getenv("VERBOSE") == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "RES:\n%s\n", res)
		_, _ = fmt.Fprintf(os.Stderr, "ERR:%v\n", err)
	}
	if r.test.Err != "" {
		require.Error(r.t, err)
		assert.Contains(r.t, err.Error(), r.test.Err)
		return
	}
	require.NoError(r.t, err)
	var actual interface{}
	err = json.Unmarshal([]byte(res), &actual)
	require.NoError(r.t, err)
	assert.EqualValues(r.t, r.test.Result, actual)
}

type suiteRunner struct {
	t      *testing.T
	dir    string
	config Suite
	genDir string
}

func (s *suiteRunner) run() {

	var includePaths []string
	if s.config.ProtoFiles == nil {
		files, err := filepath.Glob(fmt.Sprintf("%s/*.proto", s.dir))
		require.NoError(s.t, err)
		var protoFiles []string
		for _, file := range files {
			// Include the proto file's parent path so that protoc files can be looked up
			// Verbatim from protoc's error message:
			// Note that the proto_path must be an exact prefix of the .proto file names --
			// protoc is too dumb to figure out when two paths (e.g. absolute and relative)
			// are equivalent (it's harder than you think).
			// The lookup is problematic on windows but not on *nix systems
			includePaths = append(includePaths, filepath.Dir(file))
			protoFiles = append(protoFiles, filepath.Base(file))
		}
		s.config.ProtoFiles = protoFiles
	}

	if s.config.IncludeValidate {
		includePaths = append(includePaths, ".", "..")
	}
	req := testutil.Request(s.t, testutil.ProtocConfig{
		Files:        s.config.ProtoFiles,
		IncludePaths: includePaths,
		Parameter:    "",
	})
	cg := codegen.NewCodeGenerator(codegen.Options{})
	generatedDir := testutil.GenerateCode(s.t, cg, req, s.dir)
	s.genDir = generatedDir
	file := filepath.Join(s.dir, "tests.jsonnet")
	b, err := os.ReadFile(file)
	require.NoError(s.t, err)
	evaluated, err := s.vm()(string(b), file)
	require.NoError(s.t, err)
	var tests []Test
	err = json.Unmarshal([]byte(evaluated), &tests)
	require.NoError(s.t, err)
	for _, test := range tests {
		s.t.Run(test.Name, func(t *testing.T) {
			runner := &testRunner{
				t:    t,
				test: test,
				dir:  s.dir,
				vm:   s.vm(),
			}
			runner.run()
		})
	}
}

func (s *suiteRunner) vm() func(code, name string) (string, error) {
	jvm := jsonnet.MakeVM()
	jvm.Importer(&jsonnet.FileImporter{JPaths: []string{s.genDir}})
	return func(code, name string) (string, error) {
		return jvm.EvaluateAnonymousSnippet(name, code)
	}
}

func TestAcceptance(t *testing.T) {
	suiteFiles, err := filepath.Glob("testdata/*/suite.json")
	require.NoError(t, err)
	for _, suiteFile := range suiteFiles {
		name := filepath.Base(filepath.Dir(suiteFile))
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(suiteFile)
			require.NoError(t, err)
			var suite Suite
			err = json.Unmarshal(b, &suite)
			require.NoError(t, err)
			runner := &suiteRunner{
				t:      t,
				dir:    filepath.Dir(suiteFile),
				config: suite,
			}
			runner.run()
		})
	}
}
