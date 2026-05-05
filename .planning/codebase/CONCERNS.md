# Codebase Concerns

**Analysis Date:** 2026-05-05

---

## Tech Debt

**Dead struct — `CreateVolumeResult`:**
- Issue: `CreateVolumeResult` struct (with `VolumeIDString *string` and `VolumeIDOutput *pulumi.IDOutput` fields) is defined but never used. The accompanying `// todo` comment confirms incomplete refactoring.
- Files: `ebs/CreateVolume.go:14-18`
- Impact: Dead code inflates cognitive load; the TODO signals the current return type (`*ebs.Volume`) may not satisfy a planned caller contract.
- Fix approach: Either implement the dual-path return (plain string vs. `pulumi.IDOutput`) or delete the struct and the TODO.

**Dead function — `GetAvailabilityZone`:**
- Issue: `config.GetAvailabilityZone(ctx)` is defined and exported but called nowhere in the codebase. It also reads the wrong config key (`aws:region` returns a region string like `us-east-1`, not an AZ like `us-east-1a`).
- Files: `config/GetAvailabilityZone.go`
- Impact: Misleading API surface; if called, would silently pass a region string where an AZ string is expected, causing AWS API errors at apply-time.
- Fix approach: Delete the function if unused. If AZ config is needed in the future, read the correct `aws:availabilityZone` key or derive it from the existing volume's `AvailabilityZone` field (already in `SearchVolumeOutput`).

**Unused struct field — `CreateElasticIPArgs.InstanceId`:**
- Issue: `elasticIP.CreateElasticIPArgs` declares an `InstanceId pulumi.StringInput` field, but `Create()` never reads it. Every call site passes an empty struct: `&elasticIP.CreateElasticIPArgs{}`.
- Files: `ec2/elasticIP/Create.go:8-10`, `ec2/CreateInstance.go:14`
- Impact: Dead API surface; future callers may mistakenly believe passing `InstanceId` affects behaviour.
- Fix approach: Remove the struct and its field; replace with a zero-argument function signature `Create(ctx *pulumi.Context)`.

**Go module name violates convention:**
- Issue: `go.mod` declares `module ec2-instance-docker-dev`, a bare non-URL module path.
- Files: `go.mod:1`
- Impact: Non-standard; tooling and IDE support may behave unexpectedly. The Go module specification requires modules intended for external use to be URL-rooted.
- Fix approach: Rename to a fully qualified path (e.g., `github.com/robsoned/ec2-instance-docker-dev-pulumi`) and update all internal `import` statements.

---

## Known Bugs

**Typo in EBS volume name constant — volume search will silently fail after correction:**
- Symptoms: The constant `pulumiEBSName = "ec2-instane-ebs-volume"` contains a typo ("instane" not "instance"). All existing volumes are tagged with this misspelled name. Correcting the typo in the future will break `SearchVolume`, causing it to find no volume and create a duplicate.
- Files: `ebs/CreateVolume.go:12`, `ebs/SearchVolume.go` (consumes the constant)
- Trigger: Any refactor that corrects the spelling while volumes already exist in AWS with the old tag.
- Workaround: Leave the constant as-is intentionally (at the cost of a permanent misspelling), or perform a coordinated tag-rename + constant update with a one-time migration.

**`GetAvailabilityZone` returns region string, not AZ string:**
- Symptoms: Calling `config.GetAvailabilityZone(ctx)` returns a value like `"us-east-1"`, not an AZ like `"us-east-1a"`. Passing this to `AvailabilityZone` on any AWS resource will cause an apply-time API error.
- Files: `config/GetAvailabilityZone.go:9`
- Trigger: Any code path that calls this function and uses the result as an AZ.
- Workaround: Function is currently uncalled; treat as latent bug.

---

## Security Considerations

