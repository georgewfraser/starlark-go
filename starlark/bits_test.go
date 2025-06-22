package starlark

import "testing"

func TestBitsSetClearReset(t *testing.T) {
	b := NewBits(10)
	for i := 0; i < 10; i++ {
		if b.Get(i) {
			t.Fatalf("bit %d not zero", i)
		}
	}

	b.Set(3)
	b.Set(9)
	if !b.Get(3) || !b.Get(9) {
		t.Fatalf("set bits not reported")
	}
	b.Clear(3)
	if b.Get(3) {
		t.Fatalf("bit not cleared")
	}

	b.Reset()
	for i := 0; i < 10; i++ {
		if b.Get(i) {
			t.Fatalf("reset failed; bit %d not zero", i)
		}
	}
}
