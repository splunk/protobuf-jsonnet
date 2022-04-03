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
	"fmt"
	"reflect"

	"cd.splunkdev.com/kanantheswaran/protobuf-jsonnet/internal/validate"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func extractExtension(opts proto.Message, e protoreflect.ExtensionType, out interface{}) (found bool, _ error) {
	// mostly copied from https://github.com/lyft/protoc-gen-star/blob/master/extension.go
	if opts == nil || reflect.ValueOf(opts).IsNil() {
		return false, nil
	}
	if e == nil {
		return false, errors.New("nil proto.ExtensionType parameter provided")
	}
	if out == nil {
		return false, errors.New("nil extension output parameter provided")
	}
	o := reflect.ValueOf(out)
	if o.Kind() != reflect.Ptr {
		return false, errors.New("out parameter must be a pointer type")
	}
	if !proto.HasExtension(opts, e) {
		return false, nil
	}

	val := proto.GetExtension(opts, e)
	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	for o.Kind() == reflect.Ptr || o.Kind() == reflect.Interface {
		if o.Kind() == reflect.Ptr && o.IsNil() {
			o.Set(reflect.New(o.Type().Elem()))
		}
		o = o.Elem()
	}

	if v.Type().AssignableTo(o.Type()) {
		o.Set(v)
		return true, nil
	}

	return true, fmt.Errorf("cannot assign extension type %q to output type %q",
		v.Type().String(),
		o.Type().String())
}

func shouldDisableValidation(m proto.Message) (bool, error) {
	ret := false
	_, err := extractExtension(m, validate.E_Disabled, &ret)
	if err != nil {
		return false, err
	}
	return ret, err
}

func isOneOfRequired(m proto.Message) (bool, error) {
	ret := false
	_, err := extractExtension(m, validate.E_Required, &ret)
	if err != nil {
		return false, err
	}
	return ret, err
}

func getValidationRules(m proto.Message) (*validate.FieldRules, error) {
	var ret validate.FieldRules
	found, err := extractExtension(m, validate.E_Rules, &ret)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &ret, nil
}
