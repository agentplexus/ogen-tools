# ogen-fixnull

Fixes ogen-generated code to handle null values in `Opt*` types.

## Problem

[ogen-go/ogen#1358](https://github.com/ogen-go/ogen/issues/1358): When an OpenAPI spec has a nullable `$ref` field like:

```json
"manual_verification": {
  "$ref": "#/components/schemas/ManualVerificationResponseModel",
  "nullable": true
}
```

ogen generates `Opt*` types instead of `OptNil*` types. The `Opt*` types don't handle `null` values, causing JSON decoding errors:

```
decode ManualVerificationResponseModel: "{" expected: unexpected byte 110 'n'
```

(byte 110 is `n` from `null`)

## Solution

This tool adds null handling to all `Opt*` Decode methods that are missing it.

**Before:**
```go
func (o *OptManualVerificationResponseModel) Decode(d *jx.Decoder) error {
    if o == nil {
        return errors.New("invalid: unable to decode OptManualVerificationResponseModel to nil")
    }
    o.Set = true
    if err := o.Value.Decode(d); err != nil {
        return err
    }
    return nil
}
```

**After:**
```go
func (o *OptManualVerificationResponseModel) Decode(d *jx.Decoder) error {
    if o == nil {
        return errors.New("invalid: unable to decode OptManualVerificationResponseModel to nil")
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
}
```

## Installation

```bash
go install github.com/agentplexus/ogen-tools/cmd/ogen-fixnull@latest
```

## Usage

Run after ogen code generation:

```bash
# Generate code with ogen
ogen --package api --target internal/api --clean openapi.json

# Fix null handling
ogen-fixnull internal/api/oas_json_gen.go
```

Or use `go run` without installing:

```bash
go run github.com/agentplexus/ogen-tools/cmd/ogen-fixnull@latest internal/api/oas_json_gen.go
```

## Integration with generate.sh

```bash
#!/bin/bash
set -e

# Check if ogen is installed
if ! command -v ogen &> /dev/null; then
    echo "Error: ogen is not installed."
    echo "Install with: go install github.com/ogen-go/ogen/cmd/ogen@latest"
    exit 1
fi

# Generate API code
echo "Generating API code with ogen..."
ogen --package api --target internal/api --clean openapi.json

# Fix ogen null handling bug (https://github.com/ogen-go/ogen/issues/1358)
echo "Fixing Opt* null handling..."
go run github.com/agentplexus/ogen-tools/cmd/ogen-fixnull@latest internal/api/oas_json_gen.go

# Verify build
echo "Verifying build..."
go build ./...

echo "Done!"
```

## How It Works

The tool uses a regex to find `Opt*` (non-`OptNil*`) Decode methods that don't already have null handling, and inserts the null check before `o.Set = true`.

It's safe to run multiple times - already-fixed methods are skipped.

## Example Output

```
$ ogen-fixnull internal/api/oas_json_gen.go
Fixed 208 Opt* Decode methods in internal/api/oas_json_gen.go
```

If no fixes are needed:
```
$ ogen-fixnull internal/api/oas_json_gen.go
No Opt* Decode methods needed fixing in internal/api/oas_json_gen.go
```
