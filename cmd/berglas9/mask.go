package main

import (
	"context"
	"errors"
	"strings"
)

// maskedSecret is a fixed mask so the display does not reveal the value's length.
const maskedSecret = "••••••••"

var errBlankField = errors.New("name and value are required")

func displayValue(plaintext []byte, revealed bool) string {
	if !revealed {
		return maskedSecret
	}
	return string(plaintext)
}

func copiedStatus(name string) string {
	return "Copied " + name
}

func buildCreate(name, value string) (string, []byte, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.TrimSpace(value) == "" {
		return "", nil, errBlankField
	}
	return name, []byte(value), nil
}

func submitCreate(ctx context.Context, store secretStore, name, value string) (string, error) {
	object, plaintext, err := buildCreate(name, value)
	if err != nil {
		return "", err
	}
	if err := store.Create(ctx, object, plaintext); err != nil {
		return "", err
	}
	return object, nil
}

func submitUpdate(ctx context.Context, store secretStore, ref secretRef, value string) error {
	if strings.TrimSpace(value) == "" {
		return errBlankField
	}
	return store.Update(ctx, ref, []byte(value))
}
