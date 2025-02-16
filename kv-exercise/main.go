package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	auth "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	keyvault "github.com/pulumi/pulumi-azure-native-sdk/keyvault/v2"
	resources "github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Generates a secret in N bytes and converts to UTF8 representation
func GenerateSecretValue(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func main() {
	// Define vars for config
	Env := os.Getenv("ENV")
	ResourceGroupName := fmt.Sprintf("rg-yym-%s-usce1", Env)
	KeyVaultName := fmt.Sprintf("kv-yym-%s-usce1", Env)
	CoolSecret, err := GenerateSecretValue(16)
	if err != nil {
		return
	}

	// Main entry point for Pulumi runtime, provides context from the Provider
	pulumi.Run(func(context *pulumi.Context) error {
		// Retrieves context from Provider to expose attributes, IE tenant ID.
		clientConfig, err := auth.GetClientConfig(context)
		if err != nil {
			return err
		}
		// Generic resource group.
		resourceGroup, err := resources.NewResourceGroup(context, "resourceGroup", &resources.ResourceGroupArgs{
			Location:          pulumi.String("centralus"),
			ResourceGroupName: pulumi.String(ResourceGroupName),
		})
		if err != nil {
			return err
		}

		// Base tier keyvault
		keyVault, err := keyvault.NewVault(context, KeyVaultName, &keyvault.VaultArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			VaultName:         pulumi.String(KeyVaultName),
			Properties: &keyvault.VaultPropertiesArgs{
				Sku: keyvault.SkuArgs{
					Family: pulumi.String(keyvault.SkuFamilyA),
					Name:   keyvault.SkuNameStandard,
				},
				TenantId:                pulumi.String(clientConfig.TenantId),
				EnableRbacAuthorization: pulumi.Bool(true),
			},
		})
		if err != nil {
			return err
		}

		// Adds the identity that Pulumi is utilizing through the Provider as a secrets admin.
		_, err = auth.NewRoleAssignment(context, fmt.Sprintf("ra-yym-%s-%s-usce1", "kvadm", Env), &auth.RoleAssignmentArgs{
			PrincipalId:      pulumi.String(clientConfig.ObjectId),
			PrincipalType:    auth.PrincipalTypeUser,
			RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/00482a5a-887f-4fb3-b363-3b7fe8e74483"),
			Scope:            keyVault.ID(),
		})
		if err != nil {
			return err
		}

		// Adds a secret to above referenced key vault using the generateSecretValue func.
		_, err = keyvault.NewSecret(context, "superCoolSecret", &keyvault.SecretArgs{
			ResourceGroupName: resourceGroup.Name,
			VaultName:         keyVault.Name,
			SecretName:        pulumi.String("superCoolSecret"),
			Properties: &keyvault.SecretPropertiesArgs{
				Attributes: &keyvault.SecretAttributesArgs{
					Enabled:   pulumi.Bool(true),
					Expires:   pulumi.Int(0),
					NotBefore: pulumi.Int(0),
				},
				Value:       pulumi.String(CoolSecret),
				ContentType: pulumi.String("string"),
			},
		}, pulumi.IgnoreChanges([]string{"properties.value"}))
		if err != nil {
			return err
		}

		return nil
	})
}
