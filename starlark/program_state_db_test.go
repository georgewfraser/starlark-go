package starlark

import (
	"testing"

	"go.starlark.net/internal/compile"
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
	prog := &compile.Program{}
	arg := db.Intern(MakeInt(dynamicInt(42)))
	res := db.Intern(String("result"))
	free := []Capture{
		{variable: 1, value: arg},
		{variable: 2, value: res},
	}
	db.Put(prog, 1, []Interned{arg}, Observed{free: free}, res)
	rec := db.Get(prog, 1, []Interned{arg})
	if rec == nil {
		t.Fatalf("expected record to be found")
	}
	if !rec.result.Eq(res) || rec.function != 1 || len(rec.args) != 1 || !rec.args[0].Eq(arg) {
		t.Fatalf("record mismatch")
	}
	if len(rec.free) != 2 || rec.free[0].variable != 1 || !rec.free[0].value.Eq(arg) ||
		rec.free[1].variable != 2 || !rec.free[1].value.Eq(res) {
		t.Fatalf("captures mismatch")
	}

	// request with different arg should miss
	miss := db.Get(prog, 1, []Interned{db.Intern(MakeInt(dynamicInt(43)))})
	if miss != nil {
		t.Fatalf("expected cache miss")
	}
}

func TestProgramStateDBCollision(t *testing.T) {
	db := NewProgramStateDB()
	prog := &compile.Program{}
	arg := db.Intern(MakeInt(dynamicInt(1)))
	r1 := db.Intern(String("one"))
	db.Put(prog, 0, []Interned{arg}, Observed{}, r1)

	// find a second function id that hashes to the same slot
	target := hashKey(prog, 0, []Interned{arg})
	fid2 := 1
	for ; fid2 < CACHE_SIZE*10; fid2++ {
		if hashKey(prog, fid2, []Interned{arg}) == target {
			break
		}
	}
	if hashKey(prog, fid2, []Interned{arg}) != target {
		t.Fatalf("unable to find collision")
	}
	r2 := db.Intern(String("two"))
	db.Put(prog, fid2, []Interned{arg}, Observed{}, r2)

	// first record should be evicted
	if rec := db.Get(prog, 0, []Interned{arg}); rec != nil {
		t.Fatalf("expected eviction of first record")
	}
	if rec := db.Get(prog, fid2, []Interned{arg}); rec == nil || !rec.result.Eq(r2) {
		t.Fatalf("expected second record to remain")
	}
}
