package docker

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type dockerRegistryConfig struct {
	Server   string
	Username string
	Password string
}

func GetRegistryAuthentication(ctx *pulumi.Context) dockerRegistryConfig {

	pulumiConfig := config.New(ctx, "ec2")

	var registryConfig dockerRegistryConfig

	pulumiConfig.RequireObject("dockerRegistry", &registryConfig)

	registryConfig.Password = getPassword(ctx)

	return registryConfig

}

func getPassword(ctx *pulumi.Context) string {
	pulumiConfig := config.New(ctx, "ec2")

	var registryConfig struct {
		Password string
	}

	pulumiConfig.RequireObject("dockerRegistry", &registryConfig)

	return registryConfig.Password
}
