package util

import (
	"fmt"
	"os"
)

// 環境変数を読み込む。値がなければエラーを返す
func MustGetenv(k string) (string, error) {
	v := os.Getenv(k)
	if v == "" {
		return "", fmt.Errorf("%s environment variable not set", k)
	}

	return v, nil
}
