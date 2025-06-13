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


MOUNT_VOLUME_BINARY_URL=https://github.com/robsoned/mount-ebs-volume-ec2-user-data/releases/download/0.0.2/mount-ebs-volume-ec2-user-data-0.0.2-linux-amd64.tar.gz

curl -L $MOUNT_VOLUME_BINARY_URL -o /tmp/mount-ebs-volume-ec2-user-data.tar.gz && \
tar -xzf /tmp/mount-ebs-volume-ec2-user-data.tar.gz -C /tmp 

EBS_VOLUME_DEVICE_NAME=/dev/sdh \
EBS_VOLUME_FOLDER=ebs-volume \
WAIT_EBS_VOLUME_FOLDER_RETRTY_INTERVAL=10 \
WAIT_EBS_VOLUME_FOLDER_MAX_RETRY=10 \
EBS_FOLDER_PATH=${HOME}/$EBS_VOLUME_FOLDER \
/tmp/mount-ebs-volume-ec2-user-data

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
su ec2-user -c 'echo "%s" | docker login %s -u %s --password-stdin'

# Configure docker volume to store data in EBS volume

echo "{
  \"data-root\": \"${EBS_FOLDER_PATH}/docker-data\"
}" | tee /etc/docker/daemon.json

service docker restart && \
echo "ClientAliveInterval 60" | tee -a /etc/ssh/sshd_config && \
systemctl restart sshd
`, getDockerVersion(ctx), getDockerComposeVersion(ctx), registryConfig.Password, registryConfig.Server, registryConfig.Username)

	return pulumi.String(userData)
}
