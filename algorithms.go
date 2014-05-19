package skyline

import (
	"code.google.com/p/probab/dst"
	"math"
	"time"
)

// This is no man's land. Do anything you want in here,
// as long as you return a boolean that determines whether the input
// timeseries is anomalous or not.

// To add an algorithm, define it here, and add its name to settings.ALGORITHMS

// TailAvg is a utility function used to calculate the average of the last three
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

// MedianAbsoluteDeviation function
// A timeseries is anomalous if the deviation of its latest datapoint with
// respect to the median is X times larger than the median of deviations.
func MedianAbsoluteDeviation(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	median := Median(series)
	var demedianed []float64
	for _, val := range series {
		demedianed = append(demedianed, math.Abs(val-median))
	}
	medianDeviation := Median(demedianed)
	if medianDeviation == 0 {
		return false
	}
	testStatistic := demedianed[len(demedianed)-1] / medianDeviation
	if testStatistic > 6 {
		return true
	}
	return false
}

// Grubbs score
// A timeseries is anomalous if the Z score is greater than the Grubb's score.
func Grubbs(timeseries []TimePoint) bool {
	series := ValueArray(timeseries)
	stdDev := Std(series)
	mean := Mean(series)
	tailAverage := TailAvg(series)
	// http://en.wikipedia.org/wiki/Grubbs'_test_for_outliers
	// G = (Y - Mean(Y)) / stdDev(Y)
	zScore := (tailAverage - mean) / stdDev
	lenSeries := len(series)
	// scipy.stats.t.isf(.05 / (2 * lenSeries) , lenSeries - 2)
	// when lenSeries is big, it eq stats.ZInvCDFFor(1-t)
	threshold := dst.StudentsTQtlFor(float64(lenSeries-2), 1-0.05/float64(2*lenSeries))
	thresholdSquared := threshold * threshold
	// (l-1)/l * sqr(t/(l-2+t^2))
	grubbsScore := (float64(lenSeries-1) / math.Sqrt(float64(lenSeries))) * math.Sqrt(thresholdSquared/(float64(lenSeries-2)+thresholdSquared))
	return zScore > grubbsScore
}

// FirstHourAverage function
// Calcuate the simple average over one hour, FULLDURATION seconds ago.
// A timeseries is anomalous if the average of the last three datapoints
// are outside of three standard deviations of this value.
func FirstHourAverage(timeseries []TimePoint, fullDuration int64) bool {
	var series []float64
	lastHourThreshold := time.Now().Unix() - (fullDuration - 3600)
	for _, val := range timeseries {
		if val.GetTimestamp() < lastHourThreshold {
			series = append(series, val.GetValue())
		}
	}
	mean := Mean(series)
	stdDev := Std(series)
	t := TailAvg(ValueArray(timeseries))
	return math.Abs(t-mean) > 3*stdDev
}

// SimpleStddevFromMovingAverage function
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

// StddevFromMovingAverage function
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

// MeanSubtractionCumulation function
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

// LeastSquares function
// A timeseries is anomalous if the average of the last three datapoints
// on a projected least squares model is greater than three sigma.
func LeastSquares(timeseries []TimePoint) bool {
	m, c := LinearRegressionLSE(timeseries)
	var errs []float64
	for _, val := range timeseries {
		projected := m*float64(val.GetTimestamp()) + c
		errs = append(errs, val.GetValue()-projected)
	}
	l := len(errs)
	if l < 3 {
		return false
	}
	stdDev := Std(errs)
	t := (errs[l-1] + errs[l-2] + errs[l-3]) / 3
	return math.Abs(t) > stdDev*3 && math.Trunc(stdDev) != 0 && math.Trunc(t) != 0
}

// HistogramBins function
// A timeseries is anomalous if the average of the last three datapoints falls
// into a histogram bin with less than 20 other datapoints (you'll need to tweak
// that number depending on your data)
// Returns: the size of the bin which contains the tailAvg. Smaller bin size
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

// KsTest function
// A timeseries is anomalous if 2 sample Kolmogorov-Smirnov test indicates
// that data distribution for last 10 minutes is different from last hour.
// It produces false positives on non-stationary series so Augmented
// Dickey-Fuller test applied to check for stationarity.
func KsTest(timeseries []TimePoint) bool {
	current := time.Now().Unix()
	hourAgo := current - 3600
	tenMinutesAgo := current - 600
	var reference []float64
	var probe []float64
	for _, val := range timeseries {
		if val.GetTimestamp() >= hourAgo && val.GetTimestamp() < tenMinutesAgo {
			reference = append(reference, val.GetValue())
		}
		if val.GetTimestamp() >= tenMinutesAgo {
			probe = append(probe, val.GetValue())
		}
	}
	if len(reference) < 20 || len(probe) < 20 {
		return false
	}
	ksD, ksPValue := KS2Samp(reference, probe)
	if ksPValue < 0.05 && ksD > 0.5 {
		/*
			adf := ADFuller(reference, 10)
			if adf[1] < 0.05 {
				return true
			}
		*/
	}
	return false
}

// IsAnomalouslyAnomalous function
// This method runs a meta-analysis on the metric to determine whether the
// metric has a past history of triggering. TODO: weight intervals based on datapoint
func IsAnomalouslyAnomalous(trigger_history []TimePoint, new_trigger TimePoint) (bool, []TimePoint) {
	if len(trigger_history) == 0 {
		trigger_history = append(trigger_history, new_trigger)
		return true, trigger_history
	}
	if (new_trigger.GetValue() == trigger_history[len(trigger_history)-1].GetValue()) && (new_trigger.GetTimestamp()-trigger_history[len(trigger_history)-1].GetTimestamp() <= 300) {
		return false, trigger_history
	}
	trigger_history = append(trigger_history, new_trigger)
	trigger_times := TimeArray(trigger_history)
	var intervals []float64
	for i := range trigger_times {
		if (i + 1) < len(trigger_times) {
			intervals = append(intervals, float64(trigger_times[i+1]-trigger_times[i]))
		}
	}
	mean := Mean(intervals)
	stdDev := Std(intervals)
	return math.Abs(intervals[len(intervals)-1]-mean) > 3*stdDev, trigger_history
}
