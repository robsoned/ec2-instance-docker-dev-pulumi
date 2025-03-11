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

# Add github cli repo (gh) and install docker, git and gh
type -p yum-config-manager >/dev/null || sudo yum install yum-utils
sudo yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo

sudo yum install docker-%s git gh -y && \
sudo usermod -a -G docker ec2-user && \
sudo curl -L "https://github.com/docker/compose/releases/download/%s/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
sudo chmod +x /usr/local/bin/docker-compose && \
sudo service docker start && \
# login to docker registry
su ec2-user -c 'echo "%s" | docker login %s -u %s --password-stdin' && \
sudo service docker restart && \
echo "ClientAliveInterval 60" | sudo tee -a /etc/ssh/sshd_config && \
systemctl restart sshd
`, getDockerVersion(ctx), getDockerComposeVersion(ctx), registryConfig.Password, registryConfig.Server, registryConfig.Username)

	return pulumi.String(userData)
}
