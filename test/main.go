package main

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
)

func main() {
	log.Println(string([]byte{0}))
	fmt.Println(1)
	log.Println(string([]byte{1}))
	fmt.Println(1)
	log.Println(string([]byte{2}))
	fmt.Println(1)
	log.Println(string([]byte{3}))
	fmt.Println(1)
	log.Println(string([]byte{4}))
	fmt.Println(1)
	log.Println(string([]byte{5}))
	fmt.Println(1)
	log.Println(string([]byte{6}))
	fmt.Println(1)
	log.Println(string([]byte{7}))
	fmt.Println(1)
	log.Println(string([]byte{8}))
	//fmt.Println(1)
	//log.Println(string([]byte{9}))
	//fmt.Println(1)
	//log.Println(string([]byte{10}))
	//fmt.Println(1)
	//log.Println(string([]byte{11}))
	//fmt.Println(1)

	x := "enp0s31f6,192.168.0.137/24,fe80::d202:8f46:38ae:eb05/64,224.0.0.251,224.0.0.1,ff02::1:ffae:eb05,ff02::1,ff01::1,,e8:6a:64:78:5f:b9,19,2,1500,"
	log.Println([]byte(x))
	log.Println(len(x))

	//buf := new(bytes.Buffer)
	xx1 := big.NewInt(10000000001)
	xx1.Add(xx1, big.NewInt(12000000001))
	xx1.Add(xx1, big.NewInt(10300000001))
	xx1.Add(xx1, big.NewInt(10040000001))
	xx1.Add(xx1, big.NewInt(10005000001))
	xx1.Add(xx1, big.NewInt(10000600001))
	xx1.Add(xx1, big.NewInt(10000070001))
	xx1.Add(xx1, big.NewInt(10000008001))
	xx1.Add(xx1, big.NewInt(10000000901))
	xx1.Add(xx1, big.NewInt(10000000001))
	xx1.Add(xx1, big.NewInt(33333310000000001))
	log.Println(xx1.Bytes())
	log.Println(len(xx1.Bytes()))
	//if err != nil {
	//	fmt.Println("binary.Write failed:", err)
	//}
	//log.Println(buf.Bytes())
	log.Println(xx1.String())
	//number := uint64(0)
	//err = binary.Read(buf, binary.LittleEndian, &number)

	//	if err != nil {
	//log.Fatalf("Decode failed: %s", err)
	//	}

	var final []int
	xxx := []byte{1, 22, 121, 001}
	for _, v := range xxx {
		final = append(final, int(v))
	}
	log.Println(final)

	//	log.Println(join(final))

	//xxx2 := []byte{123, 4, 56, 78, 9}
	//	xxx3 := []byte{98, 7, 65, 43, 210}
	//xxx4 := []byte{111, 122, 223, 33, 0, 34, 44, 45, 55, 56, 66, 67}

	//for i, v := range xxx2 {
	//	log.Println(int(v) - int(xxx[i]))
	//}
	//transform(xxx2)
	//transform(xxx4)
	//transform(xxx3)
}

func transform(thing []byte) {
	bigInt := big.NewInt(0)
	for _, vb := range thing {
		log.Println("vb", vb)
		v := int(vb)
		slot1 := 99
		slot2 := 99
		slot3 := 99
		log.Println("number:", v)
		if v < 10 {
			slot1 = (v / 1) % 10
		} else if v < 100 {
			slot1 = (v / 10) % 10
			slot2 = (v / 1) % 10
		} else {
			slot1 = (v / 100) % 10
			slot2 = (v / 10) % 10
			slot3 = (v / 1) % 10
		}
		if slot1 != 99 {
			log.Println("slot1:", slot1)
			bigInt.Mul(bigInt, big.NewInt(10))
			bigInt.Add(bigInt, big.NewInt(int64(slot1)))
		}
		if slot2 != 99 {
			log.Println("slot2:", slot2)
			bigInt.Mul(bigInt, big.NewInt(10))
			bigInt.Add(bigInt, big.NewInt(int64(slot2)))
		}

		if slot3 != 99 {
			bigInt.Mul(bigInt, big.NewInt(10))
			bigInt.Add(bigInt, big.NewInt(int64(slot3)))
			log.Println("slot3:", slot3)
		}
		log.Println("final post add", bigInt)
	}
	log.Println(bigInt)
}

func join(nums []int) (int, error) {
	var str string
	for i := range nums {
		str += strconv.Itoa(nums[i])
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return num, nil
}
