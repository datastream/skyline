package skyline

// double stdtri ( int k, double p );
import "C"

import (
	"math"
	"sort"
)

// series.mean
func Mean(series []float64) float64 {
	if len(series) == 0 {
		return 0
	}
	sum := 0.0
	for _, val := range series {
		sum += val
	}
	return sum / float64(len(series))
}

// series.median
func Median(series []float64) float64 {
	var median float64
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

// series.std
// corrected sample standard deviation
// http://en.wikipedia.org/wiki/Standard_deviation#Estimation
func Std(series []float64) float64 {
	mean := Mean(series)
	sum := 0.0
	for _, val := range series {
		sum += math.Pow(val-mean, 2)
	}
	return math.Sqrt(sum / float64(len(series)-1))
}

// least squares linear regression
func LinearRegressionLSE(timeseries []TimePoint) (float64, float64) {
	q := len(timeseries)
	if q == 0 {
		return 0, 0
	}
	p := float64(q)
	sum_x, sum_y, sum_xx, sum_xy := 0.0, 0.0, 0.0, 0.0
	for _, p := range timeseries {
		sum_x += float64(p.Timestamp)
		sum_y += p.Value
		sum_xx += float64(p.Timestamp * p.Timestamp)
		sum_xy += float64(p.Timestamp) * p.Value
	}
	m := (p*sum_xy - sum_x*sum_y) / (p*sum_xx - sum_x*sum_x)
	c := (sum_y - m*sum_x) / p
	return m, c
}

func Ewma(series []float64, com float64) []float64 {
	return ewma(series, com, true)
}

// ewma
func ewma(series []float64, com float64, adjust bool) []float64 {
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
	if adjust {
		ret[0] = series[0] / (1 + com)
	} else {
		ret[0] = series[0]
	}
	for i := 1; i < N; i++ {
		cur = series[i]
		prev = ret[i-1]
		if !math.IsNaN(cur) {
			if !math.IsNaN(cur) {
				ret[i] = (com*prev + cur) / (1 + com)
			} else {
				ret[i] = cur / (1 + com)
			}
		} else {
			ret[i] = prev
		}
	}
	if adjust {
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
	}
	return ret
}

// Exponentially-weighted moving std
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

// numpy.histogram
func Histogram(series []float64, bins int) ([]int, []float64) {
	var bin_edges []float64
	var hist []int
	l := len(series)
	if l == 0 {
		return hist, bin_edges
	}
	sort.Float64s(series)
	w := (series[l-1] - series[0]) / float64(bins)
	for i := 0; i < bins; i++ {
		bin_edges = append(bin_edges, w*float64(i)+series[0])
		if bin_edges[len(bin_edges)-1] >= series[l-1] {
			break
		}
	}
	bin_edges = append(bin_edges, w*float64(bins)+series[0])
	bl := len(bin_edges)
	hist = make([]int, bl-1)
	for i := 0; i < bl-1; i++ {
		for _, val := range series {
			if val >= bin_edges[i] && val < bin_edges[i+1] {
				hist[i] += 1
				continue
			}
			if i == (bl-2) && val >= bin_edges[i] && val <= bin_edges[i+1] {
				hist[i] += 1
			}
		}
	}
	return hist, bin_edges
}

//scipy.ks_2samp

func KS2Samp(data1, data2 []float64) (float64, float64) {
	sort.Float64s(data1)
	sort.Float64s(data2)
	n1 := len(data1)
	n2 := len(data2)
	var data_all []float64
	data_all = append(data_all, data1...)
	data_all = append(data_all, data2...)
	index1 := searchsorted(data1, data_all)
	index2 := searchsorted(data2, data_all)
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
	en := math.Sqrt(float64(n1*n2) / float64(n1+n2))
	prob := kolmogorov((en + 0.12 + 0.11/en) * d)
	return d, prob
}

// Kolmogorov's limiting distribution of two-sided test, returns
// probability that sqrt(n) * max deviation > y,
// or that max deviation > y/sqrt(n).
// The approximation is useful for the tail of the distribution
// when n is large.
// scipy/special/cephes/kolmogorov.c
func kolmogorov(y float64) float64 {
	if y < 1.1e-16 {
		return 1.0
	}
	x := -2.0 * y * y
	sign := 1.0
	p := 0.0
	r := 1.0
	var t float64
	for {
		t = math.Exp(x * r * r)
		p += sign * t
		if t == 0.0 {
			break
		}
		r += 1.0
		sign = -sign
		if (t / p) <= 1.1e-16 {
			break
		}
	}
	return (p + p)
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

// stdtri, t.isf(q, df) = -stdtri(df, q)
// http://www.netlib.org/cephes/{cmath.tgz,eval.tgz, cprob.tgz}
func StudentT_ISF_For(q float64, df int) float64 {
	return -float64(C.stdtri(C.int(df), C.double(q)))
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
