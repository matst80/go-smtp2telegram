package server

import (
	"crypto/md5"
	"fmt"
)

type HashGenerator interface {
	CreateHash(key string) string
}

type SimpleHash struct {
	Salt string
}

func (h *SimpleHash) CreateHash(key string) string {
	md5 := md5.New()
	md5.Write([]byte(fmt.Sprintf("%s%s", key, h.Salt)))
	return fmt.Sprintf("%x", md5.Sum(nil))
}
