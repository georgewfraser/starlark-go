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
        do_nothing = buster
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

mutable = 1
def f():
    def g():
        return mutable
    return g()

assert.eq(f(), 1)
mutable = 2
assert.eq(f(), 2) # Modification of mutable invalidates g(), which is a dependency of f(), which invalidates f.