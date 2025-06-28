package starlark

import (
	"encoding/binary"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

// CACHE_SIZE is calculated so the empty memo array is 1 MB in size.
const CACHE_SIZE = (1 << 20) / (4 * 8) // 1 MB / (size of Record)

type ProgramStateDB struct {
	// memo stores the cached results of previous function calls.
	// It is an open-addressing hash table with linear probing. The
	// table currently has a fixed size and does not resize or evict
	// entries.
	memo [CACHE_SIZE]Record
	// version is bumped every time a mutable or captured variable is updated.
	// This allows us to invalidate the cache when the program state changes,
	// but skip validation if no changes were made globally, which is common.
	version uint64
}

// Dependencies groups the values and list versions read during execution.
// The interpreter records these slices while executing a function body
// and the cache uses them to detect invalidation when values change.
type Dependencies struct {
	globals []VariableValue
	cells   []CellValue
	lists   []ListVersion
	calls   []*Record
	effects bool // true for builtin functions that have side effects that are not captured in the dependencies.
}

// Record memoizes the result of a function call along with the values of
// free variables and captured locals observed during its execution.
// The key of the ProgramStateDB is the function identifier
// and its arguments, but the recorded captures are used to invalidate the
// cache if their values change.
type Record struct {
	function *Function
	args     []Interned
	deps     Dependencies
	result   Interned
	verified uint64 // ProgramStateDB.version when this record was last verified against dependencies, or 0 if it has been shown to be stale.
}

// VariableValue records the value observed for a variable during execution of
// a function body. A single variable may appear multiple times if it was
// read and then written with a different value.
type VariableValue struct {
	variable int
	value    Interned
}

// CellValue records the value observed for a captured free variable,
// which is stored by the runtime in a cell. The same capture can generate
// any number of cells for multiple executions of the outer function,
// so we need to handle these like mutables rather than like globals.
type CellValue struct {
	cell  *cell
	value Interned
}

// ListVersion records the version of a list observed during execution.
type ListVersion struct {
	value    *List
	modified uint64
}

// Interned is a reference to an interned value in the program state database.
type Interned struct {
	value Value
	_     [0]func() // uncomparable marker
}

func NewProgramStateDB() *ProgramStateDB {
	return &ProgramStateDB{}
}

func (db *ProgramStateDB) Intern(value Value) Interned {
	return Interned{value: value}
}

func (db *ProgramStateDB) Value(value Interned) Value {
	return value.value
}

func (db *ProgramStateDB) Get(function *Function, args []Interned) *Record {
	idx := hashKey(function, args)
	for i := 0; i < CACHE_SIZE; i++ {
		rec := &db.memo[(idx+i)%CACHE_SIZE]
		// empty slot => miss
		if rec.result.Empty() {
			return nil
		}
		if rec.function == function && argsEqual(rec.args, args) {
			return rec
		}
	}
	return nil
}

func (db *ProgramStateDB) Put(function *Function, args []Interned, deps Dependencies, result Interned, verified uint64) *Record {
	idx := hashKey(function, args)
	for i := 0; i < CACHE_SIZE; i++ {
		pos := (idx + i) % CACHE_SIZE
		rec := &db.memo[pos]
		if rec.result.Empty() || (rec.function == function && argsEqual(rec.args, args)) {
			db.memo[pos] = Record{function, args, deps, result, verified}
			return &db.memo[pos]
		}
	}
	panic("ProgramStateDB memo table full")
}

// Eq checks if two Interned values are equal by identity.
func (i Interned) Eq(j Interned) bool {
	return i.words() == j.words()
}

func (i Interned) Empty() bool {
	return i.value == nil
}

func (i Interned) words() [2]uintptr {
	return *(*[2]uintptr)(unsafe.Pointer(&i.value))
}

func argsEqual(a, b []Interned) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Eq(b[i]) {
			return false
		}
	}
	return true
}

// validate checks whether the given record is still valid under the
// current ProgramStateDB version. It recursively validates any
// dependent calls.
func (db *ProgramStateDB) validate(rec *Record) bool {
	if rec.verified == db.version {
		return true
	}
	if rec.verified == 0 {
		return false
	}
	// globals
	for _, c := range rec.deps.globals {
		if !db.Intern(rec.function.module.globals[c.variable]).Eq(c.value) {
			rec.verified = 0
			return false
		}
	}
	// cells
	for _, c := range rec.deps.cells {
		if !db.Intern(c.cell.v).Eq(c.value) {
			rec.verified = 0
			return false
		}
	}
	// lists
	for _, m := range rec.deps.lists {
		if m.modified < m.value.modified {
			rec.verified = 0
			return false
		}
	}
	// calls
	for _, call := range rec.deps.calls {
		if !db.validate(call) {
			rec.verified = 0
			return false
		}
	}
	rec.verified = db.version
	return true
}

// index computes the position within the memo table for the given key.
// hash computes the memo table index for the given key.
func hashKey(function *Function, args []Interned) int {
	var buf [8]byte
	h := xxhash.New()
	binary.LittleEndian.PutUint64(buf[:], uint64(uintptr(unsafe.Pointer(function))))
	_, _ = h.Write(buf[:])
	for _, a := range args {
		words := a.words()
		binary.LittleEndian.PutUint64(buf[:], uint64(words[0]))
		_, _ = h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(words[1]))
		_, _ = h.Write(buf[:])
	}
	return int(h.Sum64() % uint64(CACHE_SIZE))
}
