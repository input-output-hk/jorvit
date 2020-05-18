package kit

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FatalOn be careful with it in production,
// since it uses os.Exit(1) which affects the control flow.
// use pattern:
// if err != nil {
// 	....
// }
func FatalOn(err error, str ...string) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Fatalf("%s:%d %s -> %s", fn, line, str, err.Error())
	}
}

// B2S converts []byte to string with all leading
// and trailing white space removed, as defined by Unicode.
func B2S(b []byte) string {
	return strings.TrimSpace(string(b))
}

// FindExecutable starting from `dir` and then PATH env
func FindExecutable(fileName string, dir string) (string, error) {
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		return dirPath, err
	}

	bin, err := exec.LookPath(dir + string(os.PathSeparator) + fileName)
	if err != nil {
		bin, err = exec.LookPath(fileName)
		if err != nil {
			return "", fmt.Errorf("%s binary not found in PATH or %s", fileName, dirPath)
		}
	}
	bin, err = filepath.Abs(bin)
	return bin, err
}
