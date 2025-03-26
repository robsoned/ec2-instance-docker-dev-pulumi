package main

import (
	"ec2-instance-docker-dev/ebs"
	"ec2-instance-docker-dev/ec2"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		ec2Instance, err := ec2.CreateInstance(ctx)

		if err != nil {
			return err
		}

		instanceAvailabilityZone := ec2Instance.AvailabilityZone

		retainedVolumeID, err := ebs.SearchVolume(ctx)

		if err != nil {
			return err
		}

		if retainedVolumeID != "" {
			ctx.Log.Info("Volume found, new volume will not be created", nil)
			_, err = ec2.CreateVolumeAttachment(ctx, pulumi.String(retainedVolumeID), ec2Instance)

			return err
		}

		volume, err := ebs.CreateVolume(ctx, instanceAvailabilityZone)

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
