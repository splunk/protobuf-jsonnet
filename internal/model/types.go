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

package model

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"cd.splunkdev.com/kanantheswaran/protobuf-jsonnet/internal/validate"
	"google.golang.org/protobuf/types/descriptorpb"
)

// FieldType indicates the type of field, "primitive", "enum" or "message".
type FieldType string

const (
	FieldTypePrimitive FieldType = "primitive"
	FieldTypeEnum      FieldType = "enum"
	FieldTypeMessage   FieldType = "message"
)

// ContainerType indicates if the field is a container type with values none (empty string), "list", and "map"
type ContainerType string

const (
	ContainerTypeNone ContainerType = ""
	ContainerTypeList ContainerType = "list"
	ContainerTypeMap  ContainerType = "map"
)

var (
	optionSkipValidateRules = false
)

func SetSkipValidations(flag bool) {
	optionSkipValidateRules = flag
}

// base is a message or an enum, possibly nested under another type.
type base struct {
	pkg     string   // the package in which it belongs
	name    string   // type name
	parents []string // list of local type names that are its parent, immediate last
}

// Name returns the base name of the type
func (b *base) Name() string {
	return b.name
}

// Package returns the package name in which the type is defined or the empty string if there is no package.
func (b *base) Package() string {
	return b.pkg
}

// IsTopLevel returns true if the type definition is not nested under a message.
func (b *base) IsTopLevel() bool {
	return len(b.parents) == 0
}

// NestedName returns the qualified name not including package.
func (b *base) NestedName() string {
	if len(b.parents) == 0 {
		return b.name
	}
	return strings.Join(b.parents, ".") + "." + b.name
}

// QualifiedName returns the qualified name of the type including the package name.
func (b *base) QualifiedName() string {
	name := b.NestedName()
	if b.pkg == "" {
		return name
	}
	return b.pkg + "." + name
}

// Field is a field in a message.
type Field struct {
	f          *descriptorpb.FieldDescriptorProto
	ft         FieldType
	ct         ContainerType
	typeName   string
	oneOfGroup string
	rules      *validate.FieldRules
}

// Name returns the canonical name for the field.
func (f *Field) Name() string {
	return f.f.GetName()
}

// JsonName returns the JSON name for the field.
func (f *Field) JsonName() string {
	return f.f.GetJsonName()
}

// AllowedNames returns the set of names allowed to refer to this field.
func (f *Field) AllowedNames() []string {
	ret := []string{f.Name()}
	if f.Name() != f.JsonName() {
		ret = append(ret, f.JsonName())
	}
	return ret
}

// OneOfGroup returns the name of the one-of if the field participates in one.
func (f *Field) OneOfGroup() string {
	return f.oneOfGroup
}

// FieldType returns the field type of the field.
func (f *Field) FieldType() FieldType {
	return f.ft
}

// ContainerType returns the container type for the field.
func (f *Field) ContainerType() ContainerType {
	return f.ct
}

// TypeName returns the fully qualified type name for messages or a primitive type name.
// All numeric types are represented as a "number" string.
func (f *Field) TypeName() string {
	return f.typeName
}

// IsList returns true if the field is a list.
func (f *Field) IsList() bool {
	return f.ct == ContainerTypeList
}

// IsMap returns true if the field is a map.
func (f *Field) IsMap() bool {
	return f.ct == ContainerTypeMap
}

// SetterName returns the name of the setter to be used for this type in generated code.
func (f *Field) SetterName() string {
	name := f.JsonName()
	return "with" + strings.ToUpper(name[0:1]) + name[1:]
}

// ValidationRules returns any field rules defined for this type using the protoc-gen-validate package.
func (f *Field) ValidationRules() *validate.FieldRules {
	return f.rules
}

