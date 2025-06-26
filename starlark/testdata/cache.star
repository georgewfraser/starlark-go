# Test of caching of function calls.

load("assert.star", "assert")

mutable = [0]
def counter():
    mutable[0] += 1
    return mutable[0]

assert.eq(counter(), 1)
assert.eq(counter(), 1) # This will be wrong once I tackle mutables.