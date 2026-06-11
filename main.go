package main
import (
	"os"
	"fmt"
)

func main() {
	client, err := NewPodman()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(client)

	RunSocket(client)
}