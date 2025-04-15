# EC2 Instance with Docker Development Environment

This project uses [Pulumi](https://www.pulumi.com/) to provision and manage an AWS EC2 instance with Docker pre-installed, configured for development environments.

## Features

- **Automated EC2 Instance Provisioning**: Creates a customizable AWS EC2 instance
- **Docker Pre-Installation**: Automatically installs Docker and Docker Compose on the instance
- **EBS Volume Management**: Creates new or attaches existing EBS volumes
- **Security Group Configuration**: Sets up security groups with customizable ingress/egress rules
- **Docker Registry Authentication**: Configures Docker registry access for private container images

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

## License

This project is licensed under the MIT License.