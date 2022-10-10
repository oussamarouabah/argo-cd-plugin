package backends

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
)

// AzureKeyVault is a struct for working with an Azure Key Vault backend
type AzureKeyVault struct {
	Client keyvault.BaseClient
}

// NewAzureKeyVaultBackend initializes a new Azure Key Vault backend
func NewAzureKeyVaultBackend(client keyvault.BaseClient) *AzureKeyVault {
	return &AzureKeyVault{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Azure SDK
func (a *AzureKeyVault) Login() error {
	return nil
}

// GetSecrets gets secrets from Azure Key Vault and returns the formatted data
// For Azure Key Vault, `kvpath` is the unique name of your vault
func (a *AzureKeyVault) GetSecrets(kvpath string, version string, _ map[string]string) (map[string]interface{}, error) {
	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := make(map[string]interface{})

	secretList, err := a.Client.GetSecretsComplete(ctx, kvpath, nil)
	if err != nil {
		return nil, err
	}
	// Gather all secrets in Key Vault
	for ; secretList.NotDone(); secretList.NextWithContext(ctx) {
		secret := path.Base(*secretList.Value().ID)
		if version == "" {
			secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, "")
			if err != nil {
				return nil, err
			}
			data[secret] = *secretResp.Value
			continue
		}
		// In Azure Key Vault the versions of a secret is first shown after running GetSecretVersions. So we need
		// to loop through the versions for each secret in order to find the secret that has the specific version.
		secretVersions, _ := a.Client.GetSecretVersionsComplete(ctx, kvpath, secret, nil)
		for ; secretVersions.NotDone(); secretVersions.NextWithContext(ctx) {
			secretVersion := secretVersions.Value()
			// Azure Key Vault has ability to enable/disable a secret, so lets honour that
			if !*secretVersion.Attributes.Enabled {
				continue
			}
			// Secret version matched given version
			if strings.Contains(*secretVersion.ID, version) {
				secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, version)
				if err != nil {
					return nil, err
				}
				data[secret] = *secretResp.Value
			}
		}
	}

	return data, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Azure Key Vault, `kvpath` is the unique name of your vault
// Secrets (placeholders) are directly addressable via the API, so only one call is needed here
func (a *AzureKeyVault) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	data, err := a.Client.GetSecret(ctx, kvpath, secret, version)
	if err != nil {
		return nil, err
	}

	return *data.Value, nil
}

func (a *AzureKeyVault) SetSecretVerion(kvpath, secret, version, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	// First disable the previos version of the wanted value, the create a new secret version with that same value
	var false = false
	_, err := a.Client.UpdateSecret(ctx, kvpath, secret, version, keyvault.SecretUpdateParameters{
		SecretAttributes: &keyvault.SecretAttributes{
			Enabled: &false,
		},
	})
	if err != nil {
		return err
	}

	_, err = a.Client.SetSecret(ctx, kvpath, secret, keyvault.SecretSetParameters{
		Value: &value,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AzureKeyVault) SetIndividualSecret(kvpath, secret, version, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	_, err := a.Client.SetSecret(ctx, kvpath, secret, keyvault.SecretSetParameters{
		Value: &value,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AzureKeyVault) GetSecret(kvpath, secretName string, annotations map[string]string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	secretVersions, err := a.Client.GetSecretVersions(ctx, kvpath, secretName, nil)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})

	// Gather all secrets in Key Vault
	for ; secretVersions.NotDone(); secretVersions.NextWithContext(ctx) {
		for _, value := range secretVersions.Values() {
			version := path.Base(*value.ID)
			secret, err := a.Client.GetSecret(ctx, kvpath, secretName, version)
			if err != nil {
				return nil, err
			}
			data[version] = *secret.Value
		}
	}

	return data, nil
}

func (a *AzureKeyVault) GetAllSecretsInPath(kvpath string, annotations map[string]string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	secretList, err := a.Client.GetSecretsComplete(ctx, kvpath, nil)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)
	for ; secretList.NotDone(); secretList.NextWithContext(ctx) {
		secret := path.Base(*secretList.Value().ID)
		v, err := a.Client.GetSecret(ctx, kvpath, secret, "")
		if err != nil {
			return nil, err
		}
		values[secret] = *v.Value
	}

	return values, nil
}
