package starlark

// interned represents an interned value identifier. Using a small unsigned
// integer keeps the backing arrays compact.
type interned uint16

// read is a packed representation of a variable ID and the interned value read
// from that variable. The high 16 bits store the variable ID and the low 16
// bits store the interned value ID.
type read uint32

// global returns the variable ID portion of the read.
func (r read) global() int { return int(uint16(uint32(r) >> 16)) }

// value returns the interned value portion of the read.
func (r read) value() interned { return interned(uint16(uint32(r))) }

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
	globals []interned
	// Per-statement read set. A value of 0 indicates the variable was not
	// read by that statement; otherwise it stores the interned value ID that
	// was read from the variable.
	readset []interned
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
		globals:       make([]interned, numGlobals*numStatements),
		readset:       make([]interned, numGlobals*numStatements),
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
	db.globals[global*db.numStatements+stmt] = interned(id)
}

// get returns the value of the specified global at the current statement,
// searching backwards through earlier statements if necessary.
func (db *programStateDB) get(global, stmt int) interned {
	i := global*db.numStatements + stmt
	first := global * db.numStatements
	for ; i >= first; i-- {
		if id := db.globals[i]; id != 0 {
			db.readset[global*db.numStatements+stmt] = id
			return id
		}
	}
	db.readset[global*db.numStatements+stmt] = 0
	return 0
}

// reads returns the read set for the specified statement as a slice of
// (global, valueID) pairs. Only variables that were actually read are
// included in the result.
func (db *programStateDB) reads(stmt int) []read {
	var rs []read
	base := stmt
	for g := 0; g < db.numGlobals; g++ {
		id := db.readset[g*db.numStatements+base]
		if id != 0 {
			rs = append(rs, read(uint32(g)<<16|uint32(id)))
		}
	}
	return rs
}

// value returns the interned value for the given id.
func (db *programStateDB) value(id interned) Value {
	idx := int(id)
	if idx <= 0 || idx >= len(db.values) {
		return nil
	}
	return db.values[idx]
}