// IsRequired returns true if the field is required to be present.
func (f *Field) IsRequired() bool {
	if f.rules == nil {
		return false
	}
	switch f.ContainerType() {
	case ContainerTypeNone:
		reqd := f.rules.Message != nil && f.rules.Message.GetRequired()
		if reqd {
			// in languages where the type is determined by reflection it is possible to
			// have a one of type set where the underlying message is nil. But in JSON, the inner type
			// is actually determined by the presence of the field, so we cannot get into a situation
			// where the field exists but is not-nil.
			// Therefore, we honor the required validation flag only for non one-of types.
			if f.f.OneofIndex == nil {
				return true
			}
		}
	case ContainerTypeList:
		if f.rules.GetRepeated() != nil && !f.rules.GetRepeated().GetIgnoreEmpty() && f.rules.GetRepeated().GetMinItems() > 0 {
			return true
		}
	case ContainerTypeMap:
		if f.rules.GetMap() != nil && !f.rules.GetMap().GetIgnoreEmpty() && f.rules.GetMap().GetMinPairs() > 0 {
			return true
		}
	}
	return false
}

// Constraints returns field rules as a JSON string.
func (f *Field) Constraints() map[string]interface{} {
	if f.rules == nil || f.rules.Type == nil {
		return nil
	}
	b, _ := json.Marshal(f.rules.Type)
	var ret map[string]interface{}
	_ = json.Unmarshal(b, &ret)
	return ret
}

// Enum is a protobuf enum definition.
type Enum struct {
	base
	e *descriptorpb.EnumDescriptorProto
}

// GetEnum implements the Type interface.
func (e *Enum) GetEnum() *Enum {
	return e
}

// GetMessage implements the Type interface.
func (e *Enum) GetMessage() *Message {
	return nil
}

// Map returns a map of enum keys to keys.
func (e *Enum) Map() map[string]string {
	ret := map[string]string{}
	for _, v := range e.e.GetValue() {
		ret[v.GetName()] = v.GetName()
	}
	return ret
}

// ValueMap returns a map of enum values converted to string to keys.
func (e *Enum) ValueMap() map[string]string {
	ret := map[string]string{}
	for _, v := range e.e.GetValue() {
		ret[v.GetName()] = fmt.Sprint(v.GetNumber())
	}
	return ret
}

// ReverseMap returns a map of names keyed by enum values converted to strings.
func (e *Enum) ReverseMap() map[string]string {
	ret := map[string]string{}
	for _, v := range e.e.GetValue() {
		ret[fmt.Sprint(v.GetNumber())] = v.GetName()
	}
	return ret
}

// NameForFirstValue returns the name for the first value defined in the enum.
func (e *Enum) NameForFirstValue() string {
	if e == nil || e.e == nil || len(e.e.GetValue()) == 0 {
		return "UNKNOWN"
	}
	return e.e.GetValue()[0].GetName()
}

func newEnum(pkg string, e *descriptorpb.EnumDescriptorProto, parent *Message) *Enum {
	b := base{
		name: e.GetName(),
		pkg:  pkg,
	}
	if parent != nil {
		b.parents = append(parent.parents[:], parent.name)
	}
	ret := &Enum{
		base: b,
		e:    e,
	}
	return ret
}

// OneOf represents a one-of field.
type OneOf struct {
	Fields   []string `json:"fields"`   // the field names that constitute the one-of
	Required bool     `json:"required"` // whether it is required in the enclosing message
	Group    string   `json:"group"`    // the name of the one-of field.
}

// Message represents a protobuf message.
type Message struct {
	base
	m              *descriptorpb.DescriptorProto
	fields         []*Field
	oneOfs         []*OneOf
	nestedMessages []*Message
	nestedEnums    []*Enum
}

// GetEnum implements the Type interface.
func (m *Message) GetEnum() *Enum {
	return nil
}

// GetMessage implements the Type interface.
func (m *Message) GetMessage() *Message {
	return m
}

// Fields returns the fields for this message.
func (m Message) Fields() []*Field {
	return m.fields
}

// NestedMessages returns the messages nested under this type.
func (m *Message) NestedMessages() []*Message {
	return m.nestedMessages
}

// NestedEnums returns the enums nested under this type.
func (m *Message) NestedEnums() []*Enum {
	return m.nestedEnums
}

// FieldMeta returns the attributes for a field needed in generated code.
type FieldMeta struct {
	Type          string                 `json:"type"`                    // the type name of the field as returned by
	AllowedNames  []string               `json:"allowedNames"`            // the allowed names for the field
	ContainerType ContainerType          `json:"containerType,omitempty"` // the container type
	Required      bool                   `json:"required,omitempty"`      // whether it is required
	Constraints   map[string]interface{} `json:"constraints,omitempty"`   // type constraints associated with the field
}

// FieldMeta returns a map of field metadata keyed by field name.
func (m *Message) FieldMeta() map[string]FieldMeta {
	ret := map[string]FieldMeta{}
	for _, f := range m.fields {
		meta := FieldMeta{
			AllowedNames:  f.AllowedNames(),
			Type:          f.TypeName(),
			ContainerType: f.ContainerType(),
			Required:      f.IsRequired(),
			Constraints:   f.Constraints(),
		}
		ret[f.Name()] = meta
	}
	return ret
}

// OneOfs returns the one-of defined for this message.
func (m *Message) OneOfs() []*OneOf {
	return m.oneOfs
}

func extractFieldTypeAndName(f *descriptorpb.FieldDescriptorProto) (fType FieldType, name string) {
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return FieldTypeMessage, strings.TrimPrefix(f.GetTypeName(), ".")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return FieldTypeEnum, strings.TrimPrefix(f.GetTypeName(), ".")
	default:
		kebabName := descriptorpb.FieldDescriptorProto_Type_name[int32(f.GetType())]
		return FieldTypePrimitive, strings.ToLower(strings.TrimPrefix(kebabName, "TYPE_"))
	}
}

// newMessage creates a new message.
func newMessage(pkg string, m *descriptorpb.DescriptorProto, parent *Message) *Message {
	b := base{
		name: m.GetName(),
		pkg:  pkg,
	}
	if parent != nil {
		b.parents = append(parent.parents[:], parent.name)
	}
	ret := &Message{
		base:   b,
		m:      m,
		oneOfs: []*OneOf{}, // use an empty array instead of nil if no entries exist
	}

	var err error
	disableValidation := optionSkipValidateRules
	if !disableValidation {
		disableValidation, err = shouldDisableValidation(m.GetOptions())
		if err != nil {
			log.Printf("Error getting disable options, %v, continue", err)
		}
	}

	for _, o := range m.GetOneofDecl() {
		var reqd bool
		if !disableValidation {
			reqd, err = isOneOfRequired(o.GetOptions())
			if err != nil {
				log.Printf("Error getting OneOf options, %v, continue", err)
			}
		}
		ret.oneOfs = append(ret.oneOfs, &OneOf{Group: o.GetName(), Required: reqd})
	}
	for _, f := range m.GetField() {
		var rules *validate.FieldRules
		if !disableValidation {
			rules, err = getValidationRules(f.GetOptions())
			if err != nil {
				log.Printf("Error getting validation rules for field %s in message %s, %v, continue", f.GetName(), ret.QualifiedName(), err)
			}
		}
		var oneOfGroup string
		if f.OneofIndex != nil {
			i := int(f.GetOneofIndex())
			ret.oneOfs[i].Fields = append(ret.oneOfs[i].Fields, f.GetName())
			oneOfGroup = ret.oneOfs[i].Group
		}
		fType, name := extractFieldTypeAndName(f)
		ct := ContainerTypeNone
		if f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			ct = ContainerTypeList // note: maps are handled in a second pass after all types have been processed
		}

		ret.fields = append(ret.fields, &Field{
			f:          f,
			ft:         fType,
			ct:         ct,
			typeName:   name,
			rules:      rules,
			oneOfGroup: oneOfGroup,
		})
	}
	var nm []*Message
	for _, t := range m.GetNestedType() {
		nm = append(nm, newMessage(pkg, t, ret))
	}
	ret.nestedMessages = nm

	var ne []*Enum
	for _, t := range m.GetEnumType() {
		ne = append(ne, newEnum(pkg, t, ret))
	}
	ret.nestedMessages = nm
	ret.nestedEnums = ne

	sort.Slice(ret.fields, func(i, j int) bool {
		left := ret.fields[i]
		right := ret.fields[j]
		return left.Name() < right.Name()
	})

	return ret
}
