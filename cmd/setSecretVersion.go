package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
)

func NewSetSecretVersionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "upsecret",
		Short: "Store secrets in Vault",
		RunE:  setSecretVersion,
	}

	command.Flags().StringVarP(&configPath, "config-path", "c", "", "path to a file containing Vault configuration (YAML, JSON, envfile) to use")
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret in the argocd namespace containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD). The namespace can be overridden by using the format <namespace>:<name>")

	return command
}

// store should
func setSecretVersion(cmd *cobra.Command, args []string) error {
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
	path, secret, value, version := args[0], args[1], args[2], args[3]
	err = cmdConfig.Backend.SetSecretVerion(path, secret, version, value)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
