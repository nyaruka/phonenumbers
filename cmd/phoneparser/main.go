package main

import (
	"fmt"
	"os"
)

func main() {
	num, err := phonenumbers.Parse(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Printf("Error parsing number: %s\n", err)
	}
	fmt.Printf("Parsed to: %s\n", phonenumbers.Format(num, phonenumbers.E164))
}
