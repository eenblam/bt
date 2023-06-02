package main

import (
	"bytes"
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
			got := value.Int()
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
			//result := parseInt(c.Input)
			value, rest, err := ParseInteger(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Log(string(c.Input))
				t.Fatalf("unexpected error: %s", err)
			}
			got := value.Int()
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
		Want      []byte
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Parses string",
			Input:     []byte(`5:12345REST`),
			Want:      []byte(`12345`),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty string",
			Input:     []byte(`0:REST`),
			Want:      []byte(``),
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
			got := value.String()
			if !bytes.Equal(got, c.Want) {
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
		Want      Value
		WantRest  []byte
		WantError bool
	}{
		{
			Name:      "Parses integer list",
			Input:     []byte(`li1234ei4321eeREST`),
			Want:      BList([]Value{BInt(1234), BInt(4321)}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses mixed value list",
			Input:     []byte(`li1234e5:12345eREST`),
			Want:      BList([]Value{BInt(1234), BString([]byte(`12345`))}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:  "Parses nested lists",
			Input: []byte(`li1234eleli4321eeeREST`),
			Want: BList([]Value{
				BInt(1234),
				BList([]Value{}),
				BList([]Value{BInt(4321)}),
			}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty list",
			Input:     []byte(`leREST`),
			Want:      BList([]Value{}),
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
			if !value.Equal(c.Want) {
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
		Want      Value
		WantRest  []byte
		WantError bool
	}{
		{
			Name:  "Parses first example from spec (dict of string:string)",
			Input: []byte(`d3:cow3:moo4:spam4:eggseREST`),
			Want: BMap(map[string]Value{
				"cow":  BString([]byte(`moo`)),
				"spam": BString([]byte(`eggs`)),
			}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:  "Parses second example from spec (dict of string:list)",
			Input: []byte(`d4:spaml1:a1:beeREST`),
			Want: BMap(map[string]Value{
				"spam": BList([]Value{
					BString([]byte(`a`)),
					BString([]byte(`b`)),
				}),
			}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:  "Parses nested dicts",
			Input: []byte(`d4:spamd3:foo3:bareeREST`),
			Want: BMap(map[string]Value{
				"spam": BMap(map[string]Value{
					"foo": BString([]byte(`bar`)),
				}),
			}),
			WantRest:  []byte(`REST`),
			WantError: false,
		},
		{
			Name:      "Parses empty dict",
			Input:     []byte(`deREST`),
			Want:      BMap(map[string]Value{}),
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
			if !value.Equal(c.Want) {
				//TODO implement value.String()
				t.Fatal("wrong result.Value")
			}
			if !bytes.Equal(rest, c.WantRest) {
				t.Fatalf("rest: Got '%v', want '%v'", string(rest), string(c.WantRest))
			}
		})
	}
}
