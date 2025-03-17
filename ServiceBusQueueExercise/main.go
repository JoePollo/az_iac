package main

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/servicebus/v2"
	// "github.com/pulumi/pulumi-azure-native/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ErrorHandler(object string, err error) error {
	return fmt.Errorf("Failed to build %s due to error:\n%w", object, err)
}

func main() {
	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce", Env)
	ServiceBusNamespaceName := fmt.Sprintf("sbns-yym-%s-usce", Env)
	ServiceBusQueueName := fmt.Sprintf("sbq-yym-%s-usce", Env)
	Location := "centralus"

	pulumi.Run(func(context *pulumi.Context) error {

		clientConfig, err := authorization.GetClientConfig(context)
		if err != nil {
			return err
		}

		principalId := clientConfig.ObjectId

		resourceGroup, err := resources.NewResourceGroup(
			context,
			ResourceGroupName,
			&resources.ResourceGroupArgs{
				ResourceGroupName: pulumi.String(ResourceGroupName),
				Location:          pulumi.String(Location),
			})
		if err != nil {
			return ErrorHandler("NewResourceGroup", err)
		}

		serviceBusNamespace, err := servicebus.NewNamespace(
			context,
			ServiceBusNamespaceName,
			&servicebus.NamespaceArgs{
				Location:          pulumi.String(Location),
				NamespaceName:     pulumi.String(ServiceBusNamespaceName),
				ResourceGroupName: resourceGroup.Name,
				Sku: &servicebus.SBSkuArgs{
					Name: pulumi.String(servicebus.SkuNameBasic),
					Tier: pulumi.String(servicebus.SkuTierBasic),
				},
			})
		if err != nil {
			return ErrorHandler("NewNamespace", err)
		}

		serviceBusQueue, err := servicebus.NewQueue(
			context, ServiceBusQueueName, &servicebus.QueueArgs{
				ResourceGroupName: resourceGroup.Name,
				QueueName:         pulumi.String(ServiceBusQueueName),
				NamespaceName:     serviceBusNamespace.Name,
			})
		if err != nil {
			return ErrorHandler("NewQueue", err)
		}

		_, err = authorization.NewRoleAssignment(
			context, fmt.Sprintf("ra-yym-sbqs-%s-usce", Env), &authorization.RoleAssignmentArgs{
				PrincipalId:      pulumi.String(principalId),
				RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/69a216fc-b8fb-44d8-bc22-1f3c2cd27a39"),
				PrincipalType:    pulumi.String("User"),
				Scope:            serviceBusQueue.ID(),
			})
		if err != nil {
			return ErrorHandler("RoleAssignment", err)
		}

		_, err = authorization.NewRoleAssignment(
			context, fmt.Sprintf("ra-yym-sbqr-%s-usce", Env), &authorization.RoleAssignmentArgs{
				PrincipalId:      pulumi.String(principalId),
				RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/4f6d3b9b-027b-4f4c-9142-0e5a2a2247e0"),
				PrincipalType:    pulumi.String("User"),
				Scope:            serviceBusQueue.ID(),
			})
		if err != nil {
			return ErrorHandler("RoleAssignment", err)
		}

		return nil
	})

}
