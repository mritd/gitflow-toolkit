package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mritd/promptx"
)

func main() {
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		} else {
			return nil
		}
	}, "Please input:")

	fmt.Println(p.Run())
}
