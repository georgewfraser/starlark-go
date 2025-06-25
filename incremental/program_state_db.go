package incremental

import "go.starlark.net/starlark"

// MAX_INTERNED is calculated so the empty memo array is 1 MB in size.
const MAX_INTERNED = (1 << 20) / (2 * 8) // 1 MB / (2 * size of Record)

type ProgramStateDB struct {
	// memo stores the cached results of previous function calls.
	// It is a quasi-perfect hash table where the key is a hash of the function identifier and its arguments.
	// In the event of a hash collision we evict the existing entry and replace it with the new one.
	memo [MAX_INTERNED]Record
}

// key represents a unique identifier for a function call in the program state.
// The arguments are part of the key but there are implicit arguments that are not.
// These implicit arguments are captured variables and mutables in the interior of explicit or implicit arguments.
// The implicit arguments are part of the record and are validated against the current state of the program.
type key struct {
	function int        // id of the function
	args     []Interned // arguments to the function, interned values
}

// Record memoizes the result of a function call along with the values read
// from variables or mutables during the execution of the function body.
// The key of the ProgramStateDB is the function identifier and the arguments
// but the reads are used to invalidate the cache if implicit arguments change.
type Record struct {
	function int
	args     []Interned
	reads    []Read
	result   Interned
}

// Read records the value read from a variable or mutable during the execution
// of a function body. It captures the variable's identifier and the value that
// was read at the time of execution.
type Read struct {
	// variable identifies the global variable, captured local variable, or mutable that was read.
	// Since we are memoizing the execution of function bodies, it only needs to be unique within the context of the named function.
	// The set of globals and captures is fixed for a given function and resolve ahead of time.
	// So 0..numGlobals is reserved for globals, numGlobals..numGlobals+numCaptures for captures, and numGlobals+numCaptures.. for mutables.
	variable int
	// value that was read on the last execution of the function body
	value Interned
}

// Interned is a reference to an interned value in the program state database.
// It is used to avoid copying large values and to ensure that the same value
// is reused across different function calls or statements.
type Interned *starlark.Value

func NewProgramStateDB() *ProgramStateDB {
	panic("todo")
}

func (db *ProgramStateDB) Intern(value starlark.Value) Interned {
	panic("todo")
}

func (db *ProgramStateDB) Get(function int, args []Interned) Record {
	panic("todo")
}

func (db *ProgramStateDB) Put(function int, args []Interned, reads []Read, value Interned) {
	panic("todo")
}
