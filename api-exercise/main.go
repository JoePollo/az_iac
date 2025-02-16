package main

import (
	"fmt"
	"os"

	api "github.com/pulumi/pulumi-azure-native-sdk/apimanagement/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce1", Env)
	ApiManagementName := fmt.Sprintf("apm-yym-%s-usce1", Env)

	pulumi.Run(func(context *pulumi.Context) error {
		resourceGroup, err := resources.NewResourceGroup(context, ResourceGroupName, nil)
		if err != nil {
			return err
		}

		_, err = api.NewApiManagementService(context, ApiManagementName, &api.ApiManagementServiceArgs{
			ResourceGroupName: resourceGroup.Name,
			Location:          resourceGroup.Location,
			PublisherEmail:    pulumi.String(os.Getenv("EMAIL")),
			PublisherName:     pulumi.String(os.Getenv("NAME")),
			Sku: api.ApiManagementServiceSkuPropertiesArgs{
				Capacity: pulumi.Int(0),
				Name:     pulumi.String("Consumption"),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
