# Codebase Structure

**Analysis Date:** 2026-05-05

## Directory Layout

```
ec2-instance-docker-dev-pulumi/
├── main.go                          # Pulumi entrypoint — orchestrates all provisioning
├── go.mod                           # Go module definition (module: ec2-instance-docker-dev)
├── go.sum                           # Dependency lockfile
├── Readme.md                        # Project documentation and config reference
├── LICENSE                          # MIT license
├── config/                          # Shared Pulumi config reader functions
│   ├── GetAvailabilityZone.go       # Reads aws:region config key
│   └── GetEBSVolumeSize.go          # Reads ebs:volumeSize config key
├── ebs/                             # EBS volume resource management
│   ├── CreateVolume.go              # Creates new gp3 EBS volume; defines shared tag constants
│   └── SearchVolume.go              # Searches existing volumes by tags; defines SearchVolumeOutput
├── ec2/                             # EC2 instance and related resources
│   ├── CreateInstance.go            # Top-level EC2 instance factory (calls sub-packages)
│   ├── CreateVolumeAttachment.go    # Attaches EBS volume to instance at /dev/sdh
│   ├── getAmi.go                    # Reads ec2:ami config
│   ├── getInstanceType.go           # Reads ec2:instanceType config
│   ├── getKeyPairName.go            # Reads ec2:keyPairName config
│   ├── getSecurityGroup.go          # Creates security group + ingress/egress rules
│   ├── getSecurityGroupCidrIpv4.go  # Reads ec2:securityGroupCidrIpv4 config
│   ├── docker/                      # Docker registry authentication config
│   │   └── GetRegistryAuthentication.go  # Reads ec2:dockerRegistry config object
│   ├── elasticIP/                   # Elastic IP resource management
│   │   ├── Create.go                # Allocates EIP resource
│   │   └── CreateEipAssociation.go  # Associates EIP with EC2 instance
│   ├── securitygroup/               # Security group supporting utilities
│   │   └── GetMappingPorts.go       # Reads ec2:ingressSecurityGroups config array
│   └── userdata/                    # EC2 cloud-init script generation
│       ├── GetInstanceUserData.go   # Assembles full bash user-data script
│       ├── getDockerVersion.go      # Reads ec2:dockerVersion config
│       └── getDockerComposeVersion.go # Reads ec2:dockerComposeVersion config
└── .planning/
    └── codebase/                    # GSD codebase analysis documents
```

## Directory Purposes

**`config/`:**
- Purpose: Shared Pulumi configuration readers that are used by multiple resource packages
- Contains: Functions that wrap `pulumi/config` calls for cross-cutting config values
- Key files: `config/GetAvailabilityZone.go`, `config/GetEBSVolumeSize.go`

**`ebs/`:**
- Purpose: All EBS volume lifecycle management — tag-based discovery and creation
- Contains: `SearchVolumeOutput` type, tag constants (`pulumiProjectTag`, `pulumiStackTag`, `pulumiEBSName`), volume search and create functions
- Key files: `ebs/CreateVolume.go`, `ebs/SearchVolume.go`

**`ec2/`:**
- Purpose: EC2 instance provisioning and all directly attached resources (EIP, security group, volume attachment)
- Contains: Public resource factory functions and private config reader functions
- Key files: `ec2/CreateInstance.go`, `ec2/CreateVolumeAttachment.go`

**`ec2/docker/`:**
- Purpose: Isolate Docker registry credential reading from user-data assembly
- Contains: `dockerRegistryConfig` struct and config reader
- Key files: `ec2/docker/GetRegistryAuthentication.go`

**`ec2/elasticIP/`:**
- Purpose: Elastic IP allocation and association as distinct, testable operations
- Contains: `CreateElasticIPArgs` struct, EIP and association creators
- Key files: `ec2/elasticIP/Create.go`, `ec2/elasticIP/CreateEipAssociation.go`

**`ec2/securitygroup/`:**
- Purpose: Parse port mapping configuration for security group ingress rules
- Contains: `PortMapping` struct, config reader that deserializes YAML array
- Key files: `ec2/securitygroup/GetMappingPorts.go`

**`ec2/userdata/`:**
- Purpose: Generate the EC2 cloud-init bash script that bootstraps Docker and mounts EBS
- Contains: Script template assembly, Docker and Docker Compose version readers
- Key files: `ec2/userdata/GetInstanceUserData.go`

## Key File Locations

**Entry Points:**
- `main.go`: `pulumi.Run` callback — the only file called by the Pulumi CLI

