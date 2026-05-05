# Coding Conventions

**Analysis Date:** 2026-05-05

## Naming Patterns

### Files
- **One primary function per file** — file name mirrors the function it contains.
- **Exported functions → PascalCase filename**: `CreateInstance.go`, `CreateVolume.go`, `SearchVolume.go`, `GetRegistryAuthentication.go`, `GetMappingPorts.go`, `GetInstanceUserData.go`, `Create.go`, `CreateEipAssociation.go`
- **Unexported functions → camelCase filename**: `getAmi.go`, `getInstanceType.go`, `getKeyPairName.go`, `getSecurityGroup.go`, `getSecurityGroupCidrIpv4.go`, `getDockerComposeVersion.go`, `getDockerVersion.go`
- The file name is the single source of truth for what a file contains. Do not place multiple exported functions in one file.

### Functions
- **Exported (cross-package):** PascalCase — `CreateInstance`, `CreateVolume`, `SearchVolume`, `GetRegistryAuthentication`, `GetMappingPorts`, `GetInstanceUserData`, `GetEBSVolumeSize`
- **Unexported (package-internal):** camelCase — `getAmi`, `getKeyPairName`, `getSecurityGroup`, `createSecurityGroupIngresses`, `getPassword`
- **Known inconsistency:** `getInstancetype` (`ec2/getInstanceType.go`) uses all-lowercase `type` — should be `getInstanceType`.

### Types / Structs
- **Args structs** (function input containers): `XxxArgs` suffix — `CreateElasticIPArgs` (`ec2/elasticIP/Create.go`)
- **Output structs** (function return containers): `XxxOutput` suffix — `SearchVolumeOutput` (`ebs/SearchVolume.go`)
- **Exported structs:** PascalCase — `PortMapping`, `SearchVolumeOutput`, `CreateElasticIPArgs`
- **Unexported structs:** camelCase — `dockerRegistryConfig` (`ec2/docker/GetRegistryAuthentication.go`)

### Constants
- Unexported, camelCase — `pulumiProjectTag`, `pulumiStackTag`, `pulumiEBSName` (`ebs/CreateVolume.go`)
- Declared at package level in the file where first needed; shared across files within the same package.

### Packages
- Lowercase single-word: `ec2`, `ebs`, `config`, `docker`, `securitygroup`, `userdata`
- **Known non-standard:** `package elasticIP` (`ec2/elasticIP/Create.go`, `ec2/elasticIP/CreateEipAssociation.go`) uses camelCase — this violates Go conventions (packages should be lowercase).

## Code Style

### Formatting
- Standard `gofmt` formatting assumed (no `.prettierrc` or `biome.json`; this is a Go project).
- No explicit linter config (`.golangci.yml`) detected. Use `go vet` and `golangci-lint` if adding CI.

### Blank Lines
- Blank line after `package` declaration before imports.
- Blank line separating function body logical sections (e.g., after error checks).
- Functions separated by a single blank line.

## Import Organization

### Order
Imports are grouped but not always strictly ordered. Observed patterns:

**In files with both local and external imports:**
```go
import (
    "ec2-instance-docker-dev/ec2/docker"   // local module
    "fmt"                                   // stdlib

    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"  // third-party
)
```

**In files with only external imports:**
```go
import (
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)
```

**Prescribed order** (follow `goimports` convention):
1. Standard library (`fmt`, `strconv`)
2. Third-party packages (`github.com/pulumi/...`)
3. Local module imports (`ec2-instance-docker-dev/...`)

### Path Aliases
- None. Local imports use full module path: `ec2-instance-docker-dev/ec2/elasticIP`.
- Module name defined in `go.mod`: `module ec2-instance-docker-dev`

## Error Handling

### Pattern: Immediate Return with Nil Guard
Every fallible call is checked immediately. The canonical pattern is:

```go
resource, err := SomeCall(ctx, args)
if err != nil {
    return nil, err
}
```

### Logging Before Returning Errors
For top-level / orchestration functions, log before returning:
```go
// main.go
retainedVolume, err := ebs.SearchVolume(ctx)
if err != nil {
    ctx.Log.Error("Error searching for existing volume", nil)
    return err
}
```

Leaf/helper functions return the error directly without logging:
```go
// ebs/CreateVolume.go
if err != nil {
    return nil, err
}
```

### Error in Side-Effect Functions
Functions that only perform side effects (no meaningful return value) return `error` only:
```go
func CreateEipAssociation(ctx *pulumi.Context, ...) error {
    _, err := ec2.NewEipAssociation(...)
    return err
}
```

