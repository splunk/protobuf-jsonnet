package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"cd.splunkdev.com/kanantheswaran/protobuf-jsonnet/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func checkMeta(t *testing.T, msg *Message, expectedFile string) {
	b, err := ioutil.ReadFile(expectedFile)
	require.NoError(t, err)
	var expected map[string]FieldMeta
	err = json.Unmarshal(b, &expected)
	require.NoError(t, err)
	assert.EqualValues(t, expected, msg.FieldMeta())
}

func fieldsByName(m *Message) map[string]*Field {
	ret := map[string]*Field{}
	for _, f := range m.fields {
		ret[f.Name()] = f
	}
	return ret
}

func dump(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

func dumpMeta(msg *Message) {
	dump(msg.FieldMeta())
}

func TestLoad(t *testing.T) {
	req := testutil.Request(t, testutil.ProtocConfig{
		Files:        []string{"simple/simple.proto"},
		IncludePaths: []string{"testdata"},
	})
	ds := &descriptorpb.FileDescriptorSet{File: req.GetProtoFile()}
	res := Load(ds)
	r := require.New(t)
	a := assert.New(t)
	a.Equal(7, len(res))
	//dump(res)

	topMsg := res["testdata.simple.TopMessage"]
	r.NotNil(topMsg)
	a.True(topMsg.IsTopLevel())
	a.Equal("TopMessage", topMsg.Name())
	a.Equal("TopMessage", topMsg.NestedName())
	a.Equal("testdata.simple.TopMessage", topMsg.QualifiedName())
	a.Equal("testdata.simple", topMsg.Package())
	r.Nil(topMsg.GetEnum())
	r.NotNil(topMsg.GetMessage())
	a.Equal(topMsg, topMsg.GetMessage())

	msg := topMsg.GetMessage()
	//dumpMeta(msg)
	flds := msg.Fields()
	a.Equal(16, len(flds))
	checkMeta(t, msg, "testdata/simple/top-message-field-meta.json")
	fldMap := fieldsByName(msg)
	for _, f := range fldMap {
		a.False(f.IsRequired())
		a.False(f.IsMap())
		a.False(f.IsList())
		a.EqualValues("", f.ContainerType())
		if f.Name() != "enum_field" {
			a.EqualValues("primitive", f.FieldType())
		} else {
			a.EqualValues("enum", f.FieldType())
			a.EqualValues("testdata.simple.TopLevelEnum", f.TypeName())
		}
		a.Equal("", f.OneOfGroup())
		a.Nil(f.ValidationRules())
	}

	msg = res["testdata.simple.TopMessage.InnerMessage1"].GetMessage()
	//dumpMeta(msg)
	a.False(msg.IsTopLevel())
	checkMeta(t, msg, "testdata/simple/inner-message1-field-meta.json")
	fldMap = fieldsByName(msg)
	f := fldMap["numbers"]
	a.EqualValues("list", f.ContainerType())
	a.False(f.IsRequired())
	a.False(f.IsMap())
	a.True(f.IsList())

	msg = res["testdata.simple.TopMessage.InnerMessage2"].GetMessage()
	//dumpMeta(msg)
	a.False(msg.IsTopLevel())
	checkMeta(t, msg, "testdata/simple/inner-message2-field-meta.json")
	oofs := msg.OneOfs()
	a.EqualValues(1, len(oofs))
	oof := oofs[0]
	a.Equal("main_or_stub", oof.Group)
	a.Equal([]string{"main", "stub"}, oof.Fields)
	a.False(oof.Required)
	fldMap = fieldsByName(msg)
	f = fldMap["msgs"]
	a.True(f.IsMap())
	a.EqualValues("message", f.FieldType())
	a.EqualValues("map", f.ContainerType())
	a.EqualValues("testdata.simple.TopMessage.InnerMessage1", f.TypeName())
	f = fldMap["simple_map"]
	a.Equal("withSimpleMap", f.SetterName())

	e := res["testdata.simple.TopLevelEnum"].GetEnum()
	a.EqualValues(map[string]string{"FIRST": "FIRST", "SECOND": "SECOND", "THIRD": "THIRD"}, e.Map())
	a.EqualValues(map[string]string{"0": "FIRST", "1": "SECOND", "2": "THIRD"}, e.ReverseMap())
	a.Equal("FIRST", e.NameForFirstValue())
}

func TestNullEnum(t *testing.T) {
	var e *Enum
	assert.Equal(t, "UNKNOWN", e.NameForFirstValue())
	e = &Enum{}
	assert.Equal(t, "UNKNOWN", e.NameForFirstValue())
}

func TestValidate(t *testing.T) {
	req := testutil.Request(t, testutil.ProtocConfig{
		Files:        []string{"genvalidate/message.proto"},
		IncludePaths: []string{"testdata", ".."},
	})
	ds := &descriptorpb.FileDescriptorSet{File: req.GetProtoFile()}
	res := Load(ds)
	topMsg := res["testdata.genvalidate.TopMessage"]
	dumpMeta(topMsg.GetMessage())
	checkMeta(t, topMsg.GetMessage(), "testdata/genvalidate/top-message-field-meta.json")
}
