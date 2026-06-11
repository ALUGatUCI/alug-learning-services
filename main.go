package main
import (
	"os"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables first
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Failed to load environment variables: %s\n", err)
		os.Exit(1)
	}

	client, err := NewPodman()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(client)

	RunSocket(client)
}