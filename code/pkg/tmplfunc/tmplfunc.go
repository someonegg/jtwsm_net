// Copyright 2013 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tmplfunc provides some general functions to template.
package tmplfunc

import (
	"text/template"
)

var FuncMap = template.FuncMap{
	"equal": equal,
	"eq":    numsEQ,
	"lt":    numsLT,
	"le":    numsLE,
	"gt":    numsGT,
	"ge":    numsGE,
	"add":   numsADD,
	"sub":   numSUB,
}
