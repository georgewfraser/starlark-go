package starlark

import (
	"encoding/binary"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"go.starlark.net/internal/compile"
)

// CACHE_SIZE is calculated so the empty memo array is 1 MB in size.
const CACHE_SIZE = (1 << 20) / (4 * 8) // 1 MB / (size of Record)

type ProgramStateDB struct {
	// memo stores the cached results of previous function calls.
	// It is a quasi-perfect hash table where the key is a hash of the
	// program pointer, function identifier, and its arguments. In the event
	// of a hash collision we evict the existing entry and replace it with
	// the new one.
	memo [CACHE_SIZE]Record
}

// Record memoizes the result of a function call along with the values of
// captures observed during its execution
// The key of the ProgramStateDB is the function identifier
// and its arguments, but the recorded captures are used to invalidate the
// cache if their values change.
type Record struct {
	program  *compile.Program
	function int
	args     []Interned
	captures []Capture
	mutables []Mutable
	result   Interned
}

// Capture records the value observed for a global or captured local during
// execution of a function body. A single global may appear multiple times in
// this slice if it was read and then written with a different value.
// Variables are numbered such that 0..numGlobals-1 are globals and
// numGlobals..numGlobals+numCaptures-1 are captured locals.
type Capture struct {
	variable int
	value    Interned
}

type Mutable struct {
	value   *List
	version int
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

func (db *ProgramStateDB) Get(program *compile.Program, function int, args []Interned) *Record {
	idx := hashKey(program, function, args)
	rec := &db.memo[idx]

	// If the entry is empty, return nil.
	if rec.result.Empty() {
		return nil
	}

	// If the entry is a collision, return nil.
	if rec.program != program || rec.function != function {
		return nil
	}
	for i := range rec.args {
		if !rec.args[i].Eq(args[i]) {
			return nil
		}
	}

	return rec
}

func (db *ProgramStateDB) Put(program *compile.Program, function int, args []Interned, captures []Capture, mutables []Mutable, result Interned) {
	idx := hashKey(program, function, args)
	db.memo[idx] = Record{program, function, args, captures, mutables, result}
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

// index computes the position within the memo table for the given key.
// hash computes the memo table index for the given key.
func hashKey(program *compile.Program, function int, args []Interned) int {
	var buf [8]byte
	h := xxhash.New()
	binary.LittleEndian.PutUint64(buf[:], uint64(uintptr(unsafe.Pointer(program))))
	_, _ = h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(function))
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
