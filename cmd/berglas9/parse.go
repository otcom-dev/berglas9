package main

import (
	"errors"
	"strings"

	"github.com/docopt/docopt-go"
)

const usageDoc = `berglas9 — TUI for Berglas secrets.

Usage:
  berglas9 <bucket> --key=<kms-key>
  berglas9 -h | --help
  berglas9 -v | --version

Options:
  -h --help           Show this help.
  -v --version        Show version.
  --key=<kms-key>     Cloud KMS key used when writing secrets.`

var (
	errHelp    = errors.New("help")
	errVersion = errors.New("version")
	errUsage   = errors.New("bucket and --key are required")
)

type launchConfig struct {
	Bucket string
	Key    string
}

func parseLaunch(argv []string) (launchConfig, error) {
	askedHelp := false
	parser := &docopt.Parser{
		HelpHandler: func(err error, _ string) {
			if err == nil {
				askedHelp = true
			}
		},
	}
	opts, err := parser.ParseArgs(usageDoc, argv, "")
	if askedHelp {
		return launchConfig{}, errHelp
	}
	if err != nil {
		return launchConfig{}, errUsage
	}
	if help, _ := opts.Bool("--help"); help {
		return launchConfig{}, errHelp
	}
	if ver, _ := opts.Bool("--version"); ver {
		return launchConfig{}, errVersion
	}
	bucket, _ := opts.String("<bucket>")
	key, _ := opts.String("--key")
	bucket = normalizeBucket(bucket)
	key = strings.TrimSpace(key)
	if bucket == "" || key == "" {
		return launchConfig{}, errUsage
	}
	return launchConfig{Bucket: bucket, Key: key}, nil
}

func normalizeBucket(bucket string) string {
	bucket = strings.TrimSpace(bucket)
	bucket = strings.TrimPrefix(bucket, "gs://")
	return strings.Trim(bucket, "/")
}
