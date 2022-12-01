/*
@Time : 2022/12/1 23:25
@Author : lianyz
@Description :
*/

package utils

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
)

// GetRunPath 获取程序执行目录
func GetRunPath() (string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return path, err
}

func LogRequest(req string, uri *url.URL) {
	decodedURL, _ := url.QueryUnescape(uri.String())
	log.Println("[Request]: " + req + " " + decodedURL)
}
