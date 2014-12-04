package skyline

import (
	"github.com/dgryski/go-onlinestats"
	"math"
	"sort"
)

func unDef(f float64) bool {
	if math.IsNaN(f) {
		return true
	}
	if math.IsInf(f, 1) {
		return true
	}
	if math.IsInf(f, -1) {
		return true
	}
	return false
}

// Mean Average
// mean = SIGMA a / len(a)
// SIGMA a total of all of the elements of a
// len(a) = length of a (aka the number of values)
func Mean(a []float64) float64 {
	s := onlinestats.NewRunning()
	for _, v := range a {
		s.Push(v)
	}
	return s.Mean()
}

// Median series.median
func Median(series []float64) float64 {
	var median float64
	sort.Float64s(series)
	Len := len(series)
	lhs := (Len - 1) / 2
	rhs := Len / 2
	if Len == 0 {
		return 0.0
	}
	if lhs == rhs {
		median = series[lhs]
	} else {
		median = (series[lhs] + series[rhs]) / 2.0
	}
	return median
}

// Std Standard Deviation of a sample
// sd = sqrt(SIGMA ((a[i] - mean) ^ 2) / (len(a)-1))
// SIGMA a total of all of the elements of a
// a[i] is the ith elemant of a
// len(a) = the number of elements in the slice a adjusted for sample
func Std(a []float64) float64 {
	s := onlinestats.NewRunning()
	for _, v := range a {
		s.Push(v)
	}
	return s.Stddev()
}

// LinearRegressionLSE least squares linear regression
func LinearRegressionLSE(timeseries []TimePoint) (float64, float64) {
	regression := onlinestats.NewRegression()
	for _, p := range timeseries {
		regression.Push(float64(p.GetTimestamp()), p.GetValue())
	}
	return regression.Slope(), regression.Intercept()
}

// Ewma
func Ewma(series []float64, com float64) []float64 {
	var cur float64
	var prev float64
	var oldw float64
	var adj float64
	N := len(series)
	ret := make([]float64, N)
	if N == 0 {
		return ret
	}
	oldw = com / (1 + com)
	adj = oldw
	ret[0] = series[0] / (1 + com)
	for i := 1; i < N; i++ {
		cur = series[i]
		prev = ret[i-1]
		if unDef(cur) {
			ret[i] = prev
		} else {
			if unDef(prev) {
				ret[i] = cur / (1 + com)
			} else {
				ret[i] = (com*prev + cur) / (1 + com)
			}
		}
	}
	for i := 0; i < N; i++ {
		cur = ret[i]
		if !math.IsNaN(cur) {
			ret[i] = ret[i] / (1. - adj)
			adj *= oldw
		} else {
			if i > 0 {
				ret[i] = ret[i-1]
			}
		}
	}
	return ret
}

// EwmStd Exponentially-weighted moving std
func EwmStd(series []float64, com float64) []float64 {
	m1st := Ewma(series, com)
	var series2 []float64
	for _, val := range series {
		series2 = append(series2, val*val)
	}
	m2nd := Ewma(series2, com)
	l := len(m1st)
	var result []float64
	for i := 0; i < l; i++ {
		t := m2nd[i] - math.Pow(m1st[i], 2)
		t *= (1.0 + 2.0*com) / (2.0 * com)
		result = append(result, math.Sqrt(t))
	}
	return result
}

// Histogram numpy.histogram
func Histogram(series []float64, bins int) ([]int, []float64) {
	var binEdges []float64
	var hist []int
	l := len(series)
	if l == 0 {
		return hist, binEdges
	}
	sort.Float64s(series)
	w := (series[l-1] - series[0]) / float64(bins)
	for i := 0; i < bins; i++ {
		binEdges = append(binEdges, w*float64(i)+series[0])
		if binEdges[len(binEdges)-1] >= series[l-1] {
			break
		}
	}
	binEdges = append(binEdges, w*float64(bins)+series[0])
	bl := len(binEdges)
	hist = make([]int, bl-1)
	for i := 0; i < bl-1; i++ {
		for _, val := range series {
			if val >= binEdges[i] && val < binEdges[i+1] {
				hist[i] += 1
				continue
			}
			if i == (bl-2) && val >= binEdges[i] && val <= binEdges[i+1] {
				hist[i] += 1
			}
		}
	}
	return hist, binEdges
}

// KS2Samp
func KS2Samp(data1, data2 []float64) (float64, float64) {
	sort.Float64s(data1)
	sort.Float64s(data2)
	n1 := len(data1)
	n2 := len(data2)
	var dataAll []float64
	dataAll = append(dataAll, data1...)
	dataAll = append(dataAll, data2...)
	index1 := searchsorted(data1, dataAll)
	index2 := searchsorted(data2, dataAll)
	var cdf1 []float64
	var cdf2 []float64
	for _, v := range index1 {
		cdf1 = append(cdf1, float64(v)/float64(n1))
	}
	for _, v := range index2 {
		cdf2 = append(cdf2, float64(v)/float64(n2))
	}
	d := 0.0
	for i := 0; i < len(cdf1); i++ {
		d = math.Max(d, math.Abs(cdf1[i]-cdf2[i]))
	}
	return d, onlinestats.KS(data1, data2)
}

//np.searchsorted
func searchsorted(array, values []float64) []int {
	var indexes []int
	for _, val := range values {
		indexes = append(indexes, location(array, val))
	}
	return indexes
}

func location(array []float64, key float64) int {
	i := 0
	size := len(array)
	for {
		mid := (i + size) / 2
		if i == size {
			break
		}
		if array[mid] < key {
			i = mid + 1
		} else {
			size = mid
		}
	}
	return i
}

/*
// adfuller
func ADFuller(x []float64, maxlog int) []float64 {
	trenddict := make(map[int]string)
	regression := "nc"
	nobs := len(x)
	xdiff := arraydiff(x)
	xdall = lagmat(xdiff)
	return xdiff
}

func arraydiff(x []float64) []float64 {
	var rst []float64
	l := len(x) -1
	for i:= 0; i< l; i++ {
		rst = append(rst, x[i+1] - x[i])
	}
	return rst
}
*/
