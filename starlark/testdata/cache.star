# Test of caching of function calls.

load("assert.star", "assert")

def a():
    return 1

x = a()
y = a()

assert.eq(x, y)