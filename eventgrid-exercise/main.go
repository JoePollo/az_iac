// Challenge:
// Enable an Event Grid resource provider
// Create a custom topic
// Create a message endpoint
// Subscribe to a custom topic
// Send an event to a custom topic

package main

import (
	"fmt"
	"os"

	auth "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	event "github.com/pulumi/pulumi-azure-native-sdk/eventgrid/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ErrorHandler(object string, err error) error {
	return fmt.Errorf("Failed to build %s due to error:\n%w", object, err)
}

func main() {
	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce1", Env)
	Location := "centralus"

	pulumi.Run(func(context *pulumi.Context) error {
		_, err := auth.GetClientConfig(context)
		if err != nil {
			return ErrorHandler("ClientConfig", err)
		}

		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(context, ResourceGroupName, &resources.ResourceGroupArgs{
			ResourceGroupName: pulumi.String(ResourceGroupName),
			Location:          pulumi.String(Location),
		})
		if err != nil {
			return ErrorHandler("NewResourceGroup", err)
		}

		_, err = event.NewTopic(context, fmt.Sprintf("tpc-yym-%s-usce1", Env), &event.TopicArgs{
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return ErrorHandler("NewTopic", err)
		}

		return nil
	})
}
