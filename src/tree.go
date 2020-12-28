package kagi

import "errors"

type Tree struct {
	root  *BranchNode
	count int
}

const (
	KEY_NOT_FOUND      error = errors.New("key not found in database.")
	KEY_ALREADY_EXISTS error = errors.New("key already exists in database.")
)

func findRootNode(b []byte) {
}

func insertNode(key string, value string) error {
}

func findLeaf(key string) (string, error) {
}

func removeNode(key string) error {
}
