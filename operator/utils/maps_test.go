package utils

import (
	"reflect"
	"testing"
)

func TestMapMerge(t *testing.T) {
	s1 := make([]int, 5)
	s2 := make([]int, 5)
	s3 := make([]int, 5)
	a := map[string]any{
		"mine":   "me",
		"theirs": "me",
		"ours":   s1,
		"eek": struct {
			cluck string
			size  float64
		}{"lee", 1.75},
	}
	b := map[string]any{
		"theirs": "you",
		"extra":  2,
		"ours":   s2,
		"eek": struct {
			cluck string
			size  float64
		}{"leeroy", 1.45},
	}
	c := map[string]any{
		"ours": s3,
		"eek": struct {
			size  float64
			jenny string
		}{1.80, "jenkins"},
	}
	r := map[string]any{
		"mine":   "me",
		"theirs": "you",
		"ours":   s3,
		"eek": struct {
			cluck string
			size  float64
			jenny string
		}{"leeroy", 1.80, "jenkins"},
	}
	o := MergeMapsShallow(a, b, c)
	eq := reflect.DeepEqual(o, r)
	if eq {
		// they are equal
	} else {
		// they are unequal
		t.Logf("A: %v", a)
		t.Logf("B: %v", b)
		t.Logf("C: %v", c)
		t.Logf("E: %v", r)
		t.Logf("O: %v", o)
		t.Log("Expected and Output are unequal.")
		t.Fatal("Deep Map Merge has failed results are not equal to expected.")
	}
}
