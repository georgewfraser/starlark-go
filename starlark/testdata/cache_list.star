load("assert.star", "assert")

list = [0]
def counter():
    list[0] += 1
    return list[0]

assert.eq(counter(), 1)
assert.eq(counter(), 2) # Modification of list busts the cache