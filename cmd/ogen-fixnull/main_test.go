package main

import (
	"strings"
	"testing"
)

func TestFixOptDecodeNullHandling(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFixed bool
	}{
		{
			name: "fixes simple Opt decode",
			input: `func (o *OptManualVerificationResponseModel) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptManualVerificationResponseModel to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}`,
			wantCount: 1,
			wantFixed: true,
		},
		{
			name: "skips OptNil types",
			input: `func (o *OptNilString) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptNilString to nil")
	}
	if d.Next() == jx.Null {
		if err := d.Null(); err != nil {
			return err
		}
		o.Null = true
		o.Set = true
		return nil
	}
	o.Set = true
	return nil
}`,
			wantCount: 0,
			wantFixed: false,
		},
		{
			name: "skips already fixed Opt types",
			input: `func (o *OptVoiceSettingsResponseModel) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptVoiceSettingsResponseModel to nil")
	}
	if d.Next() == jx.Null {
		if err := d.Null(); err != nil {
			return err
		}
		return nil
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}`,
			wantCount: 0,
			wantFixed: false,
		},
		{
			name: "fixes multiple Opt decode methods",
			input: `func (o *OptFoo) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptFoo to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}

func (o *OptBar) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptBar to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}`,
			wantCount: 2,
			wantFixed: true,
		},
		{
			name: "fixes OptName (starts with N but not Nil)",
			input: `func (o *OptName) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptName to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}`,
			wantCount: 1,
			wantFixed: true,
		},
		{
			name: "fixes OptNumber",
			input: `func (o *OptNumber) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptNumber to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}`,
			wantCount: 1,
			wantFixed: true,
		},
		{
			name: "skips OptNilInt (has different structure)",
			input: `func (o *OptNilInt) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptNilInt to nil")
	}
	if d.Next() == jx.Null {
		if err := d.Null(); err != nil {
			return err
		}
		o.Null = true
		o.Set = true
		return nil
	}
	o.Set = true
	return nil
}`,
			wantCount: 0,
			wantFixed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixed, count := FixOptDecodeNullHandling([]byte(tt.input))

			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}

			hasNullCheck := strings.Contains(string(fixed), "d.Next() == jx.Null")
			if tt.wantFixed && !hasNullCheck {
				t.Error("expected null check to be added, but it wasn't")
				t.Logf("output:\n%s", fixed)
			}

			if tt.wantFixed && count > 0 {
				// Verify the structure is correct
				if !strings.Contains(string(fixed), "if d.Next() == jx.Null {") {
					t.Error("null check not properly formatted")
				}
				if !strings.Contains(string(fixed), "if err := d.Null(); err != nil {") {
					t.Error("d.Null() call not found")
				}
			}
		})
	}
}

func TestFixOptDecodeNullHandling_PreservesOtherCode(t *testing.T) {
	input := `package api

import "errors"

// Some comment
func (o *OptFoo) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptFoo to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}

// Another function
func SomeOtherFunc() {
	// do something
}
`

	fixed, count := FixOptDecodeNullHandling([]byte(input))

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Verify other code is preserved
	if !strings.Contains(string(fixed), "package api") {
		t.Error("package declaration lost")
	}
	if !strings.Contains(string(fixed), "// Some comment") {
		t.Error("comment lost")
	}
	if !strings.Contains(string(fixed), "func SomeOtherFunc()") {
		t.Error("other function lost")
	}
}
