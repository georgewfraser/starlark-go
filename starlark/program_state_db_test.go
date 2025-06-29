package starlark

import (
	"testing"

	"go.starlark.net/internal/compile"
)

// helpers produce dynamic values to avoid compile-time optimizations.
func dynamicInt(i int) int          { return i + 0 }
func dynamicString(s string) string { return s + "" }

func TestInternDistinctObjectsNotEqual(t *testing.T) {
	db := NewProgramStateDB()
	s1 := Value(String(dynamicString("a")))
	s2 := Value(String(dynamicString("a")))
	i1 := db.Intern(s1)
	i2 := db.Intern(s2)
	if i1.Eq(i2) {
		t.Fatalf("interned values of distinct objects should not be equal")
	}
}

func TestInternSameObjectEqual(t *testing.T) {
	db := NewProgramStateDB()
	v := Value(String(dynamicString("b")))
	i1 := db.Intern(v)
	i2 := db.Intern(v)
	if !i1.Eq(i2) {
		t.Fatalf("interning the same object twice should yield equal references")
	}
}

func TestInternValueRoundTrip(t *testing.T) {
	db := NewProgramStateDB()
	v := Value(String(dynamicString("foo")))
	i := db.Intern(v)
	if got := db.Value(i); got != v {
		t.Fatalf("expected round trip value, got %v", got)
	}
}

func TestInternEmpty(t *testing.T) {
	var zero Interned
	if !zero.Empty() {
		t.Fatalf("zero Interned should report empty")
	}
	db := NewProgramStateDB()
	nonzero := db.Intern(MakeInt(dynamicInt(1)))
	if nonzero.Empty() {
		t.Fatalf("non-empty Interned reported empty")
	}
}

func TestProgramStateDBPutGet(t *testing.T) {
	db := NewProgramStateDB()
	fn := &Function{}
	arg := db.Intern(MakeInt(dynamicInt(42)))
	result := db.Intern(String("result"))
	deps := Dependencies{globals: []VariableValue{{variable: 1, value: arg}}}

	db.Put(fn, []Interned{arg}, deps, result, 7)
	rec := db.Get(fn, []Interned{arg})
	if rec == nil {
		t.Fatalf("expected cached record")
	}
	if rec.function != fn || !rec.result.Eq(result) || len(rec.args) != 1 || !rec.args[0].Eq(arg) {
		t.Fatalf("record contents mismatch")
	}
	if rec.deps.globals[0].variable != 1 || !rec.deps.globals[0].value.Eq(arg) {
		t.Fatalf("observed globals mismatch")
	}
}

func TestProgramStateDBGetWrongArgument(t *testing.T) {
	db := NewProgramStateDB()
	fn := &Function{}
	arg := db.Intern(MakeInt(dynamicInt(1)))
	result := db.Intern(String("ok"))
	db.Put(fn, []Interned{arg}, Dependencies{}, result, 0)

	miss := db.Get(fn, []Interned{db.Intern(MakeInt(dynamicInt(2)))})
	if miss != nil {
		t.Fatalf("expected cache miss for different argument")
	}
}

func TestProgramStateDBGetWrongFunction(t *testing.T) {
	db := NewProgramStateDB()
	fn1 := &Function{funcode: &compile.Funcode{}}
	fn2 := &Function{funcode: &compile.Funcode{}}
	arg := db.Intern(MakeInt(dynamicInt(3)))
	result := db.Intern(String("x"))
	db.Put(fn1, []Interned{arg}, Dependencies{}, result, 0)

	miss := db.Get(fn2, []Interned{arg})
	if miss != nil {
		t.Fatalf("expected cache miss for different function")
	}
}
