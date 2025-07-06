package localutils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"golang.org/x/text/encoding/unicode"
)

func HashSum(login string, password string) string {
	s := fmt.Sprint(login, password)
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	s1, _ := encoder.String(s)
	buff := []byte(s1)
	sum := md5.Sum(buff)
	hash := base64.StdEncoding.EncodeToString(sum[:])
	return hash
}
