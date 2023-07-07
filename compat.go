// go-ungrammar: Go version compatibility code.
// This file should be replaced by the standard library once Go 1.21 is out.

// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package ungrammar

// slicesEqual reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func slicesEqual[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func intMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}
