package userdata

import (
	"ec2-instance-docker-dev/ec2/docker"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetInstanceUserData(ctx *pulumi.Context) pulumi.String {

	registryConfig := docker.GetRegistryAuthentication(ctx)

	userData := fmt.Sprintf(`#!/bin/bash
	
set -e

HOME=/home/ec2-user
EBS_VOLUME_DEVICE_NAME=/dev/sdh
EBS_VOLUME_FOLDER=ebs-volume
MAX_RETRIES=10
RETRY_INTERVAL=10

# Wait for the EBS volume to exist
for i in $(seq 1 $MAX_RETRIES); do
  if [ -b $EBS_VOLUME_DEVICE_NAME ]; then
    echo "EBS volume $EBS_VOLUME_DEVICE_NAME is now available."
    break
  fi

  echo "Waiting for EBS volume $EBS_VOLUME_DEVICE_NAME to become available... (attempt $i/$MAX_RETRIES)"
  sleep $RETRY_INTERVAL
done

# Check if the volume is already mounted
if grep -qs $EBS_VOLUME_DEVICE_NAME /proc/mounts; then
  echo "Volume already mounted"
  exit 0
fi

# Check if the filesystem already exists
if ! blkid $EBS_VOLUME_DEVICE_NAME; then
  mkfs -t ext4 $EBS_VOLUME_DEVICE_NAME
else
  # Check if the filesystem is valid
  if ! fsck -n $EBS_VOLUME_DEVICE_NAME; then
    echo "Invalid filesystem detected. Creating a new filesystem..."
    mkfs -t ext4 $EBS_VOLUME_DEVICE_NAME
  fi
fi

mkdir -p $HOME/$EBS_VOLUME_FOLDER

# Check if the mount point is already in use
if mountpoint -q $HOME/$EBS_VOLUME_FOLDER; then
  echo "Mount point already in use"
  exit 0
elif ! mount $EBS_VOLUME_DEVICE_NAME $HOME/$EBS_VOLUME_FOLDER; then
  echo "Failed to mount the volume. Checking filesystem..."
  fsck -y $EBS_VOLUME_DEVICE_NAME
  mount $EBS_VOLUME_DEVICE_NAME $HOME/$EBS_VOLUME_FOLDER
fi

chown -R ec2-user:ec2-user $HOME/$EBS_VOLUME_FOLDER

# Add github cli repo (gh) and install docker, git and gh
type -p yum-config-manager >/dev/null || yum install yum-utils
yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo

yum install docker-%s git gh -y && \
usermod -a -G docker ec2-user && \
curl -L "https://github.com/docker/compose/releases/download/%s/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
chmod +x /usr/local/bin/docker-compose && \
service docker start && \
# login to docker registry
su ec2-user -c 'echo "%s" | docker login %s -u %s --password-stdin' && \
service docker restart && \
echo "ClientAliveInterval 60" | tee -a /etc/ssh/sshd_config && \
systemctl restart sshd
`, getDockerVersion(ctx), getDockerComposeVersion(ctx), registryConfig.Password, registryConfig.Server, registryConfig.Username)

	return pulumi.String(userData)
}
