Code to support incremental compilation. Our goal is to efficiently memoize all function calls. 

We assume that our program depends on some inputs. 

Our basic strategy is to intercept all writes to variables and mutables and intern all derived values. Small values are inspected and interned by value; large values are assumed to be new every time they are generated. 

We memoize the execution of the body of the function rather than the call site. 

def f(args):
  body

The memo key is f, args. 

db[f, args] => ?

Memoization is complicated by two facts:

1. The function may capture variables from outer scope.
2. args or captured variables may exhibit interior mutability.

We deal with interior mutability by tracking every mutable object centrally as an anonymous quasi-variable. The variables and mutables read by a function during its execution represent data-dependent, implicit arguments to the function. 

The first time we execute a particular function body against specific args, we record all reads of variables and mutables:

def f(x):
    return x + y[0] # reads: x=?, y=?, _=?

The _ represents the mutable list. It has a fixed but meaningless identity that is chosen arbitrarily every time a new mutable is generated. After executing this body the first time we enter the memoized result in our database:

db[f, args] => result, x=?, y=?, _=?

The next time we evaluate f, we find the cached result and we compare the reads to the cache. If any don't match, the cache is considered invalid and the body is re-executed. This means that we are fundamentally relying on the implicit arguments to a function to be stable. If this assumption is not true in practice, our memoization strategy will be ineffective.