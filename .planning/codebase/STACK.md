# Technology Stack

**Analysis Date:** 2026-05-05

## Languages

**Primary:**
- Go 1.23.1 — entire codebase; all infrastructure logic written in Go

## Runtime

**Environment:**
- Go runtime 1.23.1 — required minimum version per `go.mod`

**Package Manager:**
- Go modules (`go mod`)
- Lockfile: `go.sum` present and committed

## Frameworks

**Core:**
- Pulumi SDK v3 (`github.com/pulumi/pulumi/sdk/v3 v3.158.0`) — IaC engine; all resources declared via `pulumi.Run` in `main.go`
- Pulumi AWS Provider v6 (`github.com/pulumi/pulumi-aws/sdk/v6 v6.73.0`) — AWS resource provisioning (EC2, EBS, VPC, Elastic IP)

**Build/Dev:**
- Pulumi CLI (≥ v3.0.0) — required externally; not vendored, called via shell during deployments
- Standard `go build` / `go run` — no custom build tooling

**Testing:**
- No test framework detected; no `*_test.go` files present

## Key Dependencies

**Critical:**
- `github.com/pulumi/pulumi-aws/sdk/v6 v6.73.0` — provides all AWS resource types used: `ec2.Instance`, `ec2.Eip`, `ec2.EipAssociation`, `ec2.SecurityGroup`, `ec2.VolumeAttachment`, `ebs.Volume`, `vpc.SecurityGroupIngressRule`, `vpc.SecurityGroupEgressRule`
- `github.com/pulumi/pulumi/sdk/v3 v3.158.0` — core IaC runtime, config system (`pulumi/config`), context, logging, and `pulumi.Run` entrypoint

**Infrastructure (indirect, pulled by Pulumi SDK):**
- `google.golang.org/grpc v1.71.0` — gRPC transport between Pulumi engine and Go program
- `google.golang.org/protobuf v1.36.6` — protobuf serialization for Pulumi resource RPCs
- `github.com/go-git/go-git/v5 v5.14.0` — used by Pulumi SDK internally
- `github.com/charmbracelet/bubbletea v1.3.4` — Pulumi CLI TUI components (indirect)
- `github.com/hashicorp/hcl/v2 v2.23.0` — HCL parsing used by Pulumi config system

## Configuration

**Pulumi Stack Config:**
- Configuration is read at deploy time via `config.New(ctx, "<namespace>")` from `Pulumi.<stack>.yaml` files
- Stack YAML files are **gitignored** via `.gitignore` pattern `Pulumi.*.yaml`
- Sensitive values (e.g. `ec2:dockerRegistry.password`) are stored as Pulumi secrets (encrypted)

**All Required Config Keys:**

| Namespace | Key | Used in | Description |
|-----------|-----|---------|-------------|
| `aws` | `region` | `config/GetAvailabilityZone.go` | AWS deployment region |
| `ebs` | `volumeSize` | `config/GetEBSVolumeSize.go` | EBS volume size in GB |
| `ec2` | `ami` | `ec2/getAmi.go` | Amazon Machine Image ID |
| `ec2` | `instanceType` | `ec2/getInstanceType.go` | EC2 instance type (e.g. `t3.xlarge`) |
| `ec2` | `keyPairName` | `ec2/getKeyPairName.go` | SSH key pair name |
| `ec2` | `securityGroupCidrIpv4` | `ec2/getSecurityGroupCidrIpv4.go` | CIDR block for custom-IP ingress rules |
| `ec2` | `ingressSecurityGroups` | `ec2/securitygroup/GetMappingPorts.go` | JSON array of port mapping objects |
| `ec2` | `dockerVersion` | `ec2/userdata/getDockerVersion.go` | Docker package version string |
| `ec2` | `dockerComposeVersion` | `ec2/userdata/getDockerComposeVersion.go` | Docker Compose release tag |
| `ec2` | `dockerRegistry` | `ec2/docker/GetRegistryAuthentication.go` | Object: `server`, `username`, `password` |

**Build:**
- No custom build pipeline — `pulumi up` invokes `go build` automatically via Pulumi's Go runtime

## Platform Requirements

**Development:**
- Go 1.23.1+
- Pulumi CLI v3.0.0+
- AWS credentials configured (environment variables or `~/.aws/credentials`)

**Production / Deployment Target:**
- AWS (region configurable via `aws:region` stack config)
- Provisions: EC2 instance (Amazon Linux 2023, yum-based), EBS volume (`gp3`), Elastic IP, Security Group
- Deployed EC2 instance runs Amazon Linux 2023 (implied by `yum-config-manager`, `yum install docker-*`, `amzn2023` Docker package suffix)

---

*Stack analysis: 2026-05-05*
