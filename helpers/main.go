package helpers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"

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
func WriteIntToBuffer(buf *bytes.Buffer, value int64) int {
	if value > -129 && value < 128 {
		// log.Println("writing int16", value)
		writeTo(buf, int16(value))
		return 2
	} else if value > -32769 && value < 32768 {
		writeTo(buf, int16(value))

		// log.Println("writing int16", value)
		return 2
	} else if value > -2147483649 && value < 2147483648 {
		// log.Println("writing int32", value)
		writeTo(buf, int32(value))
		return 4
	}
	writeTo(buf, value)
	return 8

}

func WriteValueList2(valueList, baseValueList, preValueList []int64, batchTag string) []byte {
	var buffer bytes.Buffer
	var data []byte
	var headers []byte
	var dataAndHeader []byte
	var base int64 = 0
	// log.Println(baseValueList)
	// log.Println(valueList)
	for _, v := range valueList {
		base = base + v
	}
	if batchTag != "" && len(batchTag) < 255 && base != 0 {
		headers = append(headers, byte(len(batchTag)))
		data = append(data, []byte(batchTag)...)
	}
	// log.Println("value list", valueList, "batch tag:", batchTag)
	for i, v := range valueList {

		var length int
		// if there is no change from the base value list we don't need to
		// send any data, just a control byte indicator
		if v == 0 {
			continue
		} else if v == baseValueList[i] {
			length = 1
		} else if v == preValueList[i] {
			continue
		} else {

			length = WriteIntToBuffer(&buffer, v-baseValueList[i])
			length = length + 1
			if v < 0 {
				data = append(data, byte(0))
			} else {
				data = append(data, byte(1))
			}
		}

		final, err := strconv.Atoi(strconv.Itoa(i) + strconv.Itoa(length))
		if err != nil {
			panic(err)
		}
		headers = append(headers, byte(final))
		if length != 1 {
			data = append(data, buffer.Bytes()...)
		}

		buffer.Reset()
	}
	dataAndHeader = append(dataAndHeader, byte(len(headers)))
	dataAndHeader = append(dataAndHeader, headers...)
	dataAndHeader = append(dataAndHeader, data...)
	log.Println(baseValueList)
	log.Println(valueList)
	log.Println(batchTag)
	// log.Println("formatted bytes", dataAndHeader)
	return dataAndHeader
}
func WriteValueList(valueList []int64, batchTag string) []byte {
	var buffer bytes.Buffer
	var data []byte
	var headers []byte
	var dataAndHeader []byte
	var base int64 = 0
	// log.Println(valueList)
	for _, v := range valueList {
		base = base + v
	}
	if batchTag != "" && len(batchTag) < 255 && base != 0 {
		headers = append(headers, byte(len(batchTag)))
		data = append(data, []byte(batchTag)...)
	}
	// log.Println("value list", valueList, "batch tag:", batchTag)
	for i, v := range valueList {
		if v < 0 {
			data = append(data, byte(0))
		} else if v == 0 {
			// headers = append(headers, []byte{0}...)
			continue
		} else {
			data = append(data, byte(1))
		}
		length := WriteIntToBuffer(&buffer, v)
		final, err := strconv.Atoi(strconv.Itoa(i) + strconv.Itoa(length+1))
		if err != nil {
			panic(err)
		}
		headers = append(headers, byte(final))
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	dataAndHeader = append(dataAndHeader, byte(len(headers)))
	dataAndHeader = append(dataAndHeader, headers...)
	dataAndHeader = append(dataAndHeader, data...)
	// log.Println("formatted bytes", dataAndHeader)
	return dataAndHeader
}

// func WriteNetworkValueList(valueList []int64, interfaceName string) []byte {
// 	var buffer bytes.Buffer
// 	var data []byte
// 	var headers []byte
// 	var dataAndHeader []byte
// 	for _, v := range valueList {
// 		if v < 0 {
// 			data = append(data, []byte{0}...)
// 		} else if v == 0 {
// 			headers = append(headers, []byte{0}...)
// 			continue
// 		} else {
// 			data = append(data, []byte{1}...)
// 		}
// 		length := WriteIntToBuffer(&buffer, v)
// 		headers = append(headers, length+1)
// 		data = append(data, buffer.Bytes()...)
// 		buffer.Reset()
// 	}
// 	dataAndHeader = append(dataAndHeader, headers...)
// 	dataAndHeader = append(dataAndHeader, data...)
// 	log.Println("formatted bytes", dataAndHeader)
// 	return dataAndHeader
// }