**Docker registry password embedded in plaintext user-data:**
- Risk: `GetInstanceUserData` uses `fmt.Sprintf` to interpolate `registryConfig.Password` directly into the bash cloud-init script. AWS EC2 user-data is stored in instance metadata and retrievable by anyone with `ec2:DescribeInstanceAttribute` permission, or from within the instance itself at `http://169.254.169.254/latest/user-data`.
- Files: `ec2/userdata/GetInstanceUserData.go:68`, `ec2/docker/GetRegistryAuthentication.go`
- Current mitigation: Pulumi config encrypts the secret at rest in `Pulumi.*.yaml` (via `secure:` key). The value is only decrypted at apply time — but is then written in cleartext to the EC2 user-data payload.
- Recommendations:
  1. Use AWS Secrets Manager or SSM Parameter Store (SecureString). Store the ARN in Pulumi config, and fetch the secret inside the user-data script via `aws ssm get-parameter --with-decryption`.
  2. Restrict the EC2 instance IAM role to only the required secret ARN.

**`Pulumi.*.yaml` excluded from git but no documented secret rotation process:**
- Risk: Stack config files (containing encrypted secrets) are gitignored (`Pulumi.*.yaml`). There is no guidance in the README for secret rotation or what to do when a Pulumi passphrase is compromised.
- Files: `.gitignore:1`, `Readme.md`
- Current mitigation: Files are not committed.
- Recommendations: Document the secret rotation procedure and the Pulumi passphrase storage location (e.g., team password manager).

**Overly broad security group CIDR in README example:**
- Risk: The documented example for `ec2:securityGroupCidrIpv4` uses `0.0.0.0/0`, which opens all configured ingress ports (including SSH port 22) to the entire internet.
- Files: `Readme.md:52`, `ec2/getSecurityGroupCidrIpv4.go`
- Current mitigation: None enforced in code; it is a config value.
- Recommendations: Add a Pulumi policy-as-code check (CrossGuard) that blocks `0.0.0.0/0` on port 22, or document a required IP restriction.

**Egress rule allows all traffic:**
- Risk: The security group egress rule allows all outbound traffic (`0.0.0.0/0`, protocol `-1`, ports 0-0).
- Files: `ec2/getSecurityGroup.go:41-51`
- Current mitigation: This is standard practice for development environments, but increases blast radius if the instance is compromised.
- Recommendations: For production hardening, restrict egress to known endpoints (e.g., registry server, package repos, SSM endpoint).

---

## Fragile Areas

**Device name `/dev/sdh` duplicated across two unlinked files:**
- Files: `ec2/CreateVolumeAttachment.go:18`, `ec2/userdata/GetInstanceUserData.go:20-21`
- Why fragile: The EBS attachment device name is hardcoded in two separate places that are not connected in code. If one is changed without updating the other, the user-data script will wait indefinitely for a device that either doesn't exist at the expected path, or is attached under a different name (modern kernels rename `/dev/sdh` → `/dev/xvdh`).
- Safe modification: Extract the device name to a shared constant in a `config` or `shared` package, imported by both files.
- Test coverage: No tests; failure surfaces only at runtime after instance launch.

**Self-IP ingress rule error silently discarded inside `ApplyT`:**
- Files: `ec2/getSecurityGroup.go:31-39`
- Why fragile: The error returned by `createSecurityGroupIngresses` inside the `ApplyT` callback is returned as the function's return value, but the output of `ApplyT` itself is never captured or checked. Pulumi does not surface errors returned from `ApplyT` callbacks as deployment errors. If this ingress rule creation fails, the deployment will succeed but the self-referencing security group rule will be missing silently.
- Safe modification: Move the self-IP ingress creation outside of `ApplyT` by passing the elastic IP's `PublicIp` output into a `pulumi.All(...)` combined apply, or use `ctx.RegisterResource`-compatible patterns.
- Test coverage: None.

**Double AWS API call in `SearchVolume` — fragile when >1 volume matches:**
- Files: `ebs/SearchVolume.go:16-56`
- Why fragile: `GetEbsVolumes` checks for existence (returns IDs), then `LookupVolume` re-fetches the single volume. If more than one volume happens to match the tag filters (e.g., a volume from a previous failed destroy), `LookupVolume` returns an error "multiple volumes found". This causes every subsequent `pulumi up` to fail until the orphaned volume is manually removed from AWS.
- Safe modification: Switch to using `GetEbsVolumes` exclusively and select `IDs[0]` / `AvailabilityZones[0]` from the result to eliminate the redundant call and the multiple-match failure mode.
- Test coverage: None.

