# |=
# option:globalreassign
load("assert.star", "assert")

mutable = {}

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 0)
mutable |= {"a": 1}
assert.eq(mutable_len(), 1)

---
# iterator
load("assert.star", "assert")

mutable = {}

def mutable_len():
    count = 0
    for _ in mutable:
        count += 1
    return count

assert.eq(mutable_len(), 0)
mutable["a"] = 1
assert.eq(mutable_len(), 1)

---
# popitem
load("assert.star", "assert")

mutable = {"a": 1, "b": 2}

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
_ = mutable.popitem()
assert.eq(mutable_len(), 1)

---
# clear
load("assert.star", "assert")

mutable = {"a": 1, "b": 2}

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
mutable.clear()
assert.eq(mutable_len(), 0)

---
# pop
load("assert.star", "assert")

mutable = {"a": 1, "b": 2}

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 2)
mutable.pop("a")
assert.eq(mutable_len(), 1)

---
# get
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_get(k):
    return mutable[k]

assert.eq(mutable_get("a"), 1)
mutable["a"] = 2
assert.eq(mutable_get("a"), 2)

---
# items
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_len():
    count = 0
    for _ in mutable.items():
        count += 1
    return count

assert.eq(mutable_len(), 1)
mutable["b"] = 2
assert.eq(mutable_len(), 2)

---
# keys
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_len():
    count = 0
    for _ in mutable.keys():
        count += 1
    return count

assert.eq(mutable_len(), 1)
mutable["b"] = 2
assert.eq(mutable_len(), 2)

---
# len
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 1)
mutable["b"] = 2
assert.eq(mutable_len(), 2)

---
# iterate
load("assert.star", "assert")

mutable = {True: 1}

def mutable_all():
    return all(mutable)

assert.eq(mutable_all(), True)
mutable[False] = 2
assert.eq(mutable_all(), False)

---
# str
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_str():
    return str(mutable)

assert.eq(mutable_str(), '{"a": 1}')
mutable["b"] = 2
assert.eq(mutable_str(), '{"a": 1, "b": 2}')

---
# freeze
load("assert.star", "assert", "freeze")

mutable = {"a": 1}

def try_to_set():
    mutable["a"] = 2

try_to_set()
freeze(mutable)
assert.fails(try_to_set, "cannot insert into frozen hash table")

---
# truth
load("assert.star", "assert")

mutable = {}

def mutable_truth():
    return bool(mutable)

assert.eq(mutable_truth(), False)
mutable["a"] = 1
assert.eq(mutable_truth(), True)

---
# union
load("assert.star", "assert")

mutable = {"a": 1}

def mutable_len_with_union():
    return len(mutable | {})

assert.eq(mutable_len_with_union(), 1)
mutable ["b"] = 2
assert.eq(mutable_len_with_union(), 2)

---
load("assert.star", "assert")

s = sneaky()
mutable = {"a": 1}

def ensure_a():
    mutable.setdefault("a", 1)
    return s()

assert.eq(ensure_a(), 1)
assert.eq(ensure_a(), 1)  # cache holds because setdefault made no change
mutable["b"] = 2
assert.eq(ensure_a(), 2)  # modification busts the cache
