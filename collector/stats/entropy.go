package stats

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/zkynetio/lynx/helpers"
)

type EntropyDynamic struct {
	Value int
}

func collectEntropy(dp *DynamicPoint) {
	file, err := ioutil.ReadFile("/proc/sys/kernel/random/entropy_avail") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	entInt, err := strconv.Atoi(strings.TrimSuffix(string(file), "\n"))
	helpers.PanicX(err)
	dp.EntropyDynamic = EntropyDynamic{Value: entInt}
}
func (ed *EntropyDynamic) GetFormattedString() string {
	var value string
	//log.Println("OLD ENTROPY", History.DynamicPointMap[HighestHistoryIndex-1].EntropyDynamic.Value)
	//log.Println("NEW ENT", ed.Value)
	if History.DynamicPointMap[HighestHistoryIndex-1].EntropyDynamic.Value != ed.Value {
		value = strconv.Itoa(ed.Value - History.DynamicPointMap[HighestHistoryIndex-1].EntropyDynamic.Value)
	} else {
		value = "x"
	}

	return value
}
