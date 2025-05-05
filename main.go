package main

import (
	"ec2-instance-docker-dev/ebs"
	"ec2-instance-docker-dev/ec2"

	awsPulumi "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		searchVolumeOutput, err := ebs.SearchVolume(ctx)

		if err != nil {
			return err
		}

		var ec2Instance *awsPulumi.Instance

		if searchVolumeOutput != nil {
			// If volume is found, create instance in the same availability zone
			ctx.Log.Info("Volume found, creating instance in the same availability zone", nil)
			availabilityZone := pulumi.String(searchVolumeOutput.AvailabilityZone)
			// Convert pulumi.String to pulumi.StringInput
			var availabilityZoneInput pulumi.StringInput = availabilityZone
			ec2Instance, err = ec2.CreateInstance(ctx, &availabilityZoneInput)
		} else {
			// If no volume found, create instance in any availability zone
			ctx.Log.Info("No volume found, creating instance in default availability zone", nil)
			ec2Instance, err = ec2.CreateInstance(ctx, nil)
		}

		if err != nil {
			return err
		}

		instanceAvailabilityZone := ec2Instance.AvailabilityZone

		if searchVolumeOutput != nil {
			// Use the ID from the SearchVolumeOutput struct
			ctx.Log.Info("Attaching existing volume to instance", nil)
			_, err = ec2.CreateVolumeAttachment(ctx, pulumi.String(searchVolumeOutput.Id), ec2Instance)
			return err
		}

		// Create a new volume in the instance's availability zone
		ctx.Log.Info("Creating new volume in instance's availability zone", nil)
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
