package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
)

const StdIn = "-"

var (
	configPath string
	secretName string
)

func NewStoreCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "store",
		Short: "Store secrets in Vault",
		RunE:  store,
	}

	command.Flags().StringVarP(&configPath, "config-path", "c", "", "path to a file containing Vault configuration (YAML, JSON, envfile) to use")
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret in the argocd namespace containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD). The namespace can be overridden by using the format <namespace>:<name>")

	return command
}

/*
We need a functionality to read- and update secrets stored
in Azure Key-vault from within ArgoCD web interface.
The main concept is to allow a user with access to ArgoCD,
to use its (ArgoCD) UI to maintain specific secret(s) in a given Azure Key-vault,
this includes requesting either all or a specific key-vault secret(s) and
have the ability to adjust them, resulting in a new secret version.

Reference to the plugin:
https://argocd-vault-plugin.readthedocs.io/
https://github.com/argoproj-labs/argocd-vault-plugin

Azure SDK:
https://docs.microsoft.com/en-us/azure/key-vault/secrets/quick-create-go

The plugin is written fully in Golang, as should be kept as a requirement.

The above plugin already has the ability to search and read secret(s)
from the Azure Key-vault. Those need to be further extended with the
following requirements:

1. Implement a function to create/store new version(s) of stored secrets.
    The function also need to get secret metadata (version history, creation date,
	description, expiration date).

2. In the ArgoCD (plugin) UI should be the ability
	(like drop-down list with available secrets)
	to select the appropriate secret in the Azure Key-vault.
	This drop-down list should be available from within each
	application plate in ArgoCD.
3. On secrets selection there should be the ability to update/create
	new version and assign it to application and also from within each
	application. Applying it on each application/pod/service level is important,
	as there could be reasons to use different versions of a secret within
	separate applications.
4. All above functionalities should be bound to access restriction and
	allowed to configure to a specific user group
*/

// store should
func store(cmd *cobra.Command, args []string) error {
	v := viper.New()
	cmdConfig, err := config.New(v, &config.Options{
		SecretName: secretName,
		ConfigPath: configPath,
	})
	if err != nil {
		return fmt.Errorf("store: failed to create config: %v", err)
	}

	err = cmdConfig.Backend.Login()
	if err != nil {
		return err
	}

	log.Println("Succeeded to login to Vault")
	err = cmdConfig.Backend.SetIndividualSecret(args[0], args[1], "", args[2])
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
