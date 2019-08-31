package helpers

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
)

func DebugLog(v ...interface{}) {
	// if os.Getenv("DEBUG") == "true" {
	log.Println(v...)
	// }
}
func PanicX(err error) {
	if err != nil {
		panic(err)
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
	for i, v := range valueList {
		base = base + (v - preValueList[i])
	}
	if batchTag != "" && len(batchTag) < 255 && base != 0 {
		headers = append(headers, byte(len(batchTag)))
		data = append(data, []byte(batchTag)...)
	}
	// log.Println("value list", valueList, "batch tag:", batchTag)
	for i, v := range valueList {

		var length int

		if v == preValueList[i] {
			continue
		}

		if v == baseValueList[i] {
			length = 1
		} else {

			value := v - baseValueList[i]

			if value < 0 {
				value = abs(value)
				data = append(data, byte(0))
			} else {
				data = append(data, byte(1))
			}
			// log.Println(v, baseValueList[i])
			// log.Println(value)
			length = WriteIntToBuffer(&buffer, value)
			length = length + 1
			data = append(data, buffer.Bytes()...)
		}

		final, err := strconv.Atoi(strconv.Itoa(i) + strconv.Itoa(length))
		if err != nil {
			panic(err)
		}
		headers = append(headers, byte(final))

		buffer.Reset()
	}
	dataAndHeader = append(dataAndHeader, byte(len(headers)))
	dataAndHeader = append(dataAndHeader, headers...)
	dataAndHeader = append(dataAndHeader, data...)
	// debug
	// log.Println(batchTag)
	// for i, v := range valueList {
	// 	fmt.Print("{", i, "}", v-baseValueList[i], "{", v-preValueList[i], "}", " .. ")
	// }
	// fmt.Println()
	// log.Println(baseValueList)
	// log.Println(valueList)

	// log.Println("formatted bytes", dataAndHeader)
	return dataAndHeader
}
func abs(n int64) int64 {
	y := n >> 63
	return ((n * y) - y) - 1
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
			v = abs(v)
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
