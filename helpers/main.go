package helpers

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func DebugLog(v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(v...)
	}
}
func PanicX(err error) {
	if err != nil {
		panic(err)
	}
}

func LoadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}
}
