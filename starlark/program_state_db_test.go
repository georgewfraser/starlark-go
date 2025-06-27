package starlark

import (
	"testing"
)

func TestIntern(t *testing.T) {
	db := NewProgramStateDB()

	s1 := Value(String(dynamicString("1")))
	s2 := Value(String(dynamicString("1")))

	// Large values should be interned as distinct values even if they are equal.
	i1 := db.Intern(s1)
	i1again := db.Intern(s1)
	i2 := db.Intern(s2)
	if i1.Eq(i2) {
		t.Fatalf("expected interned values to be distinct for strings")
	}

	// But if they're the same object, they should be equal.
	if !i1again.Eq(i1) {
		t.Fatalf("expected interned values to be equal for same object")
	}

	// Identical values should be equal when I convert them back to Value.
	if db.Value(i1) != db.Value(i2) {
		t.Fatalf("expected interned values to match when converted back to Value")
	}
}

func dynamicInt(i int) int {
	return i + 0
}

func dynamicString(s string) string {
	return s + ""
}

func TestProgramStateDBPutGet(t *testing.T) {
	db := NewProgramStateDB()
	fn := &Function{}
	arg := db.Intern(MakeInt(dynamicInt(42)))
	res := db.Intern(String("result"))
	free := []VariableValue{
		{variable: 1, value: arg},
		{variable: 2, value: res},
	}
	db.Put(fn, []Interned{arg}, Observed{globals: free}, 0, res)
	rec := db.Get(fn, []Interned{arg})
	if rec == nil {
		t.Fatalf("expected record to be found")
	}
	if !rec.result.Eq(res) || rec.function != fn || len(rec.args) != 1 || !rec.args[0].Eq(arg) {
		t.Fatalf("record mismatch")
	}
	if len(rec.globals) != 2 || rec.globals[0].variable != 1 || !rec.globals[0].value.Eq(arg) ||
		rec.globals[1].variable != 2 || !rec.globals[1].value.Eq(res) {
		t.Fatalf("captures mismatch")
	}

	// request with different arg should miss
	miss := db.Get(fn, []Interned{db.Intern(MakeInt(dynamicInt(43)))})
	if miss != nil {
		t.Fatalf("expected cache miss")
	}
}

func TestProgramStateDBCollision(t *testing.T) {
	// TODO
}
