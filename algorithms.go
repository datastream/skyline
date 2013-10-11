package skyline

import (
	"math"
	"time"
)

// This is no man's land. Do anything you want in here,
// as long as you return a boolean that determines whether the input
// timeseries is anomalous or not.

// To add an algorithm, define it here, and add its name to settings.ALGORITHMS

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
	median := Median(series)
	var demedianed []float64
	for _, val := range series {
		demedianed = append(demedianed, math.Abs(val-median))
	}
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
	series := ValueArray(timeseries)
	stdDev := Std(series)
	mean := Mean(series)
	tail_average := TailAvg(series)
	z_score := (tail_average - mean) / stdDev
	len_series := len(series)
	// scipy.stats.t.isf(.05 / (2 * len_series) , len_series - 2)
	threshold := StudentT_ISF_For(0.05/float64(2*len_series), len_series-2)
	threshold_squared := threshold * threshold
	grubbs_score := (float64(len_series-1) / math.Sqrt(float64(len_series))) * math.Sqrt(threshold_squared/(float64(len_series-2)+threshold_squared))
	return z_score > grubbs_score
}

// Calcuate the simple average over one hour, FULL_DURATION seconds ago.
// A timeseries is anomalous if the average of the last three datapoints
// are outside of three standard deviations of this value.
func FirstHourAverage(timeseries []TimePoint, Full_duration int64) bool {
	var series []float64
	last_hour_threshold := time.Now().Unix() - (full_duration - 3600)
	for _, val := range timeseries {
		if val.Timestamp < last_hour_threshold {
			series = append(series, val.Value)
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
	expAverage := Ewma(series, 50)
	stdDev := EwmStd(series, 50)
	return math.Abs(series[len(series)-1]-expAverage[len(expAverage)-1]) > (3 * stdDev[len(stdDev)-1])
}

// A timeseries is anomalous if the value of the next datapoint in the
// series is farther than a standard deviation out in cumulative terms
// after subtracting the mean from each data point.
func MeanSubtractionCumulation(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	mean := Mean(series[:len(series)-1])
	for i, val := range series {
		series[i] = val - mean
	}
	stdDev := Std(series[:len(series)-1])
	// expAverage = pandas.stats.moments.ewma(series, com=15)
	return math.Abs(series[len(series)-1]) > 3*stdDev
}

// A timeseries is anomalous if the average of the last three datapoints
// on a projected least squares model is greater than three sigma.
func LeastSquares(timeseries []TimePoint) bool {
	m, c := LinearRegressionLSE(timeseries)
	var errs []float64
	for _, val := range timeseries {
		projected := m*float64(val.Timestamp) + c
		errs = append(errs, val.Value-projected)
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
func HistogramBins(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	t := TailAvg(series)
	hist, bins := Histogram(series, 15)
	for i, v := range hist {
		if v <= 20 {
			if i == 0 {
				if t <= bins[0] {
					return true
				}
			} else if t > bins[i] && t < bins[i+1] {
				return true
			}
		}
	}
	return false
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
		if val.Timestamp >= hour_ago && val.Timestamp < ten_minutes_ago {
			reference = append(reference, val.Value)
		}
		if val.Timestamp >= ten_minutes_ago {
			probe = append(probe, val.Value)
		}
	}
	if len(reference) < 20 || len(probe) < 20 {
		return false
	}
	ks_d, ks_p_value := KS2Samp(reference, probe)
	if ks_p_value < 0.05 && ks_d > 0.5 {
		/*
			adf := ADFuller(reference, 10)
			if adf[1] < 0.05 {
				return true
			}
		*/
	}
	return false
}

// Filter timeseries and run selected algorithm.
/*
func RunSelectedAlgorithm(f func([]TimePoint) float64, timeseries []TimePoint) {
	 ensemble := f(timeseries)
	 threshold := len(ensemble) - CONSENSUS
	var ensemble_false_count int
	for _,v := range ensemble {
		if !v {
			ensemble_false_count ++
		}
	}
	if ensemble_false_count <= threshold {
		return true, ensemble, timeseries[:len(timeseries)-1].Value
	 }
	return false, ensemble, timeseries[:len(timeseries)-1].Value
}
*/
