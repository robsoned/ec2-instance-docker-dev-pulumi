package userdata

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getDockerComposeVersion(ctx *pulumi.Context) pulumi.String {
	dockerComposeVersion := config.New(ctx, "ec2").Require("dockerComposeVersion")
	return pulumi.String(dockerComposeVersion)
}
