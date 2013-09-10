package skyline

type TimePoint struct {
	timestamp int64    //x time
	value     float64  //y value
}

func TimeArray(timeseries []TimePoint) []int64 {
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.timestamp)
	}
	return t
}

func ValueArray(timeseries []TimePoint) []float64 {
	var v []float64
	for _, val := range timeseries {
		v = append(v, val.value)
	}
	return v
}

func TimeValueArray(timeseries []TimePoint) ([]int64, []float64) {
	var v []float64
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.timestamp)
		v = append(v, val.value)
	}
	return t, v
}

const (
	FULL_DURATION = 1
)
