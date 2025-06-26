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