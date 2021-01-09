package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/go-acme/lego/v4/registration"
)

type userObject struct {
	Email      string
	Resource   *registration.Resource
	privateKey *ecdsa.PrivateKey
}

func (p *userObject) GetEmail() string {
	return p.Email
}

func (p *userObject) GetRegistration() *registration.Resource {
	return p.Resource
}

func (p *userObject) GetPrivateKey() crypto.PrivateKey {
	return p.privateKey
}

func loadOrCreateUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*userObject, error) {
	u, err := loadUser(ctx, sm, email)
	if isKeyNotFoundError(err) {
		return createUser(ctx, email)
	}

	return u, err
}

func secretsManagerUserName(email string) string {
	return "mapcert-user-" + email
}

func secretsManagerPrivateKey(email string) string {
	return "mapcert-privkey-" + email
}

func isKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	awsErr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if awsErr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
		return true
	}
	return false
}

func createUser(ctx context.Context, email string) (*userObject, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &userObject{
		Email:      email,
		privateKey: key,
	}, nil
}

func saveUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, u *userObject) error {
	userJson, err := json.Marshal(u)
	if err != nil {
		return err
	}

	keyDer, err := x509.MarshalECPrivateKey(u.privateKey)
	if err != nil {
		return err
	}
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDer,
	})

	_, err = sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Description:  aws.String("goacme/lego private key for: " + u.Email),
		Name:         aws.String(secretsManagerPrivateKey(u.Email)),
		SecretBinary: privateKeyPem,
	})
	if err != nil {
		return err
	}

	_, err = sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Description:  aws.String("goacme/lego user for: " + u.Email),
		Name:         aws.String(secretsManagerUserName(u.Email)),
		SecretBinary: userJson,
	})
	return err
}

func loadUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*userObject, error) {
	userOut, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerUserName(email)),
	})
	if err != nil {
		return nil, err
	}

	privateKeyOut, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerPrivateKey(email)),
	})
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyOut.SecretBinary)
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	var u userObject
	err = json.Unmarshal(userOut.SecretBinary, &u)
	if err != nil {
		return nil, err
	}

	u.privateKey = privateKey

	return &u, nil
}