**External binary fetched from personal GitHub repo at a pinned version:**
- Files: `ec2/userdata/GetInstanceUserData.go:23`
- Why fragile: The user-data script downloads `https://github.com/robsoned/mount-ebs-volume-ec2-user-data/releases/download/0.0.2/mount-ebs-volume-ec2-user-data-0.0.2-linux-amd64.tar.gz` at instance boot. If the release is deleted, the repo is renamed, or GitHub is unreachable at boot time, instance initialisation fails silently (the user-data `set -e` will abort but cloud-init logs are the only indication).
- Safe modification: Mirror the binary to an S3 bucket under the same AWS account and fetch from there. Add SHA-256 checksum verification after download.
- Test coverage: None; failure surfaces only after instance launch.

**Root block device size hardcoded to 100 GB:**
- Files: `ec2/CreateInstance.go:33`
- Why fragile: `VolumeSize: pulumi.Int(100)` is not configurable, unlike the EBS data volume size which reads from `ebs:volumeSize`. If the root volume needs to be resized, a code change is required.
- Safe modification: Add a `ec2:rootVolumeSize` config key following the same pattern as `config.GetEBSVolumeSize`.

---

## Performance Bottlenecks

**`SearchVolume` makes two sequential AWS API calls on every `pulumi up`:**
- Problem: `GetEbsVolumes` followed by `LookupVolume` are two separate AWS API calls that could be collapsed into one (see Fragile Areas above).
- Files: `ebs/SearchVolume.go`
- Cause: Redundant API call design.
- Improvement path: Use only `GetEbsVolumes` and index directly into results.

---

## Dependencies at Risk

**Personal project dependency — `mount-ebs-volume-ec2-user-data`:**
- Risk: The user-data script depends on a binary from `github.com/robsoned/mount-ebs-volume-ec2-user-data` — a personal project with a single release (`0.0.2`). It has no stated maintenance commitment, no verification checksum, and is architecture-specific (`linux-amd64`).
- Impact: Instance initialisation fails if the release is removed or the repo is unavailable.
- Migration plan: Either vendor the binary into an owned S3 bucket, or inline its functionality (EBS mount and fstab configuration) directly in the user-data bash script.

---

## Test Coverage Gaps

**Zero test coverage across the entire codebase:**
- What's not tested: All resource creation functions, config retrieval helpers, volume search logic, user-data generation, security group construction, and the main orchestration flow in `main.go`.
- Files: Every `.go` file in the project. No `*_test.go` files exist.
- Risk: Any change to config key names, resource argument shapes, or user-data template logic is undetected until `pulumi preview` or `pulumi up` runs against a live AWS account.
- Priority: High
- Recommended test types:
  1. Unit tests for pure functions: `config/GetEBSVolumeSize.go`, `ec2/securitygroup/GetMappingPorts.go`, `ec2/userdata/GetInstanceUserData.go` (validate template output).
  2. Pulumi unit tests using `pulumi.RunWithContext` with mocked providers for `CreateInstance`, `SearchVolume`, `CreateVolume`.
  3. Integration/smoke test verifying `pulumi preview` completes successfully against a test stack config.

---

## Missing Critical Features

**No IAM instance profile attached to the EC2 instance:**
- Problem: The EC2 instance is created with no IAM role or instance profile. Any AWS API calls from inside the instance (e.g., fetching secrets from SSM, pulling from ECR, writing to S3) require either hardcoded credentials or manual role attachment after deploy.
- Blocks: Implementing the recommended SSM-based secret retrieval for the Docker registry password.

**No SSH key pair creation — relies on pre-existing key pair:**
- Problem: `getKeyPairName` reads an existing key pair name from config. If the key pair does not exist in the target AWS region, the instance launches but is unreachable. There is no validation or key pair creation step.
- Files: `ec2/getKeyPairName.go`

**No outputs exported beyond the Elastic IP:**
- Problem: `main.go` exports only `elastic-ip`. The EBS volume ID, instance ID, and security group ID are not exported, making it difficult to reference these resources from other Pulumi stacks or scripts without running `pulumi stack export`.
- Files: `ec2/CreateInstance.go:54`, `main.go`

---

*Concerns audit: 2026-05-05*
