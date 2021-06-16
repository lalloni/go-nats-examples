package helpers

import (
	"encoding/ascii85"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
)

const (
	hexPrefix = "hex:"
	b64Prefix = "b64:"
	a85Prefix = "a85:"
)

func Bytes(arg string) (bool, []byte, error) {
	if strings.HasPrefix(arg, "@") {
		bs, err := os.ReadFile(strings.TrimPrefix(arg, "@"))
		return false, bs, err
	}
	if strings.HasPrefix(arg, hexPrefix) {
		bs, err := hex.DecodeString(strings.TrimPrefix(arg, hexPrefix))
		return false, bs, err
	}
	if strings.HasPrefix(arg, b64Prefix) {
		bs, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(arg, b64Prefix))
		return false, bs, err
	}
	if strings.HasPrefix(arg, a85Prefix) {
		var bs []byte
		_, err := ascii85.NewDecoder(strings.NewReader(strings.TrimPrefix(arg, a85Prefix))).Read(bs)
		return false, bs, err
	}
	return true, []byte(arg), nil
}
