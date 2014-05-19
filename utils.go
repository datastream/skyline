package skyline

// TimePoint is basic data struct
type TimePoint interface {
	GetTimestamp() int64 //x time
	GetValue() float64   //y value
}

// TimeArray return all timestamps in timeseries array
func TimeArray(timeseries []TimePoint) []int64 {
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.GetTimestamp())
	}
	return t
}

// ValueArray return all values in timeseries array
func ValueArray(timeseries []TimePoint) []float64 {
	var v []float64
	for _, val := range timeseries {
		v = append(v, val.GetValue())
	}
	return v
}

// TimeValueArray return all timestamps & values in timeseries array
func TimeValueArray(timeseries []TimePoint) ([]int64, []float64) {
	var v []float64
	var t []int64
	for _, val := range timeseries {
		t = append(t, val.GetTimestamp())
		v = append(v, val.GetValue())
	}
	return t, v
}
