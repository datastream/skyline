package skyline

import (
	"github.com/GaryBoone/GoStats/stats"
	"github.com/VividCortex/ewma"
	"math"
	"sort"
	"time"
)

// This is no man's land. Do anything you want in here,
// as long as you return a boolean that determines whether the input
// timeseries is anomalous or not.

// To add an algorithm, define it here, and add its name to settings.ALGORITHMS

const (
	FULL_DURATION = 1
)

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
		sum += math.Pow(val - mean, 2)
	}
	return math.Sqrt(sum)
}

// ewmstd
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

// This is a utility function used to calculate the average of the last three
// datapoints in the series as a measure, instead of just the last datapoint.
// It reduces noise, but it also reduces sensitivity and increases the delay
// to detection.
func TailAvg(series []float64) float64 {
	l := len(series)
	if l == 0 {
		return 0
	}
	if l < 3 {
		return series[l-1]
	}
	return (series[l-1] + series[l-2] + series[l-3]) / 3
}

// A timeseries is anomalous if the deviation of its latest datapoint with
// respect to the median is X times larger than the median of deviations.
func MedianAbsoluteDeviation(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	sort.Float64s(series)
	median := Median(series)
	var demedianed []float64
	for _, val := range series {
		demedianed = append(demedianed, math.Abs(val - median))
	}
	sort.Float64s(demedianed)
	median_deviation := Median(demedianed)
	if median_deviation == 0 {
		return false
	}
	test_statistic := demedianed[len(demedianed)-1] / median_deviation
	if test_statistic > 6 {
		return true
	}
	return false
}

// A timeseries is anomalous if the Z score is greater than the Grubb's score.
func Grubbs(timeseries []TimePoint) bool {
	/*
	series := ValueArray(timeseries)
	stdDev := Std(series)
	mean := Mean(series)
	tail_average := TailAvg(series)
	z_score := (tail_average - mean) / stdDev
	len_series := len(series)

	t := stat.NextStudentsT(float64(len_series - 2))
	 */
	// scipy.stats.t.isf(len_series -2, 1 - 0.05, len_series -2)
	// len_series - 2 is studentsT's arg
	// t.isf(a,b) == t.ppf(1-a,b)
	// ppf:  Percent point function (inverse of cdf), Quantile function
	// need StudentsT_InvCDF_For, TINV
	// stat.StudentsT_InvCDF_For(len_series-2, 1-0.05/(2*len_series), 1)
	/*
	threshold := 0.0
	threshold_squared := threshold * threshold
	grubbs_score := (float64(len_series - 1) / math.Sqrt(float64(len_series))) * math.Sqrt(threshold_squared/(float64(len_series-2)+threshold_squared))
	return z_score > grubbs_score
	 */
	return true
}

// Calcuate the simple average over one hour, FULL_DURATION seconds ago.
// A timeseries is anomalous if the average of the last three datapoints
// are outside of three standard deviations of this value.
func FirstHourAverage(timeseries []TimePoint) bool {
	var series []float64
	last_hour_threshold := time.Now().Unix() - (FULL_DURATION - 3600)
	for _, val := range timeseries {
		if val.timestamp < last_hour_threshold {
			series = append(series, val.value)
		}
	}
	mean := Mean(series)
	stdDev := Std(series)
	t := TailAvg(series)
	return math.Abs(t-mean) > 3*stdDev
}

// A timeseries is anomalous if the absolute value of the average of the latest
// three datapoint minus the moving average is greater than one standard
// deviation of the average. This does not exponentially weight the MA and so
// is better for detecting anomalies with respect to the entire series.
func SimpleStddevFromMovingAverage(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	mean := Mean(series)
	stdDev := Std(series)
	t := TailAvg(series)
	return math.Abs(t-mean) > 3*stdDev
}

