package backends

import (
	"context"
	"fmt"
	"log"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

// YandexCloudLockbox is a struct for working with a Yandex Cloud lockbox backend
type YandexCloudLockbox struct {
	client lockbox.PayloadServiceClient
}

// NewYandexCloudLockboxBackend initializes a new Yandex Cloud lockbox backend
func NewYandexCloudLockboxBackend(client lockbox.PayloadServiceClient) *YandexCloudLockbox {
	return &YandexCloudLockbox{
		client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the lockbox
func (ycl *YandexCloudLockbox) Login() error {
	return nil
}

// GetSecrets gets secrets from lockbox and returns the formatted data
func (ycl *YandexCloudLockbox) GetSecrets(secretID string, version string, _ map[string]string) (map[string]interface{}, error) {
	req := &lockbox.GetPayloadRequest{
		SecretId: secretID,
	}

	if version != "" {
		req.SetVersionId(version)
	}

	resp, err := ycl.client.Get(context.Background(), req)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(resp.GetEntries()))
	for _, v := range resp.GetEntries() {
		result[v.GetKey()] = v.GetTextValue()
	}

	return result, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the lockbox backend
func (ycl *YandexCloudLockbox) GetIndividualSecret(secretID, key, version string, _ map[string]string) (interface{}, error) {
	secrets, err := ycl.GetSecrets(secretID, version, nil)
	if err != nil {
		return nil, err
	}

	secret, found := secrets[key]
	if !found {
		return nil, fmt.Errorf("secretID: %s, key: %s, version: %s not found", secretID, key, version)
	}

	return secret, nil
}

func (ycl *YandexCloudLockbox) SetIndividualSecret(kvpath, secret, version, value string) error {
	log.Println("This functionality is not implemented for this backend")

	return nil
}

func (a *YandexCloudLockbox) GetSecret(kvpath, secretName string, annotations map[string]string) (map[string]interface{}, error) {
	log.Println("This functionality is not implemented for this backend")

	return nil, nil
}

func (a *YandexCloudLockbox) GetAllSecretsInPath(kvpath string, annotations map[string]string) (map[string]string, error) {
	return nil, nil
}
