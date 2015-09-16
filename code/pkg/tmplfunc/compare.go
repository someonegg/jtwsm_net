// Copyright 2013 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmplfunc

import (
	"fmt"
	_ "reflect"
)

func equal(x, y interface{}) (r bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func equal panic [%v]", e)
		}
	}()
	return x == y, nil
}

func numsEQ(x, y interface{}) (r bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func eq panic [%v]", e)
		}
	}()
	xv, yv := operandB(x, y, true)
	switch k := xv.Kind(); {
	case isInt(k):
		return xv.Int() == yv.Int(), nil
	case isUint(k):
		return xv.Uint() == yv.Uint(), nil
	case isFloat(k):
		return xv.Float() == yv.Float(), nil
	case isStr(k):
		return xv.String() == yv.String(), nil
	}
	panic("unreachable code")
}

func numsLT(x, y interface{}) (r bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func lt panic [%v]", e)
		}
	}()
	xv, yv := operandB(x, y, true)
	switch k := xv.Kind(); {
	case isInt(k):
		return xv.Int() < yv.Int(), nil
	case isUint(k):
		return xv.Uint() < yv.Uint(), nil
	case isFloat(k):
		return xv.Float() < yv.Float(), nil
	case isStr(k):
		return xv.String() < yv.String(), nil
	}
	panic("unreachable code")
}

func numsLE(x, y interface{}) (r bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("func le panic [%v]", e)
		}
	}()
	xv, yv := operandB(x, y, true)
	switch k := xv.Kind(); {
	case isInt(k):
		return xv.Int() <= yv.Int(), nil
	case isUint(k):
		return xv.Uint() <= yv.Uint(), nil
	case isFloat(k):
		return xv.Float() <= yv.Float(), nil
	case isStr(k):
		return xv.String() <= yv.String(), nil
	}
	panic("unreachable code")
}

func numsGT(x, y interface{}) (r bool, err error) {
	return numsLT(y, x)
}

func numsGE(x, y interface{}) (r bool, err error) {
	return numsLE(y, x)
}
