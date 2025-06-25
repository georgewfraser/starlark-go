package incremental

import (
	"encoding/binary"
	"slices"

	"unsafe"

	"github.com/cespare/xxhash/v2"
	"go.starlark.net/starlark"
)

// CACHE_SIZE is calculated so the empty memo array is 1 MB in size.
const CACHE_SIZE = (1 << 20) / (4 * 8) // 1 MB / (size of Record)

type ProgramStateDB struct {
	// memo stores the cached results of previous function calls.
	// It is a quasi-perfect hash table where the key is a hash of the function identifier and its arguments.
	// In the event of a hash collision we evict the existing entry and replace it with the new one.
	memo [CACHE_SIZE]Record
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
	return &ProgramStateDB{}
}

func (db *ProgramStateDB) Intern(value starlark.Value) Interned {
	v := value
	return &v
}

func (db *ProgramStateDB) Get(function int, args []Interned) *Record {
	idx := hash(function, args)
	rec := &db.memo[idx]

	// If the entry is empty, return nil.
	if rec.result == nil {
		return nil
	}

	// If the entry is a collision, return nil.
	if rec.function != function || !slices.Equal(rec.args, args) {
		return nil
	}

	return rec
}

func (db *ProgramStateDB) Put(function int, args []Interned, reads []Read, value Interned) {
	idx := hash(function, args)
	db.memo[idx] = Record{function: function, args: args, reads: reads, result: value}
}

// index computes the position within the memo table for the given key.
// hash computes the memo table index for the given key.
func hash(function int, args []Interned) int {
	var buf [8]byte
	h := xxhash.New()
	binary.LittleEndian.PutUint64(buf[:], uint64(function))
	_, _ = h.Write(buf[:])
	for _, a := range args {
		binary.LittleEndian.PutUint64(buf[:], uint64(uintptr(unsafe.Pointer(a))))
		_, _ = h.Write(buf[:])
	}
	return int(h.Sum64() % uint64(CACHE_SIZE))
}
