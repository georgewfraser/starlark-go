package starlark

import "testing"

// TestProgramStateDB verifies basic operations of programStateDB.
func TestProgramStateDB(t *testing.T) {
	const numGlobals = 2
	const numStatements = 3

	db := newProgramStateDB(numGlobals, numStatements)

	// statement 0
	db.reset(0)
	db.put(0, 0, String("foo"))

	if got := db.get(0, 0); got != String("foo") {
		t.Fatalf("get g0 stmt0 = %v, want foo", got)
	}
	if val := db.get(1, 0); val != nil {
		t.Fatalf("get g1 stmt0 = %v, want nil", val)
	}

	// statement 1
	db.reset(1)
	db.put(1, 1, String("bar"))

	if got := db.get(0, 1); got != String("foo") {
		t.Fatalf("get g0 stmt1 = %v, want foo", got)
	}
	if got := db.get(1, 1); got != String("bar") {
		t.Fatalf("get g1 stmt1 = %v, want bar", got)
	}

	// statement 2
	db.reset(2)
	db.put(0, 2, String("baz"))

	if got := db.get(0, 2); got != String("baz") {
		t.Fatalf("get g0 stmt2 = %v, want baz", got)
	}
	if got := db.get(1, 2); got != String("bar") {
		t.Fatalf("get g1 stmt2 = %v, want bar", got)
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
