// Copyright 2013 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package present

import "encoding/json"
import "fmt"

var unmarshalers = make(map[string]UnmarshalFunc)

type UnmarshalWork func(data []byte, v interface{}) error

type UnmarshalFunc func([]byte, UnmarshalWork) (Elem, error)

func RegisterUnmarshaler(name string, unmarshaler UnmarshalFunc) {
	if len(name) == 0 {
		panic("bad name in RegisterUnmarshaler: " + name)
	}
	unmarshalers[name] = unmarshaler
}

type ElemProxy struct {
	Inner interface{}
	Type  string
}

func (ep *ElemProxy) UnmarshalElem(data []byte, worker UnmarshalWork) error {
	var et struct {
		Type string
	}
	err := worker(data, &et)
	if err != nil {
		return err
	}
	unmarshaler := unmarshalers[et.Type]
	if unmarshaler == nil {
		return fmt.Errorf("Unmarshal : unknown type %s\n", et.Type)
	}
	e, err := unmarshaler(data, worker)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("Unmarshal : inner is nil, type is %s\n", et.Type)
	}
	ep.Inner = e
	ep.Type = e.TemplateName()
	return nil
}

func (ep *ElemProxy) UnmarshalJSON(data []byte) error {
	return ep.UnmarshalElem(data, json.Unmarshal)
}
