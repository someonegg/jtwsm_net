// Copyright 2013 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmplfunc

import (
	"fmt"
	"reflect"
)

func indirect(v reflect.Value) reflect.Value {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
	}
	return v
}

func numOrStr(k reflect.Kind) bool {
	return isNum(k) || isStr(k)
}

func isNum(k reflect.Kind) bool {
	return isInt(k) || isUint(k) || isFloat(k)
}

func isInt(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUint(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloat(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func isStr(k reflect.Kind) bool {
	switch k {
	case reflect.String:
		return true
	default:
		return false
	}
}

func operandB(x, y interface{}, allowStr bool) (xv, yv reflect.Value) {
	xv = indirect(reflect.ValueOf(x))
	yv = indirect(reflect.ValueOf(y))
	xk := xv.Kind()
	yk := yv.Kind()

	if allowStr {
		if !numOrStr(xk) {
			panic(fmt.Sprintf("[%v] is not number or string", x))
		}
		if !numOrStr(yk) {
			panic(fmt.Sprintf("[%v] is not number or string", y))
		}

	} else {
		if !isNum(xk) {
			panic(fmt.Sprintf("[%v] is not number", x))
		}
		if !isNum(yk) {
			panic(fmt.Sprintf("[%v] is not number", y))
		}
	}

	if xk == yk {
		return
	}

	if isStr(xk) && isStr(yk) {
		return
	}
	if isStr(xk) {
		yv = reflect.ValueOf(fmt.Sprint(yv.Interface()))
		return
	}
	if isStr(yk) {
		xv = reflect.ValueOf(fmt.Sprint(xv.Interface()))
		return
	}

	if isFloat(xk) && isFloat(yk) {
		return
	}
	if isFloat(xk) {
		if isInt(yk) {
			yv = reflect.ValueOf(float64(yv.Int()))
		} else {
			yv = reflect.ValueOf(float64(yv.Uint()))
		}
		return
	}
	if isFloat(yk) {
		if isInt(xk) {
			xv = reflect.ValueOf(float64(xv.Int()))
		} else {
			xv = reflect.ValueOf(float64(xv.Uint()))
		}
		return
	}

	if isUint(xk) && isUint(yk) {
		return
	}

	if isInt(xk) && isInt(yk) {
		return
	}
	if isUint(xk) {
		xv = reflect.ValueOf(int64(xv.Uint()))
		return
	}
	if isUint(yk) {
		yv = reflect.ValueOf(int64(yv.Uint()))
		return
	}

	panic("unreachable code")
}
