package main

import (
	"fmt"
	"os"
	"os/exec"
)

const warningText = `
***************************************************************
*                  !!!!!!!! WARNING !!!!!!!!                  *
*  The current HEAD is the same commit with the last deploy!  *
***************************************************************
`

const LAST_HASH_PATH = "tmp/check-commit/last-hash"

func main() {
	err := run()
	if err != nil {
		return
	}
}

func run() error {
	tmpHash, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return err
	}

	if existFile(LAST_HASH_PATH) {
		lastHash, err := os.ReadFile(LAST_HASH_PATH)
		if err != nil {
			return err
		}

		if string(lastHash) == string(tmpHash) {
			fmt.Println(warningText)
		}
	}

	err = os.WriteFile(LAST_HASH_PATH, tmpHash, 0777)
	if err != nil {
		return err
	}

	return nil
}

func existFile(filePath string) bool {
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return false
	}

	return true
}
