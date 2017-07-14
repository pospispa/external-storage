package openstackmicroversions

import (
	"reflect"
	"testing"
)

func TestInvalidMicroversions(t *testing.T) {
	tests := []string{
		"",
		"2",
		"2.0.0",
		"a.0",
		"-1.3",
		"0.b",
		"3.-3",
	}
	for _, tt := range tests {
		if got, err := New(tt); err == nil {
			t.Errorf("New(%q) = (%v, %v) want (nil, an error)", tt, got, err)
		}
	}
}

func TestValidMicroversions(t *testing.T) {
	tests := []struct {
		validInput string
		want       *Microversion
	}{
		{
			validInput: "2.0",
			want:       &Microversion{major: 2, minor: 0},
		},
		{
			validInput: "3.1",
			want:       &Microversion{major: 3, minor: 1},
		},
	}
	for _, tt := range tests {
		if got, err := New(tt.validInput); err != nil {
			t.Errorf("New(%q) = (%v, %v) want (%v, %v)", tt.validInput, got, err.Error(), tt.want, nil)
		} else {
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New(%q) = (%v, %v) want (%v, %v)", tt.validInput, got, err, tt.want, nil)
			}
		}
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		v1   *Microversion
		v2   *Microversion
		want bool
	}{
		{
			v1:   &Microversion{major: 0, minor: 1},
			v2:   &Microversion{major: 1, minor: 0},
			want: true,
		},
		{
			v1:   &Microversion{major: 2, minor: 1},
			v2:   &Microversion{major: 1, minor: 0},
			want: false,
		},
		{
			v1:   &Microversion{major: 2, minor: 1},
			v2:   &Microversion{major: 2, minor: 0},
			want: false,
		},
		{
			v1:   &Microversion{major: 2, minor: 1},
			v2:   &Microversion{major: 2, minor: 1},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := tt.v1.LessThan(tt.v2); got != tt.want {
			t.Errorf("LessThan(%v, %v) = (%v) want (%v)", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestGetMajor(t *testing.T) {
	want := uint(2)
	test, err := New("2.5")
	if err != nil {
		t.Errorf("internal error: (%v)", err.Error())
	}
	if got := test.GetMajor(); got != want {
		t.Errorf("(%v).GetMajor() = (%v) want (%v)", test, got, want)
	}
}

func TestGetMinor(t *testing.T) {
	want := uint(7)
	test, err := New("2.7")
	if err != nil {
		t.Errorf("internal error: (%v)", err.Error())
	}
	if got := test.GetMinor(); got != want {
		t.Errorf("(%v).GetMinor() = (%v) want (%v)", test, got, want)
	}
}

func TestString(t *testing.T) {
	want := "3.7"
	test, err := New(want)
	if err != nil {
		t.Errorf("internal error: (%v)", err.Error())
	}
	if got := test.String(); got != want {
		t.Errorf("(%v).String() = %q want %q", test, got, want)
	}
}
