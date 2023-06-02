package main

import (
	"testing"
)

func TestGetAnnounce(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name      string
		Input     Value
		Want      string
		WantError bool
	}{
		{
			Name: "Happy path",
			Input: BMap(map[string]Value{
				"announce": BString([]byte(`an announcement`)),
			}),
			Want:      "an announcement",
			WantError: false,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, err := GetAnnounce(c.Input)
			if c.WantError {
				if err == nil {
					t.Fatal("wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != c.Want {
				t.Fatalf("want '%s', got '%s'", c.Want, got)
			}
		})
	}
}
