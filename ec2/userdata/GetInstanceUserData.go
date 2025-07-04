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
export EBS_VOLUME_FOLDER=ebs-volume
export EBS_FOLDER_PATH=${HOME}/$EBS_VOLUME_FOLDER
export EBS_VOLUME_DEVICE_NAME=/dev/sdh

MOUNT_VOLUME_BINARY_URL=https://github.com/robsoned/mount-ebs-volume-ec2-user-data/releases/download/0.0.2/mount-ebs-volume-ec2-user-data-0.0.2-linux-amd64.tar.gz

curl -L $MOUNT_VOLUME_BINARY_URL -o /tmp/mount-ebs-volume-ec2-user-data.tar.gz && \
tar -xzf /tmp/mount-ebs-volume-ec2-user-data.tar.gz -C /usr/local/bin

# Set the EBS volume device name for later use

EBS_VOLUME_DEVICE_NAME=${EBS_VOLUME_DEVICE_NAME} \
WAIT_EBS_VOLUME_FOLDER_RETRTY_INTERVAL=10 \
WAIT_EBS_VOLUME_FOLDER_MAX_RETRY=10 \
EBS_FOLDER_PATH=${EBS_FOLDER_PATH} \
EBS_VOLUME_FOLDER=${EBS_VOLUME_FOLDER} \
mount-ebs-volume-ec2-user-data

find "${EBS_FOLDER_PATH}" -path "${EBS_FOLDER_PATH}/docker-data" -prune -o -exec chown ec2-user:ec2-user {} +

# Add the EBS volume to fstab for automatic mounting on boot
# Get the UUID of the mounted device (device should be available since mount was successful)
EBS_VOLUME_UUID=$(blkid -s UUID -o value "${EBS_VOLUME_DEVICE_NAME}" 2>/dev/null | head -n1 | tr -d '\n\r')

# Add entry using UUID for persistent mounting across reboots
if [ ! -z "${EBS_VOLUME_UUID}" ] && [ ${#EBS_VOLUME_UUID} -eq 36 ]; then
  # Create backup of fstab before making changes
  cp /etc/fstab /etc/fstab.backup
  
  # Check if this UUID is already in fstab to avoid duplicates
  if ! grep -q "UUID=${EBS_VOLUME_UUID}" /etc/fstab; then
    # Add the new entry with proper formatting
    echo "UUID=${EBS_VOLUME_UUID} ${EBS_FOLDER_PATH} ext4 defaults,nofail 0 2" >> /etc/fstab
    echo "Successfully added EBS volume to fstab: UUID=${EBS_VOLUME_UUID} ${EBS_FOLDER_PATH}"
  else
    echo "EBS volume UUID=${EBS_VOLUME_UUID} already exists in fstab"
  fi
else
  echo "Warning: Could not get valid UUID for device ${EBS_VOLUME_DEVICE_NAME}, skipping fstab entry" >&2
fi

# Add github cli repo (gh) and install docker, git and gh
type -p yum-config-manager >/dev/null || yum install yum-utils
yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo

yum install docker-%s git gh -y && \
usermod -a -G docker ec2-user && \
curl -L "https://github.com/docker/compose/releases/download/%s/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
chmod +x /usr/local/bin/docker-compose && \
service docker start && \
# login to docker registry
su ec2-user -c 'echo "%s" | docker login %s -u %s --password-stdin'

# Configure docker volume to store data in EBS volume

echo "{
  \"data-root\": \"${EBS_FOLDER_PATH}/docker-data\"
}" | tee /etc/docker/daemon.json

mkdir -p ${EBS_FOLDER_PATH}/docker-data && \

service docker restart && \
echo "ClientAliveInterval 60" | tee -a /etc/ssh/sshd_config && \
systemctl restart sshd
`, getDockerVersion(ctx), getDockerComposeVersion(ctx), registryConfig.Password, registryConfig.Server, registryConfig.Username)

	return pulumi.String(userData)
}
