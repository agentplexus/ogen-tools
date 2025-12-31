# ogen-fixerror

Fixes ogen-generated code to preserve error response bodies.

## Problem

When an ogen-generated client receives an unexpected HTTP status code (e.g., 400, 403, 500), the error contains the `*http.Response` but the body is already closed by the time you try to read it.

This happens because:
1. The client code has `defer resp.Body.Close()`
2. The response decoder returns `UnexpectedStatusCodeWithResponse(resp)` without reading the body
3. When the function returns, the defer runs and closes the body
4. By the time you call `ogenerror.Parse()` or try to read the body, it's already closed

**Result:** You get the status code but lose the error message:
```
Status: 403
Body: (empty)
```

## Solution

This tool modifies the response decoders to buffer the body before returning the error.

**Before:**
```go
return res, validate.UnexpectedStatusCodeWithResponse(resp)
```

**After:**
```go
// Buffer the response body so it survives resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
resp.Body = io.NopCloser(bytes.NewReader(body))
return res, validate.UnexpectedStatusCodeWithResponse(resp)
```

**Result:** Error details are now accessible:
```
Status: 400
Message: An invalid ID has been received: 'invalid-voice-id'. Make sure to provide a correct one.
Detail: invalid_uid
```

## Installation

```bash
go install github.com/agentplexus/ogen-tools/cmd/ogen-fixerror@latest
```

## Usage

Run after ogen code generation:

```bash
# Generate code with ogen
ogen --package api --target internal/api --clean openapi.json

# Fix error body preservation
ogen-fixerror internal/api/oas_response_decoders_gen.go
```

Or use `go run` without installing:

```bash
go run github.com/agentplexus/ogen-tools/cmd/ogen-fixerror@latest internal/api/oas_response_decoders_gen.go
```

## Integration with generate.sh

```bash
#!/bin/bash
set -e

# Generate API code
ogen --package api --target internal/api --clean openapi.json

# Fix ogen bugs
go run github.com/agentplexus/ogen-tools/cmd/ogen-fixnull@latest internal/api/oas_json_gen.go
go run github.com/agentplexus/ogen-tools/cmd/ogen-fixerror@latest internal/api/oas_response_decoders_gen.go

# Verify build
go build ./...
```

## Reading Error Details

Use the [ogenerror](../../ogenerror/) package to extract error details:

```go
import "github.com/agentplexus/ogen-tools/ogenerror"

resp, err := client.SomeMethod(ctx, req)
if err != nil {
    if status := ogenerror.Parse(err); status != nil {
        fmt.Printf("Status: %d\n", status.StatusCode)
        fmt.Printf("Body: %s\n", status.Body)
    }
}
```

## Example Output

```
$ ogen-fixerror internal/api/oas_response_decoders_gen.go
Fixed 204 UnexpectedStatusCode returns in internal/api/oas_response_decoders_gen.go
```

If no fixes are needed:
```
$ ogen-fixerror internal/api/oas_response_decoders_gen.go
No UnexpectedStatusCode returns needed fixing in internal/api/oas_response_decoders_gen.go
```
