package userdata

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getDockerVersion(ctx *pulumi.Context) pulumi.String {

	dockerVersion := config.New(ctx, "ec2").Require("dockerVersion")

	return pulumi.String(dockerVersion)

}
