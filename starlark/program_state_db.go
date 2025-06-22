package starlark

// programStateDB tracks the values of global variables after each top-level
// statement of a program. Values are interned so that repeated values use the
// same storage.
type programStateDB struct {
	// Total number of globals in the program.
	numGlobals int
	// Total number of statements in the program.
	numStatements int
	// Mapping of (global, statement) to an interned value index. The layout
	// is per-global blocks so that all versions of a variable are
	// contiguous.
	globals []int
	// Per-statement read set. A value of 0 indicates the variable was not
	// read by that statement; otherwise it stores the interned value ID that
	// was read from the variable.
	readset []int
	// Intern table of values. Index 0 is reserved to represent "use the
	// previous value" in globals and to denote "not read" in the read set.
	values []Value
}

// newProgramStateDB returns a new programStateDB capable of storing the values
// of numGlobals globals for numStatements statements.
func newProgramStateDB(numGlobals, numStatements int) *programStateDB {
	db := &programStateDB{
		numGlobals:    numGlobals,
		numStatements: numStatements,
		globals:       make([]int, numGlobals*numStatements),
		readset:       make([]int, numGlobals*numStatements),
		values:        make([]Value, 1), // values[0] unused
	}
	return db
}

// reset clears the value slots for the given statement.
func (db *programStateDB) reset(stmt int) {
	base := stmt
	for g := 0; g < db.numGlobals; g++ {
		db.globals[g*db.numStatements+base] = 0
		db.readset[g*db.numStatements+base] = 0
	}
}

// put records the value of a global variable at the current statement.
func (db *programStateDB) put(global, stmt int, value Value) {
	db.values = append(db.values, value)
	id := len(db.values) - 1
	db.globals[global*db.numStatements+stmt] = id
}

// get returns the value of the specified global at the current statement,
// searching backwards through earlier statements if necessary.
func (db *programStateDB) get(global, stmt int) Value {
	i := global*db.numStatements + stmt
	first := global * db.numStatements
	for ; i >= first; i-- {
		if id := db.globals[i]; id != 0 {
			db.readset[global*db.numStatements+stmt] = id
			return db.values[id]
		}
	}
	db.readset[global*db.numStatements+stmt] = 0
	return nil
}

// reads returns the read set for the specified statement as a slice of
// (global, valueID) pairs. Only variables that were actually read are
// included in the result.
func (db *programStateDB) reads(stmt int) [][2]int {
	var rs [][2]int
	base := stmt
	for g := 0; g < db.numGlobals; g++ {
		id := db.readset[g*db.numStatements+base]
		if id != 0 {
			rs = append(rs, [2]int{g, id})
		}
	}
	return rs
}

// value returns the interned value for the given id.
func (db *programStateDB) value(id int) Value {
	if id <= 0 || id >= len(db.values) {
		return nil
	}
	return db.values[id]
}
