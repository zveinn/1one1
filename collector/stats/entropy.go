package stats

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/zkynetio/lynx/helpers"
)

type EntropyDynamic struct {
	Value     int
	ValueList []int64
}

func collectEntropy(dp *DynamicPoint) {
	file, err := ioutil.ReadFile("/proc/sys/kernel/random/entropy_avail") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	entInt, err := strconv.Atoi(strings.TrimSuffix(string(file), "\n"))
	helpers.PanicX(err)
	dp.EntropyDynamic = &EntropyDynamic{Value: entInt}
}

func (d *EntropyDynamic) GetFormattedBytes(basePoint bool) []byte {
	base := History.DynamicBasePoint.EntropyDynamic
	if basePoint {
		base.ValueList = append(base.ValueList, int64(base.Value))
		return helpers.WriteValueList(base.ValueList, "")
	}
	prev := History.DynamicPreviousUpdatePoint.EntropyDynamic
	d.ValueList = append(d.ValueList, int64(d.Value))
	return helpers.WriteValueList2(d.ValueList, base.ValueList, prev.ValueList, "")
}
