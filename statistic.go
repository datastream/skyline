package skyline

import (
	"github.com/VividCortex/ewma"
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
func Std(series []float64) float64 {
	mean := Mean(series)
	sum := 0.0
	for _, val := range series {
		sum += math.Pow(val-mean, 2)
	}
	return math.Sqrt(sum)
}

// least squares linear regression
func linearRegressionLSE(timeseries []TimePoint) (float64, float64) {
	q := len(timeseries)
	if q == 0 {
		return 0, 0
	}
	p := float64(q)
	sum_x, sum_y, sum_xx, sum_xy := 0.0, 0.0, 0.0, 0.0
	for _, p := range timeseries {
		sum_x += float64(p.timestamp)
		sum_y += p.value
		sum_xx += float64(p.timestamp * p.timestamp)
		sum_xy += float64(p.timestamp) * p.value
	}
	m := (p*sum_xy - sum_x*sum_y) / (p*sum_xx - sum_x*sum_x)
	c := (sum_y / p) - (m * sum_x / p)
	return m, c
}

// Exponentially-weighted moving std
func ewmstd(series []float64, com float64) float64 {
	m2nd := ewma.NewMovingAverage(com)
	m1st := ewma.NewMovingAverage(com)
	for _, val := range series {
		m2nd.Add(val * val)
	}
	for _, val := range series {
		m1st.Add(val)
	}
	result := m2nd.Value() - math.Pow(m1st.Value(), 2)
	result *= (1.0 + 2.0*com) / (2.0 * com)
	return math.Sqrt(result)
}

// numpy.histogram
func histogram(series []float64, bins int) ([]float64, []float64) {
	var bin_edges []float64
	var hist []float64
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
	bl := len(bin_edges)
	hist = make([]float64, bl-1)
	for i := 0; i < bl-1; i++ {
		for _, val := range series {
			if val >= bin_edges[i] && val < bin_edges[i+1] {
				hist[i] += 1
			}
			if (i+1) == bl && val == bin_edges[i+1] {
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
