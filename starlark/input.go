package starlark

import "fmt"

// inputBuiltin returns the value of the specified input from thread.inputs
// and records the dependency for incremental execution.
func inputBuiltin(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var name string
	if err := UnpackArgs(b.name, args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	if thread.cache.inputs == nil {
		return nil, fmt.Errorf("input %s: no inputs available", name)
	}
	v, ok := thread.cache.inputs[name]
	if !ok {
		return nil, fmt.Errorf("input %s not provided", name)
	}
	thread.dependencies.inputs = append(thread.dependencies.inputs, InputValue{
		name:  name,
		value: thread.cache.Intern(v),
	})
	return v, nil
}

// InputBuiltin is a convenience value that clients may use as a predeclared
// built-in to access Thread inputs.
var InputBuiltin = NewBuiltin("input", inputBuiltin)
