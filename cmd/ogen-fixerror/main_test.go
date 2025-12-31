package main

import (
	"strings"
	"testing"
)

func TestFixUnexpectedStatusCodeBody(t *testing.T) {
	input := `package api

import (
	"github.com/ogen-go/ogen/validate"
)

func decodeTestResponse(resp *http.Response) (res TestRes, _ error) {
	switch resp.StatusCode {
	case 200:
		return &TestOK{}, nil
	}
	return res, validate.UnexpectedStatusCodeWithResponse(resp)
}
`

	expected := `// Buffer the response body so it survives resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return res, validate.UnexpectedStatusCodeWithResponse(resp)`

	fixed, count := FixUnexpectedStatusCodeBody([]byte(input))

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	if !strings.Contains(string(fixed), expected) {
		t.Errorf("output does not contain expected fix:\n%s", string(fixed))
	}

	// Check imports were added
	if !strings.Contains(string(fixed), `"bytes"`) {
		t.Error("missing bytes import")
	}
	if !strings.Contains(string(fixed), `"io"`) {
		t.Error("missing io import")
	}
}

func TestFixUnexpectedStatusCodeBody_MultipleReturns(t *testing.T) {
	input := `package api

import (
	"github.com/ogen-go/ogen/validate"
)

func decode1(resp *http.Response) (res Res1, _ error) {
	return res, validate.UnexpectedStatusCodeWithResponse(resp)
}

func decode2(resp *http.Response) (res Res2, _ error) {
	return res, validate.UnexpectedStatusCodeWithResponse(resp)
}
`

	fixed, count := FixUnexpectedStatusCodeBody([]byte(input))

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// Count occurrences of the fix
	occurrences := strings.Count(string(fixed), "Buffer the response body")
	if occurrences != 2 {
		t.Errorf("fix occurrences = %d, want 2", occurrences)
	}
}

func TestFixUnexpectedStatusCodeBody_AlreadyHasImports(t *testing.T) {
	input := `package api

import (
	"bytes"
	"io"
	"github.com/ogen-go/ogen/validate"
)

func decodeTest(resp *http.Response) (res TestRes, _ error) {
	return res, validate.UnexpectedStatusCodeWithResponse(resp)
}
`

	fixed, count := FixUnexpectedStatusCodeBody([]byte(input))

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Should not duplicate imports
	bytesCount := strings.Count(string(fixed), `"bytes"`)
	if bytesCount != 1 {
		t.Errorf("bytes import count = %d, want 1", bytesCount)
	}
}
