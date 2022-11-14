package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

type PGPSigningKey struct {
	KeyID      string
	ASCIIArmor string
}

func GetPublicSigningKey(pgpID string) (*PGPSigningKey, error) {
	cmd := exec.Command("gpg", "--armor", "--export", pgpID)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		msg, err := ioutil.ReadAll(stderr)
		if err != nil {
			return nil, err

		}
		return nil, fmt.Errorf("failed to retrieve public key %s, %s", pgpID, string(msg))
	}
	return &PGPSigningKey{pgpID, string(key)}, nil
}
