package starlark

import "testing"

// TestProgramStateDB verifies basic operations of programStateDB.
func TestProgramStateDB(t *testing.T) {
	const numGlobals = 2
	const numStatements = 3

	db := newProgramStateDB(numGlobals, numStatements)

	// statement 0
	db.reset(0)
	if rs := db.reads(0); len(rs) != 0 {
		t.Fatalf("reads stmt0 after reset = %v, want empty", rs)
	}
	db.put(0, 0, String("foo"))

	if got := db.value(db.get(0, 0)); got != String("foo") {
		t.Fatalf("get g0 stmt0 = %v, want foo", got)
	}
	if val := db.value(db.get(1, 0)); val != nil {
		t.Fatalf("get g1 stmt0 = %v, want nil", val)
	}
	if rs := db.reads(0); len(rs) != 1 || rs[0].global() != 0 || db.value(rs[0].value()) != String("foo") {
		t.Fatalf("reads stmt0 = %v, want [(0,foo)]", rs)
	}

	// statement 1
	db.reset(1)
	if rs := db.reads(1); len(rs) != 0 {
		t.Fatalf("reads stmt1 after reset = %v, want empty", rs)
	}
	db.put(1, 1, String("bar"))

	if got := db.value(db.get(0, 1)); got != String("foo") {
		t.Fatalf("get g0 stmt1 = %v, want foo", got)
	}
	if got := db.value(db.get(1, 1)); got != String("bar") {
		t.Fatalf("get g1 stmt1 = %v, want bar", got)
	}
	if rs := db.reads(1); len(rs) != 2 || db.value(rs[0].value()) != String("foo") || db.value(rs[1].value()) != String("bar") {
		t.Fatalf("reads stmt1 = %v, want two entries foo/bar", rs)
	}

	// statement 2
	db.reset(2)
	if rs := db.reads(2); len(rs) != 0 {
		t.Fatalf("reads stmt2 after reset = %v, want empty", rs)
	}
	db.put(0, 2, String("baz"))

	if got := db.value(db.get(0, 2)); got != String("baz") {
		t.Fatalf("get g0 stmt2 = %v, want baz", got)
	}
	if got := db.value(db.get(1, 2)); got != String("bar") {
		t.Fatalf("get g1 stmt2 = %v, want bar", got)
	}
	if rs := db.reads(2); len(rs) != 2 || db.value(rs[0].value()) != String("baz") || db.value(rs[1].value()) != String("bar") {
		t.Fatalf("reads stmt2 = %v, want two entries baz/bar", rs)
	}
}

// TestProgramStateDBInterning verifies that values are interned and reused.
func TestProgramStateDBInterning(t *testing.T) {
	db := newProgramStateDB(1, 3)
	db.reset(0)
	db.put(0, 0, String("x"))
	db.reset(1)
	db.put(0, 1, String("y"))
	db.reset(2)
	db.put(0, 2, String("x"))

	if len(db.values) != 4 { // nil + three puts
		t.Fatalf("got %d interned values, want 4", len(db.values))
	}
}

// TestProgramStateDBReadWriteSameGlobal verifies that the readset records the
// value from the first read of a global even when it is subsequently written
// and read again within the same statement.
func TestProgramStateDBReadWriteSameGlobal(t *testing.T) {
	db := newProgramStateDB(1, 2)

	// statement 0
	db.reset(0)
	db.put(0, 0, String("foo"))

	// statement 1
	db.reset(1)

	if got := db.value(db.get(0, 1)); got != String("foo") {
		t.Fatalf("first get = %v, want foo", got)
	}

	db.put(0, 1, String("bar"))

	if got := db.value(db.get(0, 1)); got != String("bar") {
		t.Fatalf("second get = %v, want bar", got)
	}

	rs := db.reads(1)
	if len(rs) != 1 || rs[0].global() != 0 || db.value(rs[0].value()) != String("foo") {
		t.Fatalf("reads stmt1 = %v, want [(0,foo)]", rs)
	}
}

// TestProgramStateDBWriteThenRead verifies that a read that occurs after a
// write in the same statement records the written value.
func TestProgramStateDBWriteThenRead(t *testing.T) {
	db := newProgramStateDB(1, 2)

	// statement 0
	db.reset(0)
	db.put(0, 0, String("foo"))

	// statement 1
	db.reset(1)
	db.put(0, 1, String("bar"))

	if got := db.value(db.get(0, 1)); got != String("bar") {
		t.Fatalf("get = %v, want bar", got)
	}

	rs := db.reads(1)
	if len(rs) != 1 || rs[0].global() != 0 || db.value(rs[0].value()) != String("bar") {
		t.Fatalf("reads stmt1 = %v, want [(0,bar)]", rs)
	}
}
