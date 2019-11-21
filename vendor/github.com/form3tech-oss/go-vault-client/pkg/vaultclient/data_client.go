package vaultclient

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

var vaultClient *api.Client

func Configure(config *Config) error {
	vaultClientFactory, err := NewVaultAuth(config)

	if err != nil {
		return errors.Wrapf(err, "creating vault client factory")
	}
	client, err := vaultClientFactory.VaultClient()
	if err != nil {
		return errors.Wrapf(err, "creating vault client")
	}
	vaultClient = client
	return nil
}

func ConfigureDefault() error {
	config := NewDefaultConfig()
	if config.Error != nil {
		return errors.Wrapf(config.Error, "configuring vault client")
	}

	return Configure(config)
}

// Onus is on the caller to make sure the client has been configured
func GetClient() *api.Client {
	return vaultClient
}

func Read(path string) (*api.Secret, error) {
	return vaultClient.Logical().Read(path)
}

// ReadData returns the Data held in the Secret, use Read if you need metadata
func ReadData(path string) (map[string]interface{}, error) {
	data, err := Read(path)

	// Logical operations can legitimately return nil, nil
	if data == nil && err == nil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "vault error - fail to perform read operation on path '%s'", path)
	}

	return data.Data, nil
}

func Write(path string, data map[string]interface{}) (*api.Secret, error) {
	return vaultClient.Logical().Write(path, data)
}

// WriteData returns the Data held in the Secret, use Write if you need metadata
func WriteData(path string, data map[string]interface{}) (map[string]interface{}, error) {
	secret, err := Write(path, data)

	// Logical operations can legitimately return nil, nil
	if secret == nil && err == nil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "vault error - fail to perform write operation on path '%s'", path)
	}

	return secret.Data, nil
}

func List(path string) (*api.Secret, error) {
	return vaultClient.Logical().List(path)
}

// ListData returns the Data held in the Secrets, use List if you need metadata
func ListData(path string) ([]interface{}, error) {
	secret, err := List(path)

	// Logical operations can legitimately return nil, nil
	if secret == nil && err == nil {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "vault error - fail to perform list operation on path '%s'", path)
	}

	rawKeys, found := secret.Data["keys"]
	if !found {
		return nil, fmt.Errorf("vault error - the path '%v' doesn't contain any keys", path)
	}

	keys, ok := rawKeys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("vault error - the path '%v' doesn't contain a valid list of keys", path)
	}

	return keys, nil
}

func Delete(path string) (*api.Secret, error) {
	return vaultClient.Logical().Delete(path)
}

// DeleteData returns the Data held in the Secret, use Delete if you need metadata
func DeleteData(path string) (map[string]interface{}, error) {
	secret, err := Delete(path)

	// Logical operations can legitimately return nil, nil
	if secret == nil && err == nil {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("vault error - fail to perform delete operation on path '%s'", path)
	}

	return secret.Data, nil
}
