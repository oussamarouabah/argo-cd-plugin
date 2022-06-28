/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/kube"
)

func NewGetAllCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "getall",
		Short: "Get secrets from Vault",
		RunE:  getAll,
	}

	command.Flags().StringVarP(&configPath, "config-path", "c", "", "path to a file containing Vault configuration (YAML, JSON, envfile) to use")
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret in the argocd namespace containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD). The namespace can be overridden by using the format <namespace>:<name>")

	return command
}

func getAll(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments")
	}

	var manifests []unstructured.Unstructured
	var err error

	kvpath, secret := args[1], args[2]
	path := args[0]
	if path == StdIn {
		manifests, err = readManifestData(cmd.InOrStdin())
		if err != nil {
			return err
		}
	} else {
		files, err := listFiles(path)
		if len(files) < 1 {
			return fmt.Errorf("no YAML or JSON files were found in %s", path)
		}
		if err != nil {
			return err
		}

		var errs []error
		manifests, errs = readFilesAsManifests(files)
		if len(errs) != 0 {
			errMessages := make([]string, len(errs))
			for idx, err := range errs {
				errMessages[idx] = err.Error()
			}
			return fmt.Errorf("could not read YAML/JSON files:\n%s", strings.Join(errMessages, "\n"))
		}
	}

	if len(manifests) == 0 {
		return fmt.Errorf("No manifests")
	}

	v := viper.New()
	cmdConfig, err := config.New(v, &config.Options{
		SecretName: secretName,
		ConfigPath: configPath,
	})
	if err != nil {
		return fmt.Errorf("Store: failed to create config: %v", err)
	}

	err = cmdConfig.Backend.Login()
	if err != nil {
		return err
	}

	var secMan unstructured.Unstructured
	for _, manifest := range manifests {
		if manifest.GetKind() != "Secret" {
			continue
		}
		if metadata, ok := manifest.Object["metadata"]; ok {
			if metadata.(map[string]interface{})["name"] == secret {
				secMan = manifest
			}
		}
	}

	data, err := cmdConfig.Backend.GetSecret(kvpath, secret, nil)
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range data[secret] {
		temp := kube.NewSecretTemplate(secMan, value.(string), key)
		err := temp.ReplaceValue()
		if err != nil {
			return err
		}
		temp.SetVersion(key)
		output, err := temp.ToYAML()
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s---\n", output)
	}

	return nil
}
