package main

import (
	"reflect"
	"testing"
)

type FooBar struct {
	Foo string
	Bar []int
	Baz map[string]string
}

var inputFooBar = []byte(`d3:Foo3:oof3:Barli1234ei-4321ei0ee3:Bazd3:cow3:moo4:spam4:eggsee`)
var expectedFooBar = FooBar{
	Foo: "oof",
	Bar: []int{1234, -4321, 0},
	Baz: map[string]string{"cow": "moo", "spam": "eggs"},
}

func TestFromBencode(t *testing.T) {
	cases := []struct {
		Name      string
		Input     []byte
		Want      FooBar
		WantError bool
	}{
		{
			Name:      "Test happy path",
			Input:     inputFooBar,
			Want:      expectedFooBar,
			WantError: false,
		},
	}
	for _, c := range cases {
		got, err := FromBencode[FooBar](c.Input)
		if c.WantError {
			if err == nil {
				t.Fatal("wanted error, got nil")
			}
			return
		}
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !reflect.DeepEqual(got, c.Want) {
			t.Fatalf("want '%#v', got '%#v'", c.Want, got)
		}
	}
}
