package main

import (
	"fmt"
	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"os"
)

func ErrorHandler(object string, err error) error {
	return fmt.Errorf("Failed to build %s due to error:\n%w", object, err)
}

func main() {
	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce", Env)

	pulumi.Run(func(context *pulumi.Context) error {
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(context, ResourceGroupName, nil)
		if err != nil {
			return ErrorHandler("NewResourceGroup", err)
		}

		clientConfig, err := authorization.GetClientConfig(context)
		if err != nil {
			return ErrorHandler("GetClientConfig", err)
		}

		principalId := clientConfig.ObjectId

		// Create an Azure resource (Storage Account)
		account, err := storage.NewStorageAccount(context, fmt.Sprintf("styym%susce", Env), &storage.StorageAccountArgs{
			ResourceGroupName: resourceGroup.Name,
			Sku: &storage.SkuArgs{
				Name: storage.SkuName_Standard_LRS,
			},
			Kind: pulumi.String("StorageV2"),
		})
		if err != nil {
			return ErrorHandler("NewStorageAccount", err)
		}

		storageQueue, err := storage.NewQueue(context, fmt.Sprintf("stq-yym-%s-usce", Env), &storage.QueueArgs{
			AccountName:       account.Name,
			QueueName:         pulumi.String(fmt.Sprintf("stq-yym-%s-usce", Env)),
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return ErrorHandler("NewQueue", err)
		}

		_, err = authorization.NewRoleAssignment(
			context,
			fmt.Sprintf("ra-yym-sqc-%s-usce", Env),
			&authorization.RoleAssignmentArgs{
				PrincipalId:      pulumi.String(principalId),
				RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/974c5e8b-45b9-4653-ba55-5f855dd0fb88"),
				PrincipalType:    pulumi.String("User"),
				Scope:            storageQueue.ID(),
			})
		if err != nil {
			return ErrorHandler("QueueContributorRoleAssignment", err)
		}

		return nil
	})
}