**Resource Creation:**
- `ec2/CreateInstance.go`: Full EC2 instance provisioning (EIP → security group → user data → instance → EIP association)
- `ec2/CreateVolumeAttachment.go`: EBS ↔ EC2 binding at device `/dev/sdh`
- `ebs/CreateVolume.go`: EBS volume with `RetainOnDelete(true)` and project/stack tags
- `ec2/elasticIP/Create.go`: Elastic IP allocation
- `ec2/elasticIP/CreateEipAssociation.go`: EIP → instance association

**State & Discovery:**
- `ebs/SearchVolume.go`: Tag-based volume search (the persistence mechanism); defines `SearchVolumeOutput` used throughout

**Configuration:**
- `ec2/securitygroup/GetMappingPorts.go`: Reads `ec2:ingressSecurityGroups` YAML array into `[]PortMapping`
- `ec2/docker/GetRegistryAuthentication.go`: Reads `ec2:dockerRegistry` YAML object into `dockerRegistryConfig`
- `config/GetEBSVolumeSize.go`: Reads and parses `ebs:volumeSize` string → `pulumi.Int`

**Bootstrapping Script:**
- `ec2/userdata/GetInstanceUserData.go`: The full cloud-init bash script template (EBS mount, fstab, Docker install, registry login, daemon config)

## Naming Conventions

**Files:**
- Public functions (called from outside package): `PascalCase` filename — e.g., `CreateInstance.go`, `GetMappingPorts.go`, `SearchVolume.go`
- Private functions (internal to package): `camelCase` filename — e.g., `getAmi.go`, `getInstanceType.go`, `getDockerVersion.go`
- File name matches the primary function it contains (one function per file pattern)

**Functions:**
- Public resource creators: `PascalCase` — `CreateInstance`, `CreateVolume`, `GetInstanceUserData`
- Private config readers: `camelCase` — `getAmi`, `getInstancetype`, `getDockerVersion`
- Config readers follow pattern: `get<ConfigKey>` or `Get<ConfigKey>` depending on visibility

**Packages:**
- `lowercase` single-word names matching directory name: `ebs`, `ec2`, `config`, `userdata`, `elasticIP`, `securitygroup`, `docker`
- Note: `elasticIP` uses mixed case in package declaration (`package elasticIP`)

**Types:**
- Exported structs: `PascalCase` — `SearchVolumeOutput`, `PortMapping`, `CreateElasticIPArgs`
- Unexported structs: `camelCase` — `dockerRegistryConfig`

**Constants:**
- All lowercase with `camelCase` — `pulumiProjectTag`, `pulumiStackTag`, `pulumiEBSName` (all in `ebs/CreateVolume.go`)

## Where to Add New Code

**New AWS resource (e.g., RDS, S3, Route53):**
- Create a new top-level package directory matching the service: e.g., `rds/`, `s3/`
- Public creator function in `rds/CreateInstance.go` following `func Create*(ctx *pulumi.Context, ...) (*rds.Instance, error)` signature
- Wire it in `main.go` after the EC2 instance is created

**New EC2 sub-resource (e.g., IAM role, launch template):**
- Add a new sub-package under `ec2/`: e.g., `ec2/iam/`
- Call from `ec2/CreateInstance.go` (keeping `CreateInstance` as the single EC2 coordinator)

**New config value:**
- If scoped to one package (e.g., another `ec2:` key): add a private `camelCase` config reader function in the relevant package directory, following the pattern in `ec2/getAmi.go`
- If shared across packages (e.g., a new `aws:` or `ebs:` key): add a public function in `config/`

**New ingress port rule:**
- No code change needed — add to the `ec2:ingressSecurityGroups` array in the Pulumi stack YAML (`Pulumi.dev.yaml`). Read by `ec2/securitygroup/GetMappingPorts.go`

**New user-data step:**
- Edit the bash script template in `ec2/userdata/GetInstanceUserData.go`
- If a new versioned tool is needed, add a config reader following the pattern in `ec2/userdata/getDockerVersion.go`

**Tests:**
- No test files currently exist. If added, co-locate as `<file>_test.go` in the same package directory (Go convention)

## Special Directories

**`.planning/codebase/`:**
- Purpose: GSD codebase analysis documents (STACK.md, INTEGRATIONS.md, ARCHITECTURE.md, STRUCTURE.md, etc.)
- Generated: By GSD mapping commands
- Committed: Yes (planning artifacts tracked in git)

**`go.sum`:**
- Purpose: Cryptographic lockfile for all Go module dependencies
- Generated: Automatically by `go mod tidy` / `go get`
- Committed: Yes (required for reproducible builds)

---

*Structure analysis: 2026-05-05*
