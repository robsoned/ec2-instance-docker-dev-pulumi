package userdata

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetInstanceUserData(ctx *pulumi.Context) pulumi.String {

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
# install lazydocker
curl https://raw.githubusercontent.com/jesseduffield/lazydocker/master/scripts/install_update_linux.sh | bash
`, getDockerVersion(ctx), getDockerComposeVersion(ctx))

	return pulumi.String(userData)
}
