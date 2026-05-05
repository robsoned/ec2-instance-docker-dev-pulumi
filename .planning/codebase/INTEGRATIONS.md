# External Integrations

**Analysis Date:** 2026-05-05

## APIs & External Services

**AWS (via Pulumi AWS Provider v6):**
- **EC2** — Creates and manages the core instance, security groups, elastic IP, volume attachments
  - SDK/Client: `github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2`
  - Auth: AWS credentials (env vars `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` or `~/.aws/credentials`)
  - Resources used: `ec2.Instance`, `ec2.Eip`, `ec2.EipAssociation`, `ec2.SecurityGroup`, `ec2.VolumeAttachment`
  - Files: `ec2/CreateInstance.go`, `ec2/CreateVolumeAttachment.go`, `ec2/elasticIP/Create.go`, `ec2/elasticIP/CreateEipAssociation.go`, `ec2/getSecurityGroup.go`

- **EBS** — Creates and discovers persistent data volumes (`gp3` type, `RetainOnDelete`)
  - SDK/Client: `github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs`
  - Auth: Same AWS credentials as EC2
  - Resources used: `ebs.NewVolume`, `ebs.GetEbsVolumes` (lookup), `ebs.LookupVolume` (single lookup)
  - Files: `ebs/CreateVolume.go`, `ebs/SearchVolume.go`

- **VPC** — Manages security group ingress/egress rules
  - SDK/Client: `github.com/pulumi/pulumi-aws/sdk/v6/go/aws/vpc`
  - Resources used: `vpc.NewSecurityGroupIngressRule`, `vpc.NewSecurityGroupEgressRule`
  - Files: `ec2/getSecurityGroup.go`

**GitHub Releases (HTTP, unauthenticated):**
- `mount-ebs-volume-ec2-user-data` binary — downloaded at EC2 boot time inside the user-data script
  - URL: `https://github.com/robsoned/mount-ebs-volume-ec2-user-data/releases/download/0.0.2/mount-ebs-volume-ec2-user-data-0.0.2-linux-amd64.tar.gz`
  - Called via `curl -L` in the bash user-data
  - File: `ec2/userdata/GetInstanceUserData.go` (line 23)

**GitHub CLI RPM Repository (HTTP, unauthenticated):**
- `gh` CLI tool — installed via yum at EC2 boot time
  - Repo added: `https://cli.github.com/packages/rpm/gh-cli.repo`
  - File: `ec2/userdata/GetInstanceUserData.go` (line 59–60)

**Docker Registry (private, configurable):**
- A private Docker registry is authenticated during instance boot via `docker login`
  - Auth: `server`, `username`, `password` — read from Pulumi stack config key `ec2:dockerRegistry`
  - Password stored as Pulumi encrypted secret in `Pulumi.<stack>.yaml`
  - Default example server: `ghcr.io` (GitHub Container Registry), but any registry is supported
  - Files: `ec2/docker/GetRegistryAuthentication.go`, `ec2/userdata/GetInstanceUserData.go` (line 68)

**Docker Hub / GitHub Compose Releases (HTTP, unauthenticated):**
- Docker Compose binary — downloaded at EC2 boot time
  - URL: `https://github.com/docker/compose/releases/download/<version>/docker-compose-$(uname -s)-$(uname -m)`
  - File: `ec2/userdata/GetInstanceUserData.go` (line 64)

## Data Storage

**Databases:**
- None — this is infrastructure provisioning code only; no application database is managed

**File Storage:**
- **AWS EBS** (`gp3` volume)
  - Mounted at `/home/ec2-user/ebs-volume` on the provisioned instance
  - Docker daemon's `data-root` is redirected to `${EBS_FOLDER_PATH}/docker-data`
  - Tagged with `PulumiProject`, `PulumiStack`, and `Name` for cross-deployment discovery
  - Created with `pulumi.RetainOnDelete(true)` to survive stack teardowns
  - Volume size: configured via `ebs:volumeSize` Pulumi stack config
  - Files: `ebs/CreateVolume.go`, `ebs/SearchVolume.go`

**Caching:**
- None

## Authentication & Identity

**AWS IAM:**
- Implicit — Pulumi uses the ambient AWS credentials from the execution environment
- No explicit IAM role or policy resources are created in this program

**Docker Registry Auth:**
- Custom — `docker login` credentials injected via EC2 user-data at instance boot
- Implementation: `ec2/docker/GetRegistryAuthentication.go` reads `ec2:dockerRegistry` config object
- Credentials flow: Pulumi config → Go struct `dockerRegistryConfig{Server, Username, Password}` → interpolated into bash user-data script

**SSH Key Pair:**
- AWS EC2 Key Pair referenced by name (`ec2:keyPairName` config)
- Key pair must exist in AWS before deployment; not created by this program
- File: `ec2/getKeyPairName.go`

## Monitoring & Observability

**Error Tracking:**
- None — no external error tracking integrated

**Logs:**
- Pulumi context logging only (`ctx.Log.Info`, `ctx.Log.Error`) during provisioning
- On the provisioned EC2 instance: cloud-init output available at `/var/log/cloud-init-output.log`

## CI/CD & Deployment

**Hosting:**
- AWS (region specified via `aws:region` Pulumi config)
- Target: EC2 (Amazon Linux 2023) with attached EBS volume and Elastic IP

**CI Pipeline:**
- None detected in this repository (no `.github/workflows/`, no CI config files)
- Deployments are manual via `pulumi up`

## Environment Configuration

**Required Pulumi stack config keys (all under `Pulumi.<stack>.yaml`, gitignored):**
- `aws:region` — AWS region
- `ebs:volumeSize` — EBS volume size in GB
- `ec2:ami` — AMI ID string
- `ec2:instanceType` — instance type (e.g. `t3.xlarge`)
- `ec2:keyPairName` — name of existing AWS key pair
- `ec2:securityGroupCidrIpv4` — CIDR for custom-IP ingress
- `ec2:ingressSecurityGroups` — JSON array of `{name, protocol, fromPort, toPort}`
- `ec2:dockerVersion` — Docker package version (e.g. `25.0.6-1.amzn2023.0.1`)
- `ec2:dockerComposeVersion` — Docker Compose version tag (e.g. `v2.29.2`)
- `ec2:dockerRegistry` — object with `server`, `username`, `password` (password should be encrypted secret)

**Secrets location:**
- Pulumi stack config files (`Pulumi.<stack>.yaml`) — gitignored via `.gitignore`
- `ec2:dockerRegistry.password` encrypted via `pulumi config set --secret`
- No `.env` files used

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None — this is a one-way provisioning program; no webhooks or callbacks

## Pulumi State Backend

**State:**
- Pulumi state backend not explicitly configured in repo; defaults to Pulumi Cloud (app.pulumi.com) unless overridden via `PULUMI_BACKEND_URL` environment variable
- Stack outputs exported: `elastic-ip` (public IP of the provisioned instance) — see `ec2/CreateInstance.go` line 54

---

*Integration audit: 2026-05-05*
