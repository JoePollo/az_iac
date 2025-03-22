package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/cache/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ErrorHandler(object string, err error) error {
	return fmt.Errorf("Failed to build object %s due to error:\n%v", object, err)
}

func main() {

	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce1", Env)
	CacheName := fmt.Sprintf("rc-yym-%s-usce1", Env)
	RedisCacheContributorDefinitionId := "/providers/Microsoft.Authorization/roleDefinitions/e0f68234-74aa-48ed-b826-c38b57376e17"
	RoleAssignmentName := fmt.Sprintf("ra-yym-%s-usce1", Env)

	pulumi.Run(
		func(
			context *pulumi.Context) error {

			clientConfig, err := authorization.GetClientConfig(context)
			if err != nil {
				log.Fatal(ErrorHandler("GetClientConfig", err))
			}

			principalId := clientConfig.ObjectId

			resourceGroup, err := resources.NewResourceGroup(
				context,
				ResourceGroupName,
				&resources.ResourceGroupArgs{
					Location:          pulumi.String("centralus"),
					ResourceGroupName: pulumi.String(ResourceGroupName),
				})
			if err != nil {
				log.Fatal(ErrorHandler("NewResourceGroup", err))
			}

			redis, err := cache.NewRedis(
				context,
				CacheName,
				&cache.RedisArgs{
					// RedisVersion:       pulumi.String("4"),
					// ReplicasPerPrimary: pulumi.Int(0),
					ResourceGroupName: resourceGroup.Name,
					Name:              pulumi.String(CacheName),
					// ShardCount:         pulumi.Int(0),
					EnableNonSslPort:  pulumi.Bool(true),
					Location:          resourceGroup.Location,
					MinimumTlsVersion: pulumi.String("1.2"),
					// This configuration means the C family will be used, and the instance is 0, equating to a C0 instance.
					Sku: &cache.SkuArgs{
						Capacity: pulumi.Int(0),
						Family:   cache.SkuFamilyC,
						Name:     pulumi.String("Basic"),
					},
					// This means once Redis hits capacity, it will evict data based on least recently used.
					RedisConfiguration: &cache.RedisCommonPropertiesRedisConfigurationArgs{
						MaxmemoryPolicy: pulumi.String("allkeys-lru"),
					},
				})
			if err != nil {
				log.Fatal(ErrorHandler("NewRedis", err))
			}

			_, err = authorization.NewRoleAssignment(
				context,
				RoleAssignmentName,
				&authorization.RoleAssignmentArgs{
					PrincipalId:      pulumi.String(principalId),
					PrincipalType:    pulumi.String("User"),
					Scope:            redis.ID(),
					RoleDefinitionId: pulumi.String(RedisCacheContributorDefinitionId),
				})
			if err != nil {
				log.Fatal(ErrorHandler("NewRoleAssignment", err))
			}

			return nil

		})
}
