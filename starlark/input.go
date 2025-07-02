package starlark

import "fmt"

// inputValue is a Starlark value that provides access to a Thread input.
// It exposes a single attribute, "value", which may be read and written.
type inputValue struct {
	name  string
	owner *Thread
}

var (
	_ Value       = (*inputValue)(nil)
	_ HasAttrs    = (*inputValue)(nil)
	_ HasSetField = (*inputValue)(nil)
)

func (i *inputValue) String() string { return fmt.Sprintf("input(%q)", i.name) }
func (i *inputValue) Type() string   { return "input" }
func (i *inputValue) Freeze()        {}
func (i *inputValue) Truth() Bool {
	return i.owner.cache.inputs != nil && i.owner.cache.inputs[i.name] != nil
}
func (i *inputValue) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: input") }

func (i *inputValue) Attr(name string) (Value, error) {
	if name != "value" {
		return nil, NoSuchAttrError(fmt.Sprintf("input has no .%s attribute", name))
	}
	if i.owner.cache.inputs == nil {
		return None, nil
	}
	v, ok := i.owner.cache.inputs[i.name]
	if !ok {
		return None, nil
	}
	i.owner.dependencies.inputs = append(i.owner.dependencies.inputs, InputValue{
		name:  i.name,
		value: i.owner.cache.Intern(v),
	})
	return v, nil
}

func (i *inputValue) AttrNames() []string { return []string{"value"} }

func (i *inputValue) SetField(name string, val Value) error {
	if name != "value" {
		return NoSuchAttrError(fmt.Sprintf("input has no .%s field", name))
	}
	if i.owner.cache.inputs == nil {
		i.owner.cache.inputs = make(map[string]Value)
	}
	i.owner.cache.version++
	i.owner.cache.inputs[i.name] = val
	i.owner.dependencies.inputs = append(i.owner.dependencies.inputs, InputValue{
		name:  i.name,
		value: i.owner.cache.Intern(val),
	})
	return nil
}

// inputBuiltin returns a handle to the specified input.
func inputBuiltin(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var name string
	if err := UnpackArgs(b.name, args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	return &inputValue{owner: thread, name: name}, nil
}

// InputBuiltin is a convenience value that clients may use as a predeclared
// built-in to access Thread inputs.
var InputBuiltin = NewBuiltin("input", inputBuiltin)