### Pulumi `.ApplyT()` Errors
Errors inside `.ApplyT()` callbacks are returned from the callback but **not propagated back** to the outer function. This is an existing pattern (not a convention to follow):
```go
// ec2/getSecurityGroup.go — error from ApplyT is silently lost
publicInstanceIp.ToStringOutput().ApplyT(func(ip string) error {
    err = createSecurityGroupIngresses(...)
    return err
})
```
**Do not use this pattern for critical error handling.** Prefer chaining Pulumi outputs properly.

## Logging

**Framework:** Pulumi context logger (`ctx.Log`)

```go
ctx.Log.Info("Searching for existing volume", nil)   // informational
ctx.Log.Error("Error searching for volume", nil)     // errors
```

- Second argument is always `nil` (no resource URN attached).
- `fmt.Println` / `fmt.Sprintf` are **not** used for logging — only for string construction in user data.
- Log at the start of significant operations and before returning errors.

## Function Design

### Signature Conventions
- **Resource-creating functions:** `func Name(ctx *pulumi.Context, ...) (*AwsResourceType, error)`
  - Examples: `CreateInstance`, `CreateVolume`, `Create` (elasticIP)
- **Side-effect-only functions:** `func Name(ctx *pulumi.Context, ...) error`
  - Examples: `CreateEipAssociation`, `createSecurityGroupIngresses`
- **Config reader functions (simple):** `func getName(ctx *pulumi.Context) pulumi.String`
  - Examples: `getAmi`, `getKeyPairName`, `getInstancetype`, `getDockerVersion`
- **Config reader functions (with parsing):** `func GetName(ctx *pulumi.Context) (pulumi.Int, error)`
  - Examples: `GetEBSVolumeSize`

### Size
- Files are small (8–89 lines). Functions are short — typically 10–30 lines.
- `ec2/getSecurityGroup.go` (88 lines) is the largest, containing two functions.

### Parameters: *Args Structs
When a function may grow its parameter surface, wrap inputs in an `*XxxArgs` struct:
```go
type CreateElasticIPArgs struct {
    InstanceId pulumi.StringInput
}

func Create(ctx *pulumi.Context, args *CreateElasticIPArgs) (*ec2.Eip, error) { ... }
```

## Pulumi-Specific Patterns

### Wrapping Native Values
Always wrap Go primitives in Pulumi types before passing to resource args:
```go
pulumi.String("value")        // string → pulumi.StringInput
pulumi.Int(100)               // int → pulumi.IntInput
pulumi.StringArray{...}       // []string → pulumi.StringArrayInput
pulumi.StringMap{"key": ...}  // map[string]string → pulumi.StringMapInput
```

### Config Access
Use `config.New(ctx, "namespace").Require("key")` for required values:
```go
ami := config.New(ctx, "ec2").Require("ami")
```

Use `config.New(ctx, "namespace").RequireObject("key", &target)` for structured config:
```go
pulumiConfig.RequireObject("dockerRegistry", &registryConfig)
```

Namespaces used: `"ec2"`, `"ebs"`, `"aws"`

### Stack Exports
Use `ctx.Export("key", value)` for outputs visible after `pulumi up`:
```go
ctx.Export("elastic-ip", elasticIp.PublicIp)  // ec2/CreateInstance.go
```

### Resource Options
Pass resource options as variadic last argument:
```go
ebs.NewVolume(ctx, pulumiEBSName, &ebs.VolumeArgs{...}, pulumi.RetainOnDelete(true))
```

## Comments

### When to Comment
- Inline `// todo` for incomplete implementations (lowercase, informal):
  ```go
  // todo, return idString when already exists, return ID when creted
  // ebs/CreateVolume.go:14
  ```
- Inline bash comments in user data string for script clarity.
- No GoDoc-style `//` function comments observed — not currently used.

### GoDoc
- Not used. No `// FunctionName ...` doc comments on exported functions.
- Add GoDoc when exposing packages for external consumption.

## Module Design

### Package-per-Subdirectory
Each subdirectory is its own Go package:
- `ebs/` → `package ebs`
- `ec2/` → `package ec2`
- `ec2/docker/` → `package docker`
- `ec2/elasticIP/` → `package elasticIP` *(non-standard casing)*
- `ec2/securitygroup/` → `package securitygroup`
- `ec2/userdata/` → `package userdata`
- `config/` → `package config`

### Exports
- Anything cross-package needs is exported (PascalCase).
- Package-internal helpers are unexported (camelCase).
- Constants shared across files within the same package are exported only if needed externally — currently all package constants are unexported.

### No Barrel Files
Go does not use barrel/index files. Import the specific package directly.

---

*Convention analysis: 2026-05-05*
