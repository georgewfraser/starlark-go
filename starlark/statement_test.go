package starlark

import "testing"

func TestThreadStatementCounter(t *testing.T) {
	thread := new(Thread)
	const src = `
a = 1
b = 2
c = a + b
`
	if _, err := ExecFile(thread, "test.star", src, nil); err != nil {
		t.Fatalf("ExecFile: %v", err)
	}
	if thread.stmt != 2 {
		t.Fatalf("stmt=%d, want 2", thread.stmt)
	}
}
