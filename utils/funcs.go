/*
@Time : 2022/12/1 23:25
@Author : lianyz
@Description :
*/

package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// GetRunPath 获取程序执行目录
func GetRunPath() (string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return path, err
}

func LogHandler(handler string, info string) {
	log.Println("[Handler]: " + handler + ". " + info)
}

func LogHandlerF(handler string, format string, a ...any) {
	log.Println("[Handler]: " + handler + ". " +
		fmt.Sprintf(format, a))
}

func LogRequest(req string, r *http.Request) {
	decodedURL, _ := url.QueryUnescape(r.URL.String())
	log.Println("[Request]: " + req + ". method: " + r.Method + " url: " + decodedURL)
}

func Log(req string, format string, a ...any) {
	log.Println("[Request]: " + req + ". " +
		fmt.Sprintf(format, a))
}

func LogRedirect(req string, url string) {
	log.Println("[Request]: " + req + ". redirect to " + url)
}
