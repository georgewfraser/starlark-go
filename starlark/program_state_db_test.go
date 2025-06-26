package starlark

import (
	"testing"
)

func TestIntern(t *testing.T) {
	db := NewProgramStateDB()

	// Small values should be interned by golang.
	i1 := db.Intern(MakeInt(dynamicInt(1)))
	i2 := db.Intern(MakeInt(dynamicInt(1)))
	if !i1.Eq(i2) {
		t.Fatalf("expected interned pointers to match for small int")
	}

	// But if they are different, they should not match.
	i3 := db.Intern(MakeInt(dynamicInt(2)))
	if i1.Eq(i3) {
		t.Fatalf("expected interned pointers to differ for different small ints")
	}

	// Identical values should be equal when I convert them back to Value.
	if db.Value(i1) != db.Value(i2) {
		t.Fatalf("expected interned values to match when converted back to Value")
	}

	// Large values should be interned as distinct values even if they are equal.
	s1 := db.Intern(String(dynamicString("hello")))
	s2 := db.Intern(String(dynamicString("hello")))
	if s1.Eq(s2) {
		t.Fatalf("expected interned pointers to differ for distinct strings")
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
	arg := db.Intern(MakeInt(dynamicInt(42)))
	res := db.Intern(String("result"))
	reads := []Read{{variable: 1, value: arg}}
	writes := []Write{{variable: 2, value: res}}
	db.Put(1, []Interned{arg}, reads, writes, res)
	rec := db.Get(1, []Interned{arg})
	if rec == nil {
		t.Fatalf("expected record to be found")
	}
	if !rec.result.Eq(res) || rec.function != 1 || len(rec.args) != 1 || !rec.args[0].Eq(arg) {
		t.Fatalf("record mismatch")
	}
	if len(rec.reads) != 1 || rec.reads[0].variable != 1 || !rec.reads[0].value.Eq(arg) {
		t.Fatalf("reads mismatch")
	}
	if len(rec.writes) != 1 || rec.writes[0].variable != 2 || !rec.writes[0].value.Eq(res) {
		t.Fatalf("writes mismatch")
	}

	// request with different arg should miss
	miss := db.Get(1, []Interned{db.Intern(MakeInt(dynamicInt(43)))})
	if miss != nil {
		t.Fatalf("expected cache miss")
	}
}

func TestProgramStateDBCollision(t *testing.T) {
	db := NewProgramStateDB()
	arg := db.Intern(MakeInt(dynamicInt(1)))
	r1 := db.Intern(String("one"))
	db.Put(0, []Interned{arg}, nil, nil, r1)

	// find a second function id that hashes to the same slot
	target := hashKey(0, []Interned{arg})
	fid2 := 1
	for ; fid2 < CACHE_SIZE*10; fid2++ {
		if hashKey(fid2, []Interned{arg}) == target {
			break
		}
	}
	if hashKey(fid2, []Interned{arg}) != target {
		t.Fatalf("unable to find collision")
	}
	r2 := db.Intern(String("two"))
	db.Put(fid2, []Interned{arg}, nil, nil, r2)

	// first record should be evicted
	if rec := db.Get(0, []Interned{arg}); rec != nil {
		t.Fatalf("expected eviction of first record")
	}
	if rec := db.Get(fid2, []Interned{arg}); rec == nil || !rec.result.Eq(r2) {
		t.Fatalf("expected second record to remain")
	}
}
