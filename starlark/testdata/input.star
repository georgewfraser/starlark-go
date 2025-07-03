load("assert.star", "assert")

i = input("x", 1)
assert.eq(i.value, 1)
i.value = 2
assert.eq(i.value, 2)

