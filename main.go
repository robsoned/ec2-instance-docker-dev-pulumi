package main

import (
	"ec2-instance-docker-dev/ebs"
	"ec2-instance-docker-dev/ec2"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		retainedVolume, err := ebs.SearchVolume(ctx)

		if err != nil {
			ctx.Log.Error("Error searching for existing volume", nil)
			return err
		}

		ec2Instance, err := ec2.CreateInstance(ctx, retainedVolume)

		if err != nil {
			ctx.Log.Error("Error creating EC2 instance", nil)
			return err
		}

		if retainedVolume != nil {
			ctx.Log.Info("Volume found, new volume will not be created", nil)
			_, err = ec2.CreateVolumeAttachment(ctx, pulumi.String(retainedVolume.ID), ec2Instance)

			return err
		}

		volume, err := ebs.CreateVolume(ctx, ec2Instance.AvailabilityZone)

		if err != nil {
			return err
		}

		_, err = ec2.CreateVolumeAttachment(ctx, volume.ID(), ec2Instance)

		if err != nil {
			return err
		}

		return nil

	})
}
