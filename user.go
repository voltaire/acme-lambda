package main

import (
	"context"
	"net/http"

	"golang.org/x/crypto/acme"
)

func registerUser(ctx context.Context, acmeClient *acme.Client, email string) error {
	_, err := acmeClient.Register(ctx, &acme.Account{
		Contact: []string{
			"mailto:" + email,
		},
	}, acme.AcceptTOS)

	if err == nil {
		return nil
	}

	if err == acme.ErrAccountAlreadyExists {
		return nil
	}

	ae, ok := err.(*acme.Error)
	if ok && ae.StatusCode == http.StatusConflict {
		return nil
	}

	return err
}
