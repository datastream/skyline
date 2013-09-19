package skyline

type TimePoint struct {
	Timestamp int64   //x time
	Value     float64 //y value
}

func TimeArray(timeseries []TimePoint) []int64 {
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.Timestamp)
	}
	return t
}

func ValueArray(timeseries []TimePoint) []float64 {
	var v []float64
	for _, val := range timeseries {
		v = append(v, val.Value)
	}
	return v
}

func TimeValueArray(timeseries []TimePoint) ([]int64, []float64) {
	var v []float64
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.Timestamp)
		v = append(v, val.Value)
	}
	return t, v
}
