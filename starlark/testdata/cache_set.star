# Tests tracking of mutation of sets by the cache
# option:set
load("assert.star", "assert")

mutable = set()

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 0)
mutable.add(1)
assert.eq(mutable_len(), 1)

---
# option:set
load("assert.star", "assert")

mutable = set()

def mutable_len_iter():
    count = 0
    for _ in mutable:
        count += 1
    return count

assert.eq(mutable_len_iter(), 0)
mutable.add(1)
assert.eq(mutable_len_iter(), 1)

# option:set
---
load("assert.star", "assert")

mutable = set([1, 2])

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
mutable.pop()
assert.eq(mutable_len(), 1)
# option:set

---
load("assert.star", "assert")

mutable = set([1, 2])

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
mutable.clear()
# option:set
assert.eq(mutable_len(), 0)

---
load("assert.star", "assert")

mutable = set([1, 2])

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
# option:set
mutable.discard(1)
assert.eq(mutable_len(), 1)

---
load("assert.star", "assert")

mutable = set([1, 2])

def mutable_len():
    return len(mutable)

# option:set
assert.eq(mutable_len(), 2)
mutable.remove(1)
assert.eq(mutable_len(), 1)

---
load("assert.star", "assert")

mutable = set([1])

def mutable_len():
    return len(mutable)
# option:set

assert.eq(mutable_len(), 1)
mutable.add(2)
assert.eq(mutable_len(), 2)

---
load("assert.star", "assert")

mutable = set([1])

def mutable_str():
# option:set
    return str(mutable)

assert.eq(mutable_str(), "set([1])")
mutable.add(2)
assert.eq(mutable_str(), "set([1, 2])")

---
load("assert.star", "assert")

mutable = set()

# option:set
def mutable_truth():
    return bool(mutable)

assert.eq(mutable_truth(), False)
mutable.add(1)
assert.eq(mutable_truth(), True)

---
load("assert.star", "assert")
# option:set

mutable = set([1])

def mutable_len_with_union():
    return len(mutable | set())

assert.eq(mutable_len_with_union(), 1)
mutable.add(2)
assert.eq(mutable_len_with_union(), 2)


