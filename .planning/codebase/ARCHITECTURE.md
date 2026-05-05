<!-- refreshed: 2026-05-05 -->
# Architecture

**Analysis Date:** 2026-05-05

## System Overview

```text
┌──────────────────────────────────────────────────────────────┐
│                        main.go                               │
│                  Pulumi entrypoint / orchestrator            │
└────────┬──────────────────────────┬──────────────────────────┘
         │                          │
         ▼                          ▼
┌─────────────────┐      ┌──────────────────────────────────────┐
│  ebs package    │      │           ec2 package                │
│  `ebs/`         │      │           `ec2/`                     │
│                 │      │                                      │
│ SearchVolume    │      │  CreateInstance                      │
│ CreateVolume    │      │    ├── elasticIP.Create              │
└─────────────────┘      │    ├── getSecurityGroup              │
                         │    │     └── securitygroup.GetMappingPorts
                         │    ├── userdata.GetInstanceUserData  │
                         │    │     └── docker.GetRegistryAuthentication
                         │    └── elasticIP.CreateEipAssociation│
                         │                                      │
                         │  CreateVolumeAttachment              │
                         └──────────────────────────────────────┘
                                       │
                                       ▼
                         ┌──────────────────────────┐
                         │     config package        │
                         │     `config/`             │
                         │  GetAvailabilityZone      │
                         │  GetEBSVolumeSize         │
                         └──────────────────────────┘
                                       │
                                       ▼
                         ┌──────────────────────────────────────┐
                         │         AWS (via pulumi-aws)          │
                         │  EC2 Instance, EBS Volume, EIP,      │
                         │  Security Group, VPC Ingress/Egress   │
                         └──────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Orchestrator | Top-level resource ordering and wiring | `main.go` |
| ebs.SearchVolume | Finds existing retained EBS volume by Pulumi project/stack tags | `ebs/SearchVolume.go` |
| ebs.CreateVolume | Creates new gp3 EBS volume with RetainOnDelete | `ebs/CreateVolume.go` |
| ec2.CreateInstance | Builds and registers the EC2 instance resource | `ec2/CreateInstance.go` |
| ec2.CreateVolumeAttachment | Attaches EBS volume to instance at `/dev/sdh` | `ec2/CreateVolumeAttachment.go` |
| elasticIP.Create | Allocates a new Elastic IP | `ec2/elasticIP/Create.go` |
| elasticIP.CreateEipAssociation | Associates allocated EIP with the instance | `ec2/elasticIP/CreateEipAssociation.go` |
| getSecurityGroup | Creates security group + all ingress/egress rules | `ec2/getSecurityGroup.go` |
| securitygroup.GetMappingPorts | Reads ingress port mappings from Pulumi config | `ec2/securitygroup/GetMappingPorts.go` |
| userdata.GetInstanceUserData | Builds the cloud-init bash script string | `ec2/userdata/GetInstanceUserData.go` |
| docker.GetRegistryAuthentication | Reads Docker registry credentials from config | `ec2/docker/GetRegistryAuthentication.go` |
| config.GetAvailabilityZone | Reads AWS region from Pulumi config | `config/GetAvailabilityZone.go` |
| config.GetEBSVolumeSize | Reads EBS volume size from Pulumi config | `config/GetEBSVolumeSize.go` |
| getAmi | Reads AMI ID from Pulumi config | `ec2/getAmi.go` |
| getInstancetype | Reads instance type from Pulumi config | `ec2/getInstanceType.go` |
| getKeyPairName | Reads SSH key pair name from Pulumi config | `ec2/getKeyPairName.go` |
| getSecurityGroupCidrIpv4 | Reads allowed CIDR block from Pulumi config | `ec2/getSecurityGroupCidrIpv4.go` |

## Pattern Overview

**Overall:** Procedural Pulumi IaC with package-per-AWS-service decomposition

**Key Characteristics:**
- `pulumi.Context` is threaded as the first argument through every function — it is the single shared dependency
- Public functions (`PascalCase`) create or return AWS resources; private functions (`camelCase`) read Pulumi config values
- Resource creation is sequential and explicit — `main.go` defines the exact order and dependency chain
- EBS volume uses `pulumi.RetainOnDelete(true)` to survive `pulumi destroy` — the core persistence mechanism
- Instance AZ is coerced to match an existing volume's AZ when one is found (volume-first AZ strategy)

## Layers

**Orchestration Layer:**
- Purpose: Sequence all infrastructure operations; wire outputs of one resource into inputs of the next
- Location: `main.go`
- Contains: `pulumi.Run` callback, top-level control flow
- Depends on: `ebs` package, `ec2` package
- Used by: Pulumi CLI

**Resource Layer (ebs):**
- Purpose: Manage EBS volume lifecycle — search for existing retained volumes, or create new ones
- Location: `ebs/`
- Contains: `SearchVolume`, `CreateVolume`, `SearchVolumeOutput` type, shared tag constants
- Depends on: `config` package, `pulumi-aws/ebs`
- Used by: `main.go`

**Resource Layer (ec2):**
- Purpose: Create all EC2-related AWS resources (instance, security group, elastic IP, volume attachment)
- Location: `ec2/`
- Contains: `CreateInstance`, `CreateVolumeAttachment`, config reader functions
- Depends on: `ebs` (for `SearchVolumeOutput` type), `ec2/elasticIP`, `ec2/userdata`, `ec2/securitygroup`, `config`
- Used by: `main.go`

**Sub-package Layer:**
- Purpose: Isolated concerns within EC2 provisioning
- Location: `ec2/elasticIP/`, `ec2/securitygroup/`, `ec2/userdata/`, `ec2/docker/`
- Contains: Elastic IP allocation/association, port mapping config, cloud-init script generation, Docker registry config
- Depends on: `pulumi/config`, `pulumi-aws`
- Used by: `ec2/CreateInstance.go`, `ec2/getSecurityGroup.go`

**Config Layer:**
- Purpose: Read and parse shared Pulumi config values not owned by a single resource package
- Location: `config/`
- Contains: `GetAvailabilityZone`, `GetEBSVolumeSize`
- Depends on: `pulumi/config`
- Used by: `ebs/CreateVolume.go`, potentially `ec2`

## Data Flow

### Primary Provisioning Path (No Existing Volume)

1. `pulumi.Run` callback begins — `main.go:11`
2. `ebs.SearchVolume(ctx)` queries AWS for EBS volumes matching project/stack tags — `ebs/SearchVolume.go:13`
3. Returns `nil` (no volume found) — `ebs/SearchVolume.go:32`
4. `ec2.CreateInstance(ctx, nil)` called — `main.go:20`
5. `elasticIP.Create(ctx, ...)` allocates EIP — `ec2/elasticIP/Create.go:12`
6. `getSecurityGroup(ctx, elasticIp.PublicIp)` creates security group + ingress rules from config — `ec2/getSecurityGroup.go:11`
7. `userdata.GetInstanceUserData(ctx)` builds cloud-init script (Docker install, EBS mount, fstab, registry login) — `ec2/userdata/GetInstanceUserData.go:10`
8. `ec2.NewInstance` registers instance resource with AZ unrestricted — `ec2/CreateInstance.go:42`
9. `elasticIP.CreateEipAssociation` binds EIP to instance — `ec2/elasticIP/CreateEipAssociation.go:8`
10. Elastic IP exported as stack output `"elastic-ip"` — `ec2/CreateInstance.go:54`
11. `ebs.CreateVolume(ctx, ec2Instance.AvailabilityZone)` creates gp3 volume in same AZ — `main.go:34`
12. `ec2.CreateVolumeAttachment` attaches volume at `/dev/sdh` — `main.go:40`

### Resumed Provisioning Path (Existing Volume Found)

1. `ebs.SearchVolume(ctx)` returns `*SearchVolumeOutput` with existing volume ID and AZ — `ebs/SearchVolume.go:65`
2. `ec2.CreateInstance(ctx, retainedVolume)` is called with non-nil volume
3. Instance `AvailabilityZone` is pinned to `volume.AvailabilityZone` — `ec2/CreateInstance.go:39`
4. All other instance creation steps identical to primary path
5. `ec2.CreateVolumeAttachment` called with `pulumi.String(retainedVolume.ID)` (plain string, not output) — `main.go:29`
6. New volume creation is skipped entirely

### Cloud-Init Execution Path (On Instance Boot)

1. Downloads `mount-ebs-volume-ec2-user-data` binary from GitHub releases
2. Mounts EBS device `/dev/sdh` → `/home/ec2-user/ebs-volume`
3. Adds UUID-based fstab entry for persistent mounting
4. Installs Docker (version from config), git, gh CLI
5. Installs Docker Compose (version from config)
6. Authenticates Docker to private registry
7. Configures Docker daemon `data-root` to `ebs-volume/docker-data`
8. Restarts Docker and sshd

**State Management:**
- Pulumi state tracks all AWS resource IDs and dependency graph
- EBS volume uses `RetainOnDelete(true)` — survives `pulumi destroy`; re-discovered on next `pulumi up` via tag search
- No application-level state; all state is in Pulumi backend + AWS

## Key Abstractions

**SearchVolumeOutput:**
- Purpose: Plain Go struct bridging the AWS EBS lookup result to EC2 provisioning (decouples `pulumi-aws` EBS types from the EC2 package)
- File: `ebs/SearchVolume.go:8-11`
- Pattern: Custom output struct — avoids leaking Pulumi async output types into the orchestrator

**pulumi.Context threading:**
- Purpose: All functions accept `*pulumi.Context` as first arg; used for logging, config reading, and resource registration
- Examples: Every function in every package
- Pattern: Dependency injection via function argument (no globals)

**Config reader functions (private):**
- Purpose: Isolate all `config.New(ctx, "ec2").Require(...)` calls into single-responsibility functions
- Examples: `ec2/getAmi.go`, `ec2/getInstanceType.go`, `ec2/getKeyPairName.go`, `ec2/getSecurityGroupCidrIpv4.go`, `ec2/userdata/getDockerVersion.go`, `ec2/userdata/getDockerComposeVersion.go`
- Pattern: Private (`camelCase`) functions, each returns a single `pulumi.String` from config

**Tag-based volume identity:**
- Purpose: Identify owned EBS volumes across Pulumi stack operations using `PulumiProject` + `PulumiStack` + `Name` tags
- Files: `ebs/CreateVolume.go:32-36`, `ebs/SearchVolume.go:16-56`
- Constants defined in: `ebs/CreateVolume.go:10-12`

## Entry Points

**Pulumi Program Entry:**
- Location: `main.go:10`
- Triggers: `pulumi up` / `pulumi destroy` CLI invocations
- Responsibilities: Calls `pulumi.Run`, sequences EBS search → EC2 creation → volume attachment

## Architectural Constraints

- **Threading:** Single-threaded Go program; Pulumi async outputs use `ApplyT` callbacks for deferred operations. The Pulumi engine handles parallelism externally.
- **Global state:** No module-level globals except constants in `ebs/CreateVolume.go` (`pulumiProjectTag`, `pulumiStackTag`, `pulumiEBSName`)
- **Circular imports:** None. Dependency direction is strict: `main` → `ebs`/`ec2` → `config`; `ec2` sub-packages have no back-references
- **AZ coupling:** EC2 instance AZ is determined by existing EBS volume — if a volume exists in `us-east-1a`, the instance must be created there too. Changing regions requires manual volume cleanup first.
- **Root block device hardcoded:** `VolumeSize: pulumi.Int(100)` is hardcoded in `ec2/CreateInstance.go:34` — not driven from config

## Anti-Patterns

### Security Group Ingress Created Inside `ApplyT`

**What happens:** In `ec2/getSecurityGroup.go:31-38`, a second set of ingress rules (for the instance's own public IP `/32`) is created inside a `pulumi.StringInput.ApplyT` callback.
**Why it's wrong:** Pulumi resource registration inside `ApplyT` is unreliable — Pulumi's dependency graph cannot track resources registered inside Apply callbacks, which can cause silent failures or missing resources during preview.
**Do this instead:** Compute the CIDR outside the Apply or use `pulumi.All(...).ApplyT(...)` and ensure the resource is returned/tracked properly. Alternatively, pass the EIP as a known-at-deploy-time value where possible.

### Unused Struct Field

**What happens:** `CreateElasticIPArgs` in `ec2/elasticIP/Create.go:8-10` defines `InstanceId pulumi.StringInput` but it is never used inside `Create` and the caller passes `&CreateElasticIPArgs{}` (empty).
**Why it's wrong:** Dead struct fields mislead future developers about the intended API.
**Do this instead:** Remove the unused field from `CreateElasticIPArgs` or use it in `ec2.NewEip`.

### Unused `CreateVolumeResult` Type

**What happens:** `CreateVolumeResult` struct is defined in `ebs/CreateVolume.go:15-18` but `CreateVolume` returns `*ebs.Volume`, not `*CreateVolumeResult`.
**Why it's wrong:** Stale type definition with a TODO comment indicates incomplete refactoring.
**Do this instead:** Remove `CreateVolumeResult` or complete the refactor to return it.

## Error Handling

**Strategy:** Explicit `(result, error)` return pairs — idiomatic Go. Errors propagate upward immediately; no retry logic.

**Patterns:**
- Every resource creation checks `err != nil` and returns immediately: `if err != nil { return nil, err }`
- `main.go` adds contextual log messages before returning errors: `ctx.Log.Error("Error creating EC2 instance", nil)`
- Pulumi config reads use `.Require(...)` — panics (hard stop) if a required config key is missing, rather than returning an error

## Cross-Cutting Concerns

**Logging:** `ctx.Log.Info(...)` and `ctx.Log.Error(...)` — Pulumi's built-in structured logger. Used at key lifecycle points (volume found/not found, attachment start). No external logging library.
**Validation:** Done implicitly via `config.Require(...)` — missing config causes immediate program termination. No custom validation functions.
**Authentication:** Docker registry credentials read from Pulumi config (encrypted secrets via `secure:` in stack YAML). AWS credentials assumed present in environment (standard Pulumi/AWS pattern).

---

*Architecture analysis: 2026-05-05*
