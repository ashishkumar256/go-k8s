package main

import (
	"fmt"
	"hello-world/utils"
)

func main() {
	fmt.Println("This is from example.go file")

	// Accessing a variable from the imported package
	fmt.Println(utils.MyVariable)

	// Calling a function from the imported package
	utils.MyFunction()
}
