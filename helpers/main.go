package helpers

import (
	"bytes"
	"encoding/binary"
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
func writeTo(buf *bytes.Buffer, value interface{}) {
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		panic(err)
	}

}
func WriteIntToBuffer(buf *bytes.Buffer, value int64) byte {
	if value > -129 && value < 128 {
		log.Println("writing int8", value)
		writeTo(buf, int8(value))
		return 0x1
	} else if value > -32769 && value < 32768 {
		writeTo(buf, int16(value))

		log.Println("writing int16", value)
		return 0x2
	} else if value > -2147483649 && value < 2147483648 {
		log.Println("writing int32", value)
		writeTo(buf, int32(value))
		return 0x4
	}
	writeTo(buf, value)
	return 0x8
}
