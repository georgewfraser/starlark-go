package starlark

import "fmt"

// Test-only built-in for accessing thread inputs.

// inputValue is a Starlark value that exposes an input handle via its
// attribute "value" which may be read and written.
type inputValue struct{ handle *Input }

var (
	_ Value       = (*inputValue)(nil)
	_ HasAttrs    = (*inputValue)(nil)
	_ HasSetField = (*inputValue)(nil)
)

func (iv *inputValue) String() string { return "input" }
func (iv *inputValue) Type() string   { return "input" }
func (iv *inputValue) Freeze()        {}
func (iv *inputValue) Truth() Bool {
	return iv.handle.Value() != None
}
func (iv *inputValue) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: input") }

func (iv *inputValue) Attr(name string) (Value, error) {
	if name != "value" {
		return nil, NoSuchAttrError("input has no ." + name + " attribute")
	}
	return iv.handle.Value(), nil
}

func (iv *inputValue) AttrNames() []string { return []string{"value"} }

func (iv *inputValue) SetField(name string, val Value) error {
	if name != "value" {
		return NoSuchAttrError("input has no ." + name + " field")
	}
	iv.handle.Update(val)
	return nil
}

var InputBuiltin = NewBuiltin("input", func(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var name string
	var def Value = None
	if err := UnpackArgs(b.Name(), args, kwargs, "name", &name, "default?", &def); err != nil {
		return nil, err
	}
	h := thread.cache.Input(thread, name, def)
	return &inputValue{handle: h}, nil
})
