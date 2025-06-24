package starlark

// interned represents an interned value identifier. It is a pointer to a
// Value stored by the program state database. Equality of interned values is
// pointer equality, not deep equality of the underlying Value.
type interned *Value

// read is a pair of a global ID and the interned value read from that global.
type read struct {
	g int
	v interned
}

// global returns the variable ID portion of the read.
func (r read) global() int { return r.g }

// value returns the interned value portion of the read.
func (r read) value() interned { return r.v }

// programStateDB tracks the values of global variables after each top-level
// statement of a program. Values are interned so that repeated values use the
// same storage.
type programStateDB struct {
	// Total number of globals in the program.
	numGlobals int
	// Total number of statements in the program.
	numStatements int
	// Values of globals before statement 0 executes.
	inputs []interned
	// Mapping of (global, statement) to an interned value. The layout is
	// per-global blocks so that all versions of a variable are contiguous.
	globals []interned
	// Per-statement read set. A nil value indicates the variable was not
	// read by that statement; otherwise it stores the interned value that
	// was read from the variable.
	readset []interned
}

// newProgramStateDB returns a new programStateDB capable of storing the values
// of numGlobals globals for numStatements statements.
func newProgramStateDB(numGlobals, numStatements int) *programStateDB {
	db := &programStateDB{
		numGlobals:    numGlobals,
		numStatements: numStatements,
		inputs:        make([]interned, numGlobals),
		globals:       make([]interned, numGlobals*numStatements),
		readset:       make([]interned, numGlobals*numStatements),
	}
	return db
}

// reset clears the value slots for the given statement.
func (db *programStateDB) reset(stmt int) {
	base := stmt
	for g := 0; g < db.numGlobals; g++ {
		db.globals[g*db.numStatements+base] = nil
		db.readset[g*db.numStatements+base] = nil
	}
}

// put records the value of a global variable at the current statement.
func (db *programStateDB) put(global, stmt int, value Value) {
	v := value
	db.globals[global*db.numStatements+stmt] = interned(&v)
}

// get returns the value of the specified global at the current statement,
// searching backwards through earlier statements if necessary.
func (db *programStateDB) get(global, stmt int) interned {
	i := global*db.numStatements + stmt
	first := global * db.numStatements
	for ; i >= first; i-- {
		if id := db.globals[i]; id != nil {
			if db.readset[global*db.numStatements+stmt] == nil {
				db.readset[global*db.numStatements+stmt] = id
			}
			return id
		}
	}
	if id := db.inputs[global]; id != nil {
		if db.readset[global*db.numStatements+stmt] == nil {
			db.readset[global*db.numStatements+stmt] = id
		}
		return id
	}
	return nil
}

// last returns the last set value of the specified global variable.
func (db *programStateDB) last(global int) interned {
	return db.get(global, db.numStatements-1)
}

// reads returns the read set for the specified statement as a slice of
// (global, valueID) pairs. Only variables that were actually read are
// included in the result.
func (db *programStateDB) reads(stmt int) []read {
	var rs []read
	base := stmt
	for g := 0; g < db.numGlobals; g++ {
		id := db.readset[g*db.numStatements+base]
		if id != nil {
			rs = append(rs, read{g: g, v: id})
		}
	}
	return rs
}

// value returns the interned value for the given id.
func (db *programStateDB) value(id interned) Value {
	if id == nil {
		return nil
	}
	return *id
}

// modified reports whether any variable read by the specified statement
// has a different value in the current program state.
func (db *programStateDB) modified(stmt int) bool {
	for _, r := range db.reads(stmt) {
		if r.value() != db.get(r.global(), stmt-1) {
			return true
		}
	}
	return false
}

// input sets the initial value of a global variable before statement 0 executes.
func (db *programStateDB) input(global int, value Value) {
	v := value
	db.inputs[global] = interned(&v)
}