// A timeseries is anomalous if the absolute value of the average of the latest
// three datapoint minus the moving average is greater than one standard
// deviation of the moving average. This is better for finding anomalies with
// respect to the short term trends.
func StddevFromMovingAverage(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	e := ewma.NewMovingAverage(50)
	for _, val := range series {
		e.Add(val)
	}
	expAverage := e.Value()
	stdDev := ewmstd(series, 50)
	return math.Abs(series[len(series)-1]-expAverage) > (3 * stdDev)
}

// A timeseries is anomalous if the value of the next datapoint in the
// series is farther than a standard deviation out in cumulative terms
// after subtracting the mean from each data point.
func MeanSubtractionCumulation(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	mean := Mean(series)
	for i, val := range series {
		series[i] = val - mean
	}
	stdDev := Std(series)
	/*
	e := ewma.NewMovingAverage(15)
	for _, val := range series {
		e.Add(val)
	}
	expAverage := e.Value()
	*/
	return math.Abs(series[len(series)-1]) > 3*stdDev
}

// A timeseries is anomalous if the average of the last three datapoints
// on a projected least squares model is greater than three sigma.
func LeastSquares(timeseries []TimePoint) bool {
	var r stats.Regression
	for _, val := range timeseries {
		r.Update(float64(val.timestamp), val.value)
	}
	m := r.Slope()
	c := r.Intercept()
	var errs []float64
	for _, val := range timeseries {
		projected := m * float64(val.timestamp) + c
		errs = append(errs, val.value - projected)
	}
	l := len(errs)
	if l < 3 {
		return false
	}
	std_dev := Std(errs)
	t := (errs[l-1] + errs[l-2] + errs[l-3]) / 3
	return math.Abs(t) > std_dev*3 && math.Trunc(std_dev) != 0 && math.Trunc(t) != 0
}

// A timeseries is anomalous if the average of the last three datapoints falls
// into a histogram bin with less than 20 other datapoints (you'll need to tweak
// that number depending on your data)
// Returns: the size of the bin which contains the tail_avg. Smaller bin size
// means more anomalous.
func HistogramBins(timeseries []TimePoint) {
	//series := ValueArray(timeseries)
	//t := TailAvg(series)
	/*
	   series = scipy.array([x[1] for x in timeseries])
	   t = tail_avg(timeseries)
	   h = np.histogram(series, bins=15)
	   bins = h[1]
	   for index, bin_size in enumerate(h[0]):
	       if bin_size <= 20:
	           if index == 0:
	               if t <= bins[0]:
	                   return True
	           elif t >= bins[index] and t < bins[index + 1]:
	                   return True

	   return False
	*/
}

// A timeseries is anomalous if 2 sample Kolmogorov-Smirnov test indicates
// that data distribution for last 10 minutes is different from last hour.
// It produces false positives on non-stationary series so Augmented
// Dickey-Fuller test applied to check for stationarity.
func KsTest(timeseries []TimePoint) bool {
	current := time.Now().Unix()
	hour_ago := current - 3600
	ten_minutes_ago := current - 600
	var reference []float64
	var probe []float64
	for _, val := range timeseries {
		if val.timestamp >= hour_ago && val.timestamp < ten_minutes_ago {
			reference = append(reference, val.value)
		}
		if val.timestamp >= ten_minutes_ago {
			probe = append(probe, val.value)
		}
	}
	if len(reference) < 20 || len(probe) < 20 {
		return false
	}
	/*
	ks_d,ks_p_value := scipy.stats.ks_2samp(reference, probe)
	if ks_p_value < 0.05 && ks_d > 0.5 {
		adf := sm.tsa.stattools.adfuller(reference, 10)
		if adf[1] < 0.05 {
			return true
		}
	}
	 */
	return false
}

// Filter timeseries and run selected algorithm.
func RunSelectedAlgorithm(f func([]TimePoint) float64, timeseries []TimePoint) {
	/*
	 ensemble := f(timeseries)
	threshold := len(ensemble) - CONSENSUS
	if ensemble <= threshold {
		return true, ensemble, TailAvg(series)
	}
	return true, ensemble, timeseries[len(timeseries)-1][1]
	 */
}
