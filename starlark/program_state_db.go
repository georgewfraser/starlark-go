package starlark

import (
	"encoding/binary"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

// CACHE_SIZE is calculated so the empty memo array is 1 MB in size.
const CACHE_SIZE = (1 << 20) / (4 * 8) // 1 MB / (size of Record)

type ProgramStateDB struct {
	// inputs provides values for the input() builtin during execution.
	inputs StringDict
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
	inputs  []InputValue
	globals []VariableValue
	cells   []CellValue
	lists   []ListVersion
	dicts   []DictVersion
	sets    []SetVersion
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

// InputValue records the value observed for an input during execution.
// Inputs are looked up by name in the Thread.inputs dictionary.
type InputValue struct {
	name  string
	value Interned
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

// DictVersion records the version of a dict observed during execution.
type DictVersion struct {
	value    *Dict
	modified uint64
}

// SetVersion records the version of a set observed during execution.
type SetVersion struct {
	value    *Set
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

// Input provides access to an input value in the ProgramStateDB.
// It records dependencies on reads and writes of the input.
type Input struct {
	name  string
	owner *Thread
}

func (db *ProgramStateDB) Input(thread *Thread, name string, def Value) *Input {
	if db.inputs == nil {
		db.inputs = make(StringDict)
	}
	if _, ok := db.inputs[name]; !ok && def != None {
		db.inputs[name] = def
	}
	return &Input{name: name, owner: thread}
}

// Value returns the current value of the input, applying the default if needed.
func (in *Input) Value() Value {
	db := &in.owner.cache
	v, ok := db.inputs[in.name]
	if !ok {
		return None
	}
	in.record(db, v)
	return v
}

// Update sets the value of the input and records the dependency.
func (in *Input) Update(val Value) {
	db := &in.owner.cache
	db.version++
	db.inputs[in.name] = val
	in.record(db, val)
}

func (in *Input) record(db *ProgramStateDB, val Value) {
	in.owner.dependencies.inputs = append(in.owner.dependencies.inputs, InputValue{
		name:  in.name,
		value: db.Intern(val),
	})
}

func (db *ProgramStateDB) Get(function *Function, args []Interned) *Record {
	idx := hashKey(function, args)
	for i := 0; i < CACHE_SIZE; i++ {
		rec := &db.memo[(idx+i)%CACHE_SIZE]
		// empty slot => miss
		if rec.result.Empty() {
			return nil
		}
		if fnEqual(rec.function, function) && argsEqual(rec.args, args) {
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
		if rec.result.Empty() || (fnEqual(rec.function, function) && argsEqual(rec.args, args)) {
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
	// inputs
	for _, inp := range rec.deps.inputs {
		if v, ok := db.inputs[inp.name]; !ok || !db.Intern(v).Eq(inp.value) {
			rec.verified = 0
			return false
		}
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
	// sets
	for _, m := range rec.deps.sets {
		if m.modified < m.value.modified {
			rec.verified = 0
			return false
		}
	}
	// dicts
	for _, m := range rec.deps.dicts {
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

// hashKey computes a hash key for the given function and arguments.
func hashKey(function *Function, args []Interned) int {
	var buf [8]byte
	h := xxhash.New()
	// Hash function based on *Funcode and freevars.
	binary.LittleEndian.PutUint64(buf[:], uint64(uintptr(unsafe.Pointer(function.funcode))))
	_, _ = h.Write(buf[:])
	for _, v := range function.freevars {
		words := Interned{value: v}.words()
		binary.LittleEndian.PutUint64(buf[:], uint64(words[0]))
		_, _ = h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(words[1]))
		_, _ = h.Write(buf[:])
	}
	// Hash the args.
	for _, a := range args {
		words := a.words()
		binary.LittleEndian.PutUint64(buf[:], uint64(words[0]))
		_, _ = h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(words[1]))
		_, _ = h.Write(buf[:])
	}
	return int(h.Sum64() % uint64(CACHE_SIZE))
}

func fnEqual(a, b *Function) bool {
	// Fast path: same function pointer.
	if a == b {
		return true
	}
	// Slow path: compare funcode and freevars.
	aCode := uint64(uintptr(unsafe.Pointer(a.funcode)))
	bCode := uint64(uintptr(unsafe.Pointer(b.funcode)))
	if aCode != bCode {
		return false
	}
	if len(a.freevars) != len(b.freevars) {
		return false
	}
	for i := range a.freevars {
		aVar := Interned{value: a.freevars[i]}
		bVar := Interned{value: b.freevars[i]}
		if !aVar.Eq(bVar) {
			return false
		}
	}
	return true
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
