package stats

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/load"
	"github.com/zkynetio/lynx/helpers"
)

type Load struct {
	MIN1  float64
	MIN5  float64
	MIN15 float64
}

func collectLoad(dp *DataPoint) {
	ld, err := load.Avg()
	helpers.PanicX(err)

	dp.Load = Load{
		MIN1:  ld.Load1,
		MIN5:  ld.Load5,
		MIN15: ld.Load15,
	}
}

func (l *Load) GetFormattedString() string {
	var loadString []string

	if History.DataPointMap[HighestHistoryIndex-1].Load.MIN1 != l.MIN1 {
		loadString = append(loadString, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Load.MIN1-l.MIN1)))
	} else {
		loadString = append(loadString, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Load.MIN5 != l.MIN5 {
		loadString = append(loadString, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Load.MIN5-l.MIN5)))
	} else {
		loadString = append(loadString, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Load.MIN15 != l.MIN15 {
		loadString = append(loadString, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Load.MIN15-l.MIN15)))
	} else {
		loadString = append(loadString, "")
	}

	return strings.Join(loadString, ",")
}
