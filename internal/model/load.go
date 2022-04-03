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

import "google.golang.org/protobuf/types/descriptorpb"

// Type is an abstraction over a protobuf message or enum. It exposes the common attributes
// for each. The GetMessage and GetEnum methods respectively return a non-nil Message or
// Enum object based on what the underlying type is.
type Type interface {
	Name() string          // Leaf name for the type
	Package() string       // Package in which type is declared
	IsTopLevel() bool      // returns true if this is a type not nested in a message
	NestedName() string    // the qualified name of the type not including package name
	QualifiedName() string // the qualified name of the type including package name
	GetMessage() *Message  // the underlying message if the type represents a message, or nil
	GetEnum() *Enum        // the underlying Enum if the type represents an enum, or nil
}

type loader struct {
	ret map[string]Type
}

func (c *loader) registerType(t Type) {
	c.ret[t.QualifiedName()] = t
}

func (c *loader) addNestedTypes(parent *Message) {
	childEnums := parent.NestedEnums()
	for _, child := range childEnums {
		c.registerType(child)
	}
	childMessages := parent.NestedMessages()
	for _, child := range childMessages {
		c.registerType(child)
		c.addNestedTypes(child)
	}
}

func (c *loader) getMapType(t Type) (found bool, name string) {
	msg := t.GetMessage()
	if msg == nil {
		return false, ""
	}
	if msg.m.Options == nil || msg.m.Options.MapEntry == nil {
		return false, ""
	}
	if !msg.m.GetOptions().GetMapEntry() {
		return false, ""
	}
	// get the field type of the "value" field of the map (for JSON purposes, key is always string)
	for _, mapField := range msg.fields {
		if mapField.Name() != "value" {
			continue
		}
		return true, mapField.TypeName()
	}
	return false, ""
}

func (c *loader) updateMapTypes() {
	for _, v := range c.ret {
		if v.GetMessage() == nil {
			continue
		}
		// a field of type map will appear as a repeated field of a message whole map_entry option is set to true
		for _, f := range v.GetMessage().fields {
			if f.ct != ContainerTypeList {
				continue
			}
			t, ok := c.ret[f.typeName]
			if !ok {
				continue
			}
			found, valueType := c.getMapType(t)
			if found {
				f.ct = ContainerTypeMap
				f.typeName = valueType
			}
		}
	}
}

// Load returns the types found in the specified descriptor set keyed by fully qualified name.
func Load(ds *descriptorpb.FileDescriptorSet) map[string]Type {
	l := &loader{ret: map[string]Type{}}
	for _, file := range ds.GetFile() {
		pkg := file.GetPackage()
		for _, e := range file.GetEnumType() {
			en := newEnum(pkg, e, nil)
			l.registerType(en)
		}
		for _, msg := range file.GetMessageType() {
			m := newMessage(pkg, msg, nil)
			l.registerType(m)
			l.addNestedTypes(m)
		}
	}
	l.updateMapTypes()
	return l.ret
}
