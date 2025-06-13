# EC2 Instance with Docker Development Environment

This project uses [Pulumi](https://www.pulumi.com/) to provision and manage an AWS EC2 instance with Docker pre-installed, configured for development environments.

## Features

- **Automated EC2 Instance Provisioning**: Creates a customizable AWS EC2 instance
- **Docker Pre-Installation**: Automatically installs Docker and Docker Compose on the instance
- **EBS Volume Management**: Creates new or attaches existing EBS volumes with automatic mounting
- **Persistent EBS Volume Configuration**: Automatically configures EBS volumes in `/etc/fstab` for persistent mounting across reboots
- **Smart Volume Retention**: Searches for existing volumes by tags and reuses them across deployments
- **Availability Zone Coordination**: Ensures EC2 instances are created in the same AZ as existing volumes
- **Security Group Configuration**: Sets up security groups with customizable ingress/egress rules
- **Docker Registry Authentication**: Configures Docker registry access for private container images
- **Docker Data Storage on EBS**: Configures Docker to store container data on the persistent EBS volume

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) (v3.0.0 or newer)
- AWS account with credentials configured
- Go 1.23.1 or later

## Project Structure

```
.
├── Pulumi.yaml                # Pulumi project configuration
├── Pulumi.dev.yaml            # Default stack configuration
├── main.go                    # Main entry point for Pulumi program
├── config/                    # Utility functions for configuration
├── ebs/                       # EBS volume management functions
└── ec2/                       # EC2 instance and related resource functions
    ├── docker/                # Docker-related configurations
    ├── elasticIP/             # Elastic IP address functions
    ├── securitygroup/         # Security group management
    └── userdata/              # User data scripts for instance initialization
```

## Configuration

The following configuration variables can be set in `Pulumi.dev.yaml`:

| Variable | Description | Example |
|----------|-------------|---------|
| aws:region | AWS region to deploy resources | `us-east-1` |
| ebs:volumeSize | Size of the EBS volume in GB | `100` |
| ec2:ami | Amazon Machine Image ID | `ami-0453ec754f44f9a4a` |
| ec2:dockerVersion | Docker version to install | `25.0.6-1.amzn2023.0.1` |
| ec2:dockerComposeVersion | Docker Compose version to install | `v2.29.2` |
| ec2:instanceType | EC2 instance type | `t3.xlarge` |
| ec2:keyPairName | SSH key pair name | `dev-ec2` |
| ec2:securityGroupCidrIpv4 | IP CIDR block for security group | `0.0.0.0/0` |
| ec2:ingressSecurityGroups | Security group ingress rules | See example below |
| ec2:dockerRegistry | Docker registry credentials | See example below |

### Volume Management

The system automatically manages EBS volumes with the following behavior:
- **Volume Search**: Searches for existing volumes using project and stack tags
- **Volume Reuse**: If an existing volume is found, it will be attached to the new instance
- **Volume Creation**: If no existing volume is found, a new one will be created
- **Availability Zone Coordination**: Instances are created in the same AZ as existing volumes
- **Automatic Mounting**: Volumes are automatically mounted to `/home/ec2-user/ebs-volume`
- **Persistent Configuration**: Volume mount points are added to `/etc/fstab` for automatic mounting on boot

### Docker Configuration

- **Docker Data Storage**: Docker daemon is configured to store all data on the EBS volume
- **Registry Authentication**: Supports private Docker registry authentication
- **Automatic Startup**: Docker service is automatically started and configured to start on boot

### Example Security Group Configuration

```yaml
ec2:ingressSecurityGroups:
  - fromPort: 22
    name: ssh
    protocol: tcp
    toPort: 22
  - fromPort: 0
    name: all
    protocol: tcp
    toPort: 65535
```

### Example Docker Registry Configuration

```yaml
ec2:dockerRegistry:
  password:
    secure: <encrypted_password>
  username: username
  server: ghcr.io
```

## Troubleshooting

### EBS Volume Mount Issues

If you encounter issues with EBS volume mounting during instance initialization:

1. **Check Cloud-Init Logs**: SSH into the instance and check `/var/log/cloud-init-output.log` for detailed error messages
2. **Verify Volume Attachment**: Ensure the EBS volume is properly attached to the instance at `/dev/sdh`
3. **Check fstab Configuration**: The system automatically adds UUID-based entries to `/etc/fstab` for persistent mounting

### Volume Management

- **Existing Volumes**: The system will automatically find and reuse existing volumes tagged with the project and stack name
- **Availability Zone Conflicts**: If you need to change regions, make sure to update or remove existing volumes first

## License

This project is licensed under the MIT License.