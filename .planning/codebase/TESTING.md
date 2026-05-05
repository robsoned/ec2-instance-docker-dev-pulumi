# Testing Patterns

**Analysis Date:** 2026-05-05

## Test Framework

**Runner:** None — no test framework is configured.

**Test files:** Zero `*_test.go` files exist anywhere in the codebase.

**CI Pipeline:** No CI configuration detected (no `.github/workflows/`, no `.gitlab-ci.yml`, no `Makefile` with test targets).

**Coverage:** Not measured. No coverage tooling configured.

## Current Test State

This codebase has **no tests**. There are no:
- Unit tests
- Integration tests
- Pulumi stack tests
- Mock-based tests

The entire codebase (19 `.go` files, ~411 lines) is untested.

## Recommended Test Framework

For a Go 1.23 Pulumi project, use the following:

**Unit / integration tests:**
```bash
go test ./...         # Run all tests
go test -v ./...      # Verbose output
go test -cover ./...  # With coverage report
```

**Pulumi stack testing (end-to-end):**
- [`github.com/pulumi/pulumi/sdk/v3/go/pulumi`](https://www.pulumi.com/docs/using-pulumi/testing/) provides `pulumi.Context` mocking support.
- Use `pulumi.RunErr` in tests with a mocked context.

**Assertion library:** Standard `testing` package + `github.com/stretchr/testify` (already present as indirect dep via `go.mod` line: `github.com/stretchr/objx v0.2.0`).

## Test File Organization

**Convention to follow (Go standard):**
- Co-locate test files alongside source files.
- Name: `<SourceFile>_test.go`

```
ec2/
├── CreateInstance.go
├── CreateInstance_test.go      ← unit tests for CreateInstance
├── getAmi.go
├── getAmi_test.go
ebs/
├── CreateVolume.go
├── CreateVolume_test.go
├── SearchVolume.go
├── SearchVolume_test.go
config/
├── GetEBSVolumeSize.go
├── GetEBSVolumeSize_test.go
```

**Package naming in tests:**
- Use `package ec2_test` (black-box) for testing public API.
- Use `package ec2` (white-box) for testing internal behavior.

## Test Structure

**Standard Go test pattern to follow:**
```go
package ebs_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetEBSVolumeSize_ValidInput(t *testing.T) {
    // arrange
    // act
    // assert
    assert.Equal(t, expected, actual)
}

func TestGetEBSVolumeSize_InvalidInput(t *testing.T) {
    // ...
    assert.Error(t, err)
}
```

## Mocking

**Framework:** Not established. Recommended approach for Pulumi:

Pulumi provides `pulumi.NewContext` with a mock runtime for unit testing resources without deploying:

```go
// Example pattern for testing Pulumi resource functions
err := pulumi.RunErr(func(ctx *pulumi.Context) error {
    volume, err := ebs.CreateVolume(ctx, pulumi.String("us-east-1a"))
    assert.NoError(t, err)
    assert.NotNil(t, volume)
    return nil
}, pulumi.WithMocks("project", "stack", mocks))
```

**Pulumi mock interface:**
```go
type MyMocks struct{}

func (m *MyMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
    return args.Name + "_id", args.Inputs, nil
}

func (m *MyMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
    return args.Args, nil
}
```

**Config mocking:**
Config values are read via `config.New(ctx, "namespace").Require(...)` — these require a real Pulumi stack config or mock context with config values set.

## Testable Units

The following functions are highest-priority candidates for tests, listed by testability:

### Easiest to Test (pure/simple logic)

**`config/GetEBSVolumeSize.go` — `GetEBSVolumeSize`:**
- Contains `strconv.Atoi` conversion that can fail.
- Test: valid int string, non-numeric string (error path).

**`ec2/securitygroup/GetMappingPorts.go` — `GetMappingPorts`:**
- Deserializes config into `[]PortMapping` struct.
- Test: valid config, empty config, malformed config.

**`ec2/docker/GetRegistryAuthentication.go` — `GetRegistryAuthentication`:**
- Deserializes structured config into `dockerRegistryConfig`.
- Test: valid config object returned correctly.

### Medium Complexity (require Pulumi mocks)

**`ebs/SearchVolume.go` — `SearchVolume`:**
- Has multiple return paths: `nil, nil` when not found; `*SearchVolumeOutput, nil` when found; `nil, err` on error.
- All three branches need tests.

**`ebs/CreateVolume.go` — `CreateVolume`:**
- Calls `GetEBSVolumeSize` then creates resource.
- Test: successful creation, error from volume size parsing.

**`ec2/elasticIP/Create.go` — `Create`:**
- Thin wrapper around `ec2.NewEip`. Test with Pulumi mock.

**`ec2/elasticIP/CreateEipAssociation.go` — `CreateEipAssociation`:**
- Returns `error` directly from resource creation. Test with Pulumi mock.

### Hardest to Test (tightly coupled to Pulumi context / side effects)

**`ec2/CreateInstance.go` — `CreateInstance`:**
- Orchestrates multiple sub-calls. Requires full Pulumi mock setup.

**`ec2/userdata/GetInstanceUserData.go` — `GetInstanceUserData`:**
- Builds a bash script via `fmt.Sprintf`. Could test that key strings appear in output.

## Coverage Gaps

**All functionality is untested.** Priority gaps by risk:

| Gap | Files | Risk | Priority |
|-----|-------|------|----------|
| Volume search logic (3 code paths) | `ebs/SearchVolume.go` | Data loss if wrong volume attached | High |
| EBS volume size parsing | `config/GetEBSVolumeSize.go` | Panic / wrong volume size | High |
| Security group ingress loop | `ec2/getSecurityGroup.go` | Firewall misconfiguration | High |
| `.ApplyT()` error loss in `getSecurityGroup` | `ec2/getSecurityGroup.go` | Silent security group rule failures | High |
| ElasticIP creation and association | `ec2/elasticIP/Create.go`, `CreateEipAssociation.go` | Unreachable instance | Medium |
| Docker registry auth config parsing | `ec2/docker/GetRegistryAuthentication.go` | Deploy-time panic if config missing | Medium |
| User data script generation | `ec2/userdata/GetInstanceUserData.go` | Broken instance bootstrap | Medium |
| Port mapping deserialization | `ec2/securitygroup/GetMappingPorts.go` | Missing firewall rules | Medium |

## Run Commands

```bash
# Once tests are added:
go test ./...                          # Run all tests
go test -v ./...                       # Verbose output
go test -run TestXxx ./pkg/...         # Run specific test
go test -cover ./...                   # Coverage summary
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out  # HTML report
```

## Test Types

**Unit Tests:**
- Scope: individual functions in isolation with mocked Pulumi context.
- Not yet implemented.

**Integration Tests:**
- Scope: full Pulumi stack deployment against real AWS (or LocalStack).
- Not yet implemented.

**E2E Tests:**
- Framework: Not used. Pulumi provides `pulumi up --yes` as a manual verification path.

---

*Testing analysis: 2026-05-05*
