package main

import (
	"crypto/md5"
	"fmt"
)

type HashGenerator interface {
	CreateHash(key string) string
}

type hash struct {
	salt string
}

func (h *hash) CreateHash(key string) string {
	md5 := md5.New()
	md5.Write([]byte(fmt.Sprintf("%s%s", key, h.salt)))
	return fmt.Sprintf("%x", md5.Sum(nil))
}
