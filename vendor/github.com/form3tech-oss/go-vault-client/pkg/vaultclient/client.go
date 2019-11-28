package vaultclient

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/vault/api"
	awsauth "github.com/hashicorp/vault/builtin/credential/aws"
)

type AuthType int

const (
	Token AuthType = iota + 1
	Iam
	AppRole
)

type iamAuth struct {
	role   string
	client *api.Client
	auth   *Auth
}

type tokenAuth struct {
	client *api.Client
}

type appRoleAuth struct {
	auth     *Auth
	client   *api.Client
	role     string
	roleId   string
	secretId string
}

type Config struct {
	*api.Config
	AuthType        AuthType
	Token           string
	IamRole         string
	AppRole         string
	AppRoleId       string
	AppRoleSecretId string
}

type Auth struct {
	token  string
	expiry time.Time
}

var (
	expirationWindow = time.Second * 10
)

type VaultAuth interface {
	VaultClient() (*api.Client, error)
	VaultClientOrPanic() *api.Client
}

func BaseConfig() *Config {
	apiConfig := api.DefaultConfig()

	config := &Config{
		Config: apiConfig,
	}

	return config
}

func NewDefaultConfig() *Config {
	config := BaseConfig()

	appRoleName := os.Getenv("VAULT_APP_ROLE")
	appRoleId := os.Getenv("VAULT_APP_ROLE_ID")
	appRoleSecretId := os.Getenv("VAULT_APP_SECRET_ID")
	if appRoleId != "" && appRoleSecretId != "" && appRoleName != "" {
		config.AuthType = AppRole
		config.AppRole = appRoleName
		config.AppRoleId = appRoleId
		config.AppRoleSecretId = appRoleSecretId

		return config
	}

	role := os.Getenv("VAULT_ROLE")
	if role != "" {
		config.AuthType = Iam
		config.IamRole = role

		return config
	}

	token := os.Getenv("VAULT_TOKEN")
	if token != "" {
		config.AuthType = Token
		config.Token = token

		return config
	}

	config.Error = fmt.Errorf("failed to determine auth type from env")
	return config
}

func NewVaultAuth(cfg *Config) (VaultAuth, error) {
	c, err := api.NewClient(cfg.Config)
	if err != nil {
		return nil, err
	}

	switch cfg.AuthType {
	case Token:
		c.SetToken(cfg.Token)
		return &tokenAuth{
			client: c,
		}, nil
	case AppRole:
		return &appRoleAuth{
			client:   c,
			role:     cfg.AppRole,
			secretId: cfg.AppRoleSecretId,
			roleId:   cfg.AppRoleId,
		}, nil
	case Iam:
		return &iamAuth{
			client: c,
			role:   cfg.IamRole,
		}, nil

	}
	return nil, fmt.Errorf("unkown auth type %s", string(cfg.AuthType))
}

func (v *iamAuth) getAuth() (*Auth, error) {
	s := session.Must(session.NewSession())
	data, err := awsauth.GenerateLoginData(s.Config.Credentials, "")
	if err != nil {
		return nil, nil
	}

	data["role"] = v.role

	resp, err := v.client.Logical().Write("auth/aws/login", data)
	if err != nil {
		return nil, err
	}

	tokenTtl, err := resp.TokenTTL()
	if err != nil {
		return nil, err
	}

	return &Auth{
		token:  resp.Auth.ClientToken,
		expiry: time.Now().UTC().Add(tokenTtl),
	}, nil
}

func (v *Auth) IsTokenExpired() bool {
	if v == nil {
		return true
	}

	return v.expiry.Before(time.Now().Add(expirationWindow).UTC())
}

func (v *iamAuth) VaultClient() (*api.Client, error) {
	if !v.auth.IsTokenExpired() {
		return v.client, nil
	}
	var err error
	v.auth, err = v.getAuth()
	if err != nil {
		return nil, err
	}
	v.client.SetToken(v.auth.token)
	return v.client, nil
}

func (v *iamAuth) VaultClientOrPanic() *api.Client {
	client, err := v.VaultClient()
	if err != nil {
		panic(err)
	}
	return client
}

func (t *tokenAuth) VaultClient() (*api.Client, error) {
	return t.client, nil
}

func (t *tokenAuth) VaultClientOrPanic() *api.Client {
	client, err := t.VaultClient()
	if err != nil {
		panic(err)
	}
	return client
}

func (a *appRoleAuth) getAuth() (*Auth, error) {
	data := map[string]interface{}{
		"role_id":   a.roleId,
		"secret_id": a.secretId,
	}

	resp, err := a.client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return nil, err
	}

	tokenTtl, err := resp.TokenTTL()
	if err != nil {
		return nil, err
	}

	return &Auth{
		token:  resp.Auth.ClientToken,
		expiry: time.Now().UTC().Add(tokenTtl),
	}, nil
}

func (a *appRoleAuth) VaultClient() (*api.Client, error) {
	if !a.auth.IsTokenExpired() {
		return a.client, nil
	}

	var err error
	a.auth, err = a.getAuth()
	if err != nil {
		return nil, err
	}
	a.client.SetToken(a.auth.token)
	return a.client, nil
}

func (a *appRoleAuth) VaultClientOrPanic() *api.Client {
	client, err := a.VaultClient()
	if err != nil {
		panic(err)
	}
	return client
}
