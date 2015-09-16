// Copyright 2013 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmplfunc

import (
	"fmt"
	"reflect"
)

func numsADD(x, y interface{}) (r interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func add panic [%v]", e)
		}
	}()
	xv, yv := operandB(x, y, true)
	var t reflect.Type
	if xv.Type().Size() > yv.Type().Size() {
		t = xv.Type()
	} else {
		t = yv.Type()
	}
	switch k := t.Kind(); {
	case isInt(k):
		i := xv.Int() + yv.Int()
		switch k {
		case reflect.Int:
			return int(i), nil
		case reflect.Int8:
			return int8(i), nil
		case reflect.Int16:
			return int16(i), nil
		case reflect.Int32:
			return int32(i), nil
		case reflect.Int64:
			return int64(i), nil
		}
	case isUint(k):
		u := xv.Uint() + yv.Uint()
		switch k {
		case reflect.Uint:
			return uint(u), nil
		case reflect.Uint8:
			return uint8(u), nil
		case reflect.Uint16:
			return uint16(u), nil
		case reflect.Uint32:
			return uint32(u), nil
		case reflect.Uint64:
			return uint64(u), nil
		}
	case isFloat(k):
		f := xv.Float() + yv.Float()
		switch k {
		case reflect.Float32:
			return float32(f), nil
		case reflect.Float64:
			return float64(f), nil
		}
	case isStr(k):
		return xv.String() + yv.String(), nil
	}
	panic("unreachable code")
}

func numSUB(x, y interface{}) (r interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func add panic [%v]", e)
		}
	}()
	xv, yv := operandB(x, y, false)
	var t reflect.Type
	if xv.Type().Size() > yv.Type().Size() {
		t = xv.Type()
	} else {
		t = yv.Type()
	}
	switch k := t.Kind(); {
	case isInt(k):
		i := xv.Int() - yv.Int()
		switch k {
		case reflect.Int:
			return int(i), nil
		case reflect.Int8:
			return int8(i), nil
		case reflect.Int16:
			return int16(i), nil
		case reflect.Int32:
			return int32(i), nil
		case reflect.Int64:
			return int64(i), nil
		}
	case isUint(k):
		u := xv.Uint() - yv.Uint()
		switch k {
		case reflect.Uint:
			return uint(u), nil
		case reflect.Uint8:
			return uint8(u), nil
		case reflect.Uint16:
			return uint16(u), nil
		case reflect.Uint32:
			return uint32(u), nil
		case reflect.Uint64:
			return uint64(u), nil
		}
	case isFloat(k):
		f := xv.Float() - yv.Float()
		switch k {
		case reflect.Float32:
			return float32(f), nil
		case reflect.Float64:
			return float64(f), nil
		}
	}
	panic("unreachable code")
}
