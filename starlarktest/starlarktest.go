// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package starlarktest defines utilities for testing Starlark programs.
//
// Clients can call LoadAssertModule to load a module that defines
// several functions useful for testing.  See assert.star for its
// definition.
//
// The assert.error function, which reports errors to the current Go
// testing.T, requires that clients call SetReporter(thread, t) before use.
package starlarktest // import "go.starlark.net/starlarktest"

import (
	_ "embed"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const localKey = "Reporter"

// A Reporter is a value to which errors may be reported.
// It is satisfied by *testing.T.
type Reporter interface {
	Error(args ...interface{})
}

// SetReporter associates an error reporter (such as a testing.T in
// a Go test) with the Starlark thread so that Starlark programs may
// report errors to it.
func SetReporter(thread *starlark.Thread, r Reporter) {
	thread.SetLocal(localKey, r)
}

// GetReporter returns the Starlark thread's error reporter.
// It must be preceded by a call to SetReporter.
func GetReporter(thread *starlark.Thread) Reporter {
	r, ok := thread.Local(localKey).(Reporter)
	if !ok {
		panic("internal error: starlarktest.SetReporter was not called")
	}
	return r
}

var (
	once   sync.Once
	assert starlark.StringDict
	//go:embed assert.star
	assertFileSrc string
	assertErr     error
)

// LoadAssertModule loads the assert module.
// It is concurrency-safe and idempotent.
func LoadAssertModule() (starlark.StringDict, error) {
	once.Do(func() {
		predeclared := starlark.StringDict{
			"error":    starlark.NewBuiltinWithEffects("error", error_),
			"catch":    starlark.NewBuiltin("catch", catch),
			"matches":  starlark.NewBuiltin("matches", matches),
			"module":   starlark.NewBuiltin("module", starlarkstruct.MakeModule),
			"_freeze":  starlark.NewBuiltin("freeze", freeze),
			"_floateq": starlark.NewBuiltin("floateq", floateq),
		}
		thread := new(starlark.Thread)
		assert, assertErr = starlark.ExecFile(thread, "assert.star", assertFileSrc, predeclared)
	})
	return assert, assertErr
}

// catch(f) evaluates f() and returns its evaluation error message
// if it failed or None if it succeeded.
func catch(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fn starlark.Callable
	if err := starlark.UnpackArgs("catch", args, kwargs, "fn", &fn); err != nil {
		return nil, err
	}
	if _, err := starlark.Call(thread, fn, nil, nil); err != nil {
		return starlark.String(err.Error()), nil
	}
	return starlark.None, nil
}

// matches(pattern, str) reports whether string str matches the regular expression pattern.
func matches(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern, str string
	if err := starlark.UnpackArgs("matches", args, kwargs, "pattern", &pattern, "str", &str); err != nil {
		return nil, err
	}
	ok, err := regexp.MatchString(pattern, str)
	if err != nil {
		return nil, fmt.Errorf("matches: %s", err)
	}
	return starlark.Bool(ok), nil
}

// error(x) reports an error to the Go test framework.
func error_(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("error: got %d arguments, want 1", len(args))
	}
	buf := new(strings.Builder)
	stk := thread.CallStack()
	stk.Pop()
	fmt.Fprintf(buf, "%sError: ", stk)
	if s, ok := starlark.AsString(args[0]); ok {
		buf.WriteString(s)
	} else {
		buf.WriteString(args[0].String())
	}
	GetReporter(thread).Error(buf.String())
	return starlark.None, nil
}

// freeze(x) freezes its operand.
func freeze(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(kwargs) > 0 {
		return nil, fmt.Errorf("freeze does not accept keyword arguments")
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("freeze got %d arguments, wants 1", len(args))
	}
	args[0].Freeze()
	return args[0], nil
}

// floateq(x, y) reports whether two floats are within 1 ULP of each other.
func floateq(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var xf, yf starlark.Float
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 2, &xf, &yf); err != nil {
		return nil, err
	}

	res := false
	switch {
	case xf == yf:
		res = true
	case math.IsNaN(float64(xf)):
		res = math.IsNaN(float64(yf))
	case math.IsNaN(float64(yf)):
		// false (non-NaN = Nan)
	default:
		x := math.Float64bits(float64(xf))
		y := math.Float64bits(float64(yf))
		res = x == y+1 || y == x+1
	}
	return starlark.Bool(res), nil
}

// DataFile returns the effective filename of the specified
// test data resource.  The function abstracts differences between
// 'go build', under which a test runs in its package directory,
// and Blaze, under which a test runs in the root of the tree.
var DataFile = func(pkgdir, filename string) string {
	// Check if we're being run by Bazel and change directories if so.
	// TEST_SRCDIR and TEST_WORKSPACE are set by the Bazel test runner, so that makes a decent check
	testSrcdir := os.Getenv("TEST_SRCDIR")
	testWorkspace := os.Getenv("TEST_WORKSPACE")
	if testSrcdir != "" && testWorkspace != "" {
		return filepath.Join(testSrcdir, "net_starlark_go", pkgdir, filename)
	}

	// Under go test, ignore pkgdir, which is the directory of the
	// current package relative to the module root.
	return filename
}
