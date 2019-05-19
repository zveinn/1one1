package stats

import (
	"github.com/shirou/gopsutil/load"
	"github.com/zkynetio/lynx/helpers"
)

type LoadDynamic struct {
	MIN1  float64
	MIN5  float64
	MIN15 float64
}

func collectLoad(dp *DynamicPoint) {
	ld, err := load.Avg()
	helpers.PanicX(err)

	dp.LoadDynamic = LoadDynamic{
		MIN1:  ld.Load1,
		MIN5:  ld.Load5,
		MIN15: ld.Load15,
	}
}
func (d *LoadDynamic) GetFormattedBytes(basePoint bool) []byte {
	var valueList []int64
	base := History.DynamicBasePoint.LoadDynamic
	if basePoint {
		valueList = append(valueList, int64(base.MIN1))
		valueList = append(valueList, int64(base.MIN5))
		valueList = append(valueList, int64(base.MIN15))
	} else {
		valueList = append(valueList, int64(d.MIN1)-int64(base.MIN1))
		valueList = append(valueList, int64(d.MIN5)-int64(base.MIN5))
		valueList = append(valueList, int64(d.MIN15)-int64(base.MIN15))
	}
	return helpers.WriteValueList(valueList)
}
