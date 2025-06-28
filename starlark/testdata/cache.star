# Test of caching of function calls.

load("assert.star", "assert")

s = sneaky()
def counter():
    return s()

assert.eq(counter(), 1)
assert.eq(counter(), 1) # Sneaky is untracked, so cache holds

---
load("assert.star", "assert")

def counter(arg):
    return arg

assert.eq(counter(1), 1)
assert.eq(counter(2), 2) # Change to argument busts the cache

---
# option:globalreassign
load("assert.star", "assert")

mutable = 1
def counter():
    return mutable

assert.eq(counter(), 1)
mutable = 2
assert.eq(counter(), 2) # change to global busts the cache

---
# option:globalreassign
load("assert.star", "assert")

s = sneaky()
buster = 0
def counter(inspect_buster):
    if inspect_buster:
        print(buster)
    return s()

assert.eq(counter(False), 1)
buster = 1
assert.eq(counter(False), 1) # buster is not in the read set

assert.eq(counter(True), 2) # cache is busted by argument change
assert.eq(counter(True), 2) # cache holds
buster = 2
assert.eq(counter(True), 3) # cache is busted by buster change, which is now in the read set

---
load("assert.star", "assert")

list = [0]
def counter():
    list[0] += 1
    return list[0]

assert.eq(counter(), 1)
assert.eq(counter(), 2) # Modification of list busts the cache

---
load("assert.star", "assert")

def outer(x):
    def inner():
        return x
    return inner

f1 = outer(1)
f2 = outer(2)

assert.eq(f1(), 1)
assert.eq(f2(), 2)

---
load("assert.star", "assert")

calls = []

def f(name):
  calls.append(name)

calls.clear()
f(1)
assert.eq(calls, [1])

calls.clear()
f(1)
f(2)
assert.eq(calls, [1, 2])

---
# option:globalreassign
load("assert.star", "assert")

counter = [0]
def f():
    def g():
        counter[0] += 1
        return counter[0]
    return g()

f1 = f()
f2 = f()

assert.eq(f1, 1)
assert.eq(f2, 2) # Modification of counter invalidates g(), which is a dependency of f(), which invalidates f.

---
load("assert.star", "assert")

mutable = []

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 0)
mutable.append(1)
assert.eq(mutable_len(), 1) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def mutable_str():
    return str(mutable)

assert.eq(mutable_str(), "[]")
mutable.append(1)
assert.eq(mutable_str(), "[1]") # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def mutable_eq():
    return mutable == []

assert.eq(mutable_eq(), True)
mutable.append(1)
assert.eq(mutable_eq(), False) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def measure_len_using_plus():
    return len(mutable + [])

assert.eq(measure_len_using_plus(), 0)
mutable.append(1)
assert.eq(measure_len_using_plus(), 1) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def measure_len_using_star_right():
    return len(mutable * 2) / 2

assert.eq(measure_len_using_star_right(), 0)
mutable.append(1)
assert.eq(measure_len_using_star_right(), 1) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def measure_len_using_star_left():
    return len(2 * mutable) / 2

assert.eq(measure_len_using_star_left(), 0)
mutable.append(1)
assert.eq(measure_len_using_star_left(), 1) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = []

def find_one():
    return 1 in mutable

assert.eq(find_one(), False)
mutable.append(1)
assert.eq(find_one(), True) # Modification of mutable busts the cache

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
assert.eq(find_one_using_iterable(), True) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = [1]

def is_one_using_slice():
    return mutable[0:1] == [1]

assert.eq(is_one_using_slice(), True)
mutable[0] = 2
assert.eq(is_one_using_slice(), False) # Modification of mutable busts the cache

---
load("assert.star", "assert")

mutable = [1]

def mutable_len():
    return len(mutable)

assert.eq(mutable_len(), 1)
mutable.clear()
assert.eq(mutable_len(), 0) # Modification of mutable busts the cache