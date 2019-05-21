package stats

import (
	"github.com/shirou/gopsutil/load"
	"github.com/zkynetio/lynx/helpers"
)

type LoadDynamic struct {
	MIN1      float64
	MIN5      float64
	MIN15     float64
	ValueList []int64
}

func collectLoad(dp *DynamicPoint) {
	ld, err := load.Avg()
	helpers.PanicX(err)

	dp.LoadDynamic = &LoadDynamic{
		MIN1:  ld.Load1,
		MIN5:  ld.Load5,
		MIN15: ld.Load15,
	}
}
func (d *LoadDynamic) GetFormattedBytes(basePoint bool) []byte {
	base := History.DynamicBasePoint.LoadDynamic
	if basePoint {
		base.ValueList = append(base.ValueList, int64(base.MIN1))
		base.ValueList = append(base.ValueList, int64(base.MIN5))
		base.ValueList = append(base.ValueList, int64(base.MIN15))
		return helpers.WriteValueList(base.ValueList, "")
	}
	prev := History.DynamicPreviousUpdatePoint.LoadDynamic
	d.ValueList = append(d.ValueList, int64(d.MIN1))
	d.ValueList = append(d.ValueList, int64(d.MIN5))
	d.ValueList = append(d.ValueList, int64(d.MIN15))
	return helpers.WriteValueList2(d.ValueList, base.ValueList, prev.ValueList, "")
}
