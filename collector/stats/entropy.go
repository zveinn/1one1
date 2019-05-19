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

func (d *EntropyDynamic) GetFormattedBytes(basePoint bool) []byte {
	var valueList []int64
	base := History.DynamicBasePoint.EntropyDynamic
	if basePoint {
		valueList = append(valueList, int64(base.Value))
	} else {
		valueList = append(valueList, int64(d.Value)-int64(base.Value))
	}
	return helpers.WriteValueList(valueList, "")
}
