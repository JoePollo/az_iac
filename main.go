package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/app"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(context *pulumi.Context) error {
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(context, "resourceGroup", nil)
		if err != nil {
			return err
		}

		// This is an immutable value of resourceGroup.Location.
		// A pointer would cause this value to be able to change the
		// value of resourceGroup.Location.
		// For example: location := &resourceGroup.Location
		location := resourceGroup.Location

		// Create an Azure resource (Storage Account)
		account, err := storage.NewStorageAccount(context, "sa", &storage.StorageAccountArgs{
			ResourceGroupName: resourceGroup.Name,
			AccountName:       pulumi.String("certexploration"),
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Standard_LRS"),
			},
			Kind: pulumi.String("StorageV2"),
		})
		if err != nil {
			return err
		}

		containerAppEnvironment, err := app.NewManagedEnvironment(context, "az204ContEnv", &app.ManagedEnvironmentArgs{
			EnvironmentName:   pulumi.String("az204ContEnv"),
			ResourceGroupName: resourceGroup.Name,
			Location:          location,
		})
		if err != nil {
			return err
		}
		// This will allow for a containerApp to be created without storing the return value
		// Go requires all return values are used, and this containerApp isn't utilized further,
		// only instantiated in Azure.

		containerApp, err := app.NewContainerApp(context, "azCertContApp", &app.ContainerAppArgs{
			ResourceGroupName: resourceGroup.Name,
			ContainerAppName:  pulumi.String("azcertcontapp"),
			Location:          location,
			EnvironmentId:     containerAppEnvironment.ID(),
			Configuration: &app.ConfigurationArgs{
				Ingress: &app.IngressArgs{
					External:   pulumi.Bool(true),
					TargetPort: pulumi.Int(80),
				},
			},
			Template: &app.TemplateArgs{
				Containers: app.ContainerArray{
					&app.ContainerArgs{
						Name:  pulumi.String("helloworldcontainer"),
						Image: pulumi.String("mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Export the primary key of the Storage Account
		context.Export("primaryStorageKey", pulumi.All(resourceGroup.Name, account.Name).ApplyT(
			func(args []interface{}) (string, error) {
				resourceGroupName := args[0].(string)
				accountName := args[1].(string)
				accountKeys, err := storage.ListStorageAccountKeys(context, &storage.ListStorageAccountKeysArgs{
					ResourceGroupName: resourceGroupName,
					AccountName:       accountName,
				})
				if err != nil {
					return "", err
				}

				return accountKeys.Keys[0].Value, nil
			},
		))

		// Export the Fully Qualified Domain Name of the Container App
		context.Export("FQDN", containerApp.LatestRevisionFqdn)

		return nil
	})
}
