"Tests tracking of mutation of lists by the cache"

load("assert.star", "assert")

mutable = []

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 0)
mutable.append(1)
assert.eq(mutable_len(), 1)

---
load("assert.star", "assert")

mutable = []

def mutable_str():
    return str(mutable)

assert.eq(mutable_str(), "[]")
mutable.append(1)
assert.eq(mutable_str(), "[1]")

---
load("assert.star", "assert")

mutable = []

def mutable_eq():
    return mutable == []

assert.eq(mutable_eq(), True)
mutable.append(1)
assert.eq(mutable_eq(), False)

---
load("assert.star", "assert")

mutable = []

def measure_len_using_plus():
    return len(mutable + [])

assert.eq(measure_len_using_plus(), 0)
mutable.append(1)
assert.eq(measure_len_using_plus(), 1)

---
load("assert.star", "assert")

mutable = []

def measure_len_using_star_right():
    return len(mutable * 2) / 2

assert.eq(measure_len_using_star_right(), 0)
mutable.append(1)
assert.eq(measure_len_using_star_right(), 1)

---
load("assert.star", "assert")

mutable = []

def measure_len_using_star_left():
    return len(2 * mutable) / 2

assert.eq(measure_len_using_star_left(), 0)
mutable.append(1)
assert.eq(measure_len_using_star_left(), 1)

---
load("assert.star", "assert")

mutable = []

def find_one():
    return 1 in mutable

assert.eq(find_one(), False)
mutable.append(1)
assert.eq(find_one(), True)

---
load("assert.star", "assert")

mutable = []

def find_one_using_iterable():
    for i in mutable:
        if i == 1:
            return True
    return False

assert.eq(find_one_using_iterable(), False)
mutable.append(1)
assert.eq(find_one_using_iterable(), True)

---
load("assert.star", "assert")

mutable = [1]

def is_one_using_slice():
    return mutable[0:1] == [1]

assert.eq(is_one_using_slice(), True)
mutable[0] = 2
assert.eq(is_one_using_slice(), False)

---
load("assert.star", "assert")

mutable = [1]

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 1)
mutable.clear()
assert.eq(mutable_len(), 0)

---
load("assert.star", "assert")

mutable = [1]

def try_to_set():
    mutable[0] = 1

try_to_set()
mutable.freeze()
assert.fails(try_to_set, "cannot assign to element of frozen list")