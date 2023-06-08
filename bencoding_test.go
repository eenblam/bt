package bt

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseInt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     []byte
		Want      int
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Fails on leading zero",
			Input:     []byte(`0912`),
			WantError: true,
		},
		{
			Name:      "Fails on negative zero",
			Input:     []byte(`-0:`),
			WantError: true,
		},
		{
			Name:      "Fails on empty",
			WantError: true,
		},
		{
			Name:      "Parses 0",
			Input:     []byte(`0asdf`),
			Want:      0,
			WantRest:  []byte(`asdf`),
			WantError: false,
		},
		{
			Name:      "Parses int",
			Input:     []byte(`912:asdf`),
			Want:      912,
			WantRest:  []byte(`:asdf`),
			WantError: false,
		},
		{
			Name:      "Parses negative int",
			Input:     []byte(`-912:asdf`),
			Want:      -912,
			WantRest:  []byte(`:asdf`),
			WantError: false,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			value, rest, err := ParseInt(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("Wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			got := value.(int)
			if got != c.Want {
				t.Fatalf("Got %v, want %v", got, c.Want)
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("Got %v, want %v", rest, c.WantRest)
			}
		})
	}
}

func TestParseInteger(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     []byte
		Want      int
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Fails on leading zero",
			Input:     []byte(`i0912`),
			WantError: true,
		},
		{
			Name:      "Fails on negative zero",
			Input:     []byte(`i-0e:`),
			WantError: true,
		},
		{
			Name:      "Fails on empty",
			WantError: true,
		},
		{
			Name:      "Parses 0",
			Input:     []byte(`i0easdf`),
			Want:      0,
			WantRest:  []byte(`asdf`),
			WantError: false,
		},
		{
			Name:      "Parses non-zero inteer",
			Input:     []byte(`i912easdf`),
			Want:      912,
			WantRest:  []byte(`asdf`),
			WantError: false,
		},
		{
			Name:      "Parses negative int",
			Input:     []byte(`i-912e:asdf`),
			Want:      -912,
			WantRest:  []byte(`:asdf`),
			WantError: false,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			value, rest, err := ParseInteger(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := value.(int)
			if got != c.Want {
				t.Fatalf("value: Got %v, want %v", got, c.Want)
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("rest: Got %v, want %v", string(rest), string(c.WantRest))
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     []byte
		Want      string
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Parses string",
			Input:     []byte(`5:12345REST`),
			Want:      "12345",
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty string",
			Input:     []byte(`0:REST`),
			Want:      "",
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Fails negative length string",
			Input:     []byte(`-5:12345REST`),
			WantError: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			value, rest, err := ParseString(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := value.(string)
			if got != c.Want {
				t.Fatalf("value: Got '%v', want '%v'", string(got), string(c.Want))
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("rest: Got '%v', want '%v'", string(rest), string(c.WantRest))
			}
		})
	}
}

func TestParseList(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     []byte
		Want      []any
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Parses integer list",
			Input:     []byte(`li1234ei4321eeREST`),
			Want:      []any{1234, 4321},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses mixed value list",
			Input:     []byte(`li1234e5:12345eREST`),
			Want:      []any{1234, "12345"},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses nested lists",
			Input:     []byte(`li1234eleli4321eeeREST`),
			Want:      []any{1234, []any{}, []any{4321}},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty list",
			Input:     []byte(`leREST`),
			Want:      []any{},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Fails non-list",
			Input:     []byte(`i1234e`),
			WantError: true,
		},
		{
			Name:      "Fails without terminal e",
			Input:     []byte(`li1234ei4321e`),
			WantError: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			value, rest, err := ParseList(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(value, c.Want) {
				//TODO implement value.String()
				t.Fatal("wrong result.Value")
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("rest: Got '%v', want '%v'", string(rest), string(c.WantRest))
			}
		})
	}
}

func TestParseDict(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     []byte
		Want      map[string]any
		WantRest  []byte
		WantError bool
	}{
		{
			Name:  "Parses first example from spec (dict of string:string)",
			Input: []byte(`d3:cow3:moo4:spam4:eggseREST`),
			Want: map[string]any{
				"cow":  "moo",
				"spam": "eggs",
			},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:  "Parses second example from spec (dict of string:list)",
			Input: []byte(`d4:spaml1:a1:beeREST`),
			Want: map[string]any{
				"spam": []any{"a", "b"},
			},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses nested dicts",
			Input:     []byte(`d4:spamd3:foo3:bareeREST`),
			Want:      map[string]any{"spam": map[string]any{"foo": "bar"}},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty dict",
			Input:     []byte(`deREST`),
			Want:      map[string]any{},
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Fails non-dict",
			Input:     []byte(`i1234e`),
			WantError: true,
		},
		{
			Name:      "Fails without terminal e",
			Input:     []byte(`d4:spaml1:a1:be`),
			WantError: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			value, rest, err := ParseDict(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(value, c.Want) {
				//TODO implement value.String()
				t.Fatal("wrong result.Value")
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("rest: Got '%v', want '%v'", string(rest), string(c.WantRest))
			}
		})
	}
}

func TestFromBencode(t *testing.T) {
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
