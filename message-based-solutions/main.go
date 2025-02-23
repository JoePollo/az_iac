// 1. Create Service Bus Namespace
// 2. Create queue
// 3. Create console application to send and receive messages

package main

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/servicebus/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ErrorHandler(object string, err error) error {
	return fmt.Errorf("Failed to build %s due to error:\n%w", object, err)
}
func main() {
	pulumi.Run(func(context *pulumi.Context) error {

		Env := os.Getenv("ENV")
		projectPrefix := "yym"
		location := "centralus"

		resourceGroup, err := resources.NewResourceGroup(context, fmt.Sprintf("rg-%s-%s-%s", Env, projectPrefix, location), nil)
		if err != nil {
			return ErrorHandler("NewResourceGroup", err)
		}

		serviceBusNameSpace, err := servicebus.NewNamespace(
			context,
			fmt.Sprintf("sbns-%s-%s-%s", Env, projectPrefix, location),
			&servicebus.NamespaceArgs{
				NamespaceName:     pulumi.String(fmt.Sprintf("sbns-%s-%s-%s", Env, projectPrefix, location)),
				ResourceGroupName: resourceGroup.Name,
				Sku: servicebus.SBSkuArgs{
					Name:     pulumi.String("Basic"),
					Tier:     pulumi.String("Basic"),
					Capacity: pulumi.Int(0),
				},
			},
		)
		if err != nil {
			return ErrorHandler("NewNamespace", err)
		}

		_, err = servicebus.NewQueue(context, "queue", &servicebus.QueueArgs{
			EnablePartitioning: pulumi.Bool(true),
			NamespaceName:      serviceBusNameSpace.Name,
			QueueName:          pulumi.String(fmt.Sprintf("q-%s-%s-%s", Env, projectPrefix, location)),
			ResourceGroupName:  resourceGroup.Name,
		})
		if err != nil {
			return ErrorHandler("NewQueue", err)
		}

		return nil
	})
}
