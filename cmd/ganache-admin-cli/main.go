package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var version = "dev"

func main() {
	fmt.Printf("ganache-admin-cli %s\n", version)
	if len(os.Args) < 2 {
		fmt.Println("usage: ganache-admin-cli [hashpw|verify <hash>]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "hashpw":
		hashpw()
	case "verify":
		if len(os.Args) < 3 {
			fmt.Println("verify requires hash argument")
			os.Exit(1)
		}
		verify(os.Args[2])
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}

func hashpw() {
	password := readPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}

func verify(hash string) {
	password := readPassword()
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		fmt.Println("invalid")
		os.Exit(1)
	}
	fmt.Println("ok")
}

func readPassword() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
