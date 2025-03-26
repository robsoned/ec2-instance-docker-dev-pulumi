package ec2

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateVolumeAttachment(ctx *pulumi.Context, volumeID pulumi.StringInput, ec2Instance *ec2.Instance) (*ec2.VolumeAttachment, error) {

	ctx.Log.Info("Creating volume attachment", nil)

	volumeID.ToStringOutput().ApplyT(func(id string) error {
		ctx.Log.Info("Volume ID: "+id, nil)
		return nil
	})

	return ec2.NewVolumeAttachment(ctx, "ebs-volume-attachment", &ec2.VolumeAttachmentArgs{
		DeviceName: pulumi.String("/dev/sdh"),
		InstanceId: ec2Instance.ID(),
		VolumeId:   volumeID,
	})

}
