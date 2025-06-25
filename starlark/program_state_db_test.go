package starlark

import (
	"slices"
	"testing"
)

func TestInternPointer(t *testing.T) {
	db := NewProgramStateDB()
	i1 := db.Intern(MakeInt(1))
	i2 := db.Intern(MakeInt(1))
	if i1 == i2 {
		t.Fatalf("expected distinct pointers")
	}
	if i1 == nil || i2 == nil {
		t.Fatalf("nil pointer returned")
	}
}

func TestProgramStateDBPutGet(t *testing.T) {
	db := NewProgramStateDB()
	arg := db.Intern(MakeInt(42))
	res := db.Intern(String("result"))
	reads := []Read{{variable: 1, value: arg}}
	db.Put(1, []Interned{arg}, reads, res)

	rec := db.Get(1, []Interned{arg})
	if rec.result != res || rec.function != 1 || !slices.Equal(rec.args, []Interned{arg}) {
		t.Fatalf("record mismatch")
	}
	if len(rec.reads) != 1 || rec.reads[0].variable != 1 || rec.reads[0].value != arg {
		t.Fatalf("reads mismatch")
	}

	// request with different arg should miss
	miss := db.Get(1, []Interned{db.Intern(MakeInt(43))})
	if miss != nil {
		t.Fatalf("expected cache miss")
	}
}

func TestProgramStateDBCollision(t *testing.T) {
	db := NewProgramStateDB()
	arg := db.Intern(MakeInt(1))
	r1 := db.Intern(String("one"))
	db.Put(0, []Interned{arg}, nil, r1)

	// find a second function id that hashes to the same slot
	target := memoHash(0, []Interned{arg})
	fid2 := 1
	for ; fid2 < CACHE_SIZE*10; fid2++ {
		if memoHash(fid2, []Interned{arg}) == target {
			break
		}
	}
	if memoHash(fid2, []Interned{arg}) != target {
		t.Fatalf("unable to find collision")
	}
	r2 := db.Intern(String("two"))
	db.Put(fid2, []Interned{arg}, nil, r2)

	// first record should be evicted
	if rec := db.Get(0, []Interned{arg}); rec != nil {
		t.Fatalf("expected eviction of first record")
	}
	if rec := db.Get(fid2, []Interned{arg}); rec == nil || rec.result != r2 {
		t.Fatalf("expected second record to remain")
	}
}
