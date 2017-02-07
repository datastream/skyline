package skyline

import (
	"github.com/gonum/stat"
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

// KolmogorovSmirnov performs the two-sample Kolmogorov–Smirnov test. The null
// hypothesis is that the two datasets are coming from the same continuous
// distribution. The α parameter specifies the significance level. If the test
// rejects the null hypothesis, the function returns true; otherwise, false is
// returned. The second and third outputs of the function are the p-value and
// Kolmogorov–Smirnov statistic of the test, respectively.
//
// https://en.wikipedia.org/wiki/Kolmogorov%E2%80%93Smirnov_test
func KolmogorovSmirnov(data1, data2 []float64, α float64) (bool, float64, float64) {
	const (
		terms = 101
	)

	statistic := stat.KolmogorovSmirnov(data1, nil, data2, nil)

	// M. Stephens. Use of the Kolmogorov–Smirnov, Cramer-Von Mises and Related
	// Statistics Without Extensive Tables. Journal of the Royal Statistical
	// Society. Series B (Methodological), vol. 32, no. 1 (1970), pp. 115–122.
	//
	// http://www.jstor.org/stable/2984408
	n1, n2 := len(data1), len(data2)
	γ := math.Sqrt(float64(n1*n2) / float64(n1+n2))
	λ := (γ + 0.12 + 0.11/γ) * statistic

	// Kolmogorov distribution
	//
	// https://en.wikipedia.org/wiki/Kolmogorov%E2%80%93Smirnov_test#Kolmogorov_distribution
	pvalue, sign, k := 0.0, 1.0, 1.0
	for i := 0; i < terms; i++ {
		pvalue += sign * math.Exp(-2*λ*λ*k*k)
		sign, k = -sign, k+1
	}
	pvalue *= 2
	if pvalue < 0 {
		pvalue = 0
	} else if pvalue > 1 {
		pvalue = 1
	}

	return α >= pvalue, pvalue, statistic
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
