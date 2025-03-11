package main

import (
	"ec2-instance-docker-dev/ec2"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return ec2.CreateInstance(ctx)
	})
}
