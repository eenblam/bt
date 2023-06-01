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
			//result := parseInt(c.Input)
			result := ParseInt(c.Input)
			if c.WantError {
				if result.Error == nil {
					t.Log(result.Value.Int())
					t.Fatal("Wanted error, got nil")
				}
				return
			}
			if result.Error != nil {
				t.Fatalf("Unexpected error: %s", result.Error)
			}
			got := result.Value.Int()
			if got != c.Want {
				t.Fatalf("Got %v, want %v", got, c.Want)
			}
			if !bytes.Equal(result.Rest, c.WantRest) {
				t.Fatalf("Got %v, want %v", result.Rest, c.WantRest)
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
			result := ParseInteger(c.Input)
			if c.WantError {
				if result.Error == nil {
					t.Log(result.Value.Int())
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if result.Error != nil {
				t.Log(string(c.Input))
				t.Fatalf("unexpected error: %s", result.Error)
			}
			got := result.Value.Int()
			if got != c.Want {
				t.Fatalf("result.Value: Got %v, want %v", got, c.Want)
			}
			if !bytes.Equal(result.Rest, c.WantRest) {
				t.Fatalf("result:Rest: Got %v, want %v", string(result.Rest), string(c.WantRest))
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
			result := ParseString(c.Input)
			if c.WantError {
				if result.Error == nil {
					t.Log(result.Value.Int())
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if result.Error != nil {
				t.Log(string(c.Input))
				t.Fatalf("unexpected error: %s", result.Error)
			}
			got := result.Value.String()
			if !bytes.Equal(got, c.Want) {
				t.Fatalf("result.Value: Got '%v', want '%v'", string(got), string(c.Want))
			}
			if !bytes.Equal(result.Rest, c.WantRest) {
				t.Fatalf("result:Rest: Got '%v', want '%v'", string(result.Rest), string(c.WantRest))
			}
		})
	}
}
