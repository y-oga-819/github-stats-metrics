package utils

import (
	"math"
	"sort"
	"time"
)

// StatisticsCalculator は統計計算を行うユーティリティ
type StatisticsCalculator struct{}

// NewStatisticsCalculator は新しい統計計算器を作成
func NewStatisticsCalculator() *StatisticsCalculator {
	return &StatisticsCalculator{}
}

// FloatStatistics は浮動小数点数の統計情報
type FloatStatistics struct {
	Count      int     `json:"count"`
	Sum        float64 `json:"sum"`
	Mean       float64 `json:"mean"`
	Median     float64 `json:"median"`
	Mode       float64 `json:"mode"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Range      float64 `json:"range"`
	Variance   float64 `json:"variance"`
	StdDev     float64 `json:"stdDev"`
	Skewness   float64 `json:"skewness"`
	Kurtosis   float64 `json:"kurtosis"`
	Percentiles Percentiles `json:"percentiles"`
}

// IntStatistics は整数の統計情報
type IntStatistics struct {
	Count      int     `json:"count"`
	Sum        int     `json:"sum"`
	Mean       float64 `json:"mean"`
	Median     float64 `json:"median"`
	Mode       int     `json:"mode"`
	Min        int     `json:"min"`
	Max        int     `json:"max"`
	Range      int     `json:"range"`
	Variance   float64 `json:"variance"`
	StdDev     float64 `json:"stdDev"`
	Percentiles Percentiles `json:"percentiles"`
}

// DurationStatistics は時間の統計情報
type DurationStatistics struct {
	Count      int           `json:"count"`
	Sum        time.Duration `json:"sum"`
	Mean       time.Duration `json:"mean"`
	Median     time.Duration `json:"median"`
	Min        time.Duration `json:"min"`
	Max        time.Duration `json:"max"`
	Range      time.Duration `json:"range"`
	Variance   time.Duration `json:"variance"`
	StdDev     time.Duration `json:"stdDev"`
	Percentiles DurationPercentiles `json:"percentiles"`
}

// Percentiles はパーセンタイル情報
type Percentiles struct {
	P10  float64 `json:"p10"`
	P25  float64 `json:"p25"`
	P50  float64 `json:"p50"`  // Median
	P75  float64 `json:"p75"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
}

// DurationPercentiles は時間のパーセンタイル情報
type DurationPercentiles struct {
	P10  time.Duration `json:"p10"`
	P25  time.Duration `json:"p25"`
	P50  time.Duration `json:"p50"`  // Median
	P75  time.Duration `json:"p75"`
	P90  time.Duration `json:"p90"`
	P95  time.Duration `json:"p95"`
	P99  time.Duration `json:"p99"`
}

// CalculateFloatStatistics は浮動小数点数の包括統計を計算
func (calc *StatisticsCalculator) CalculateFloatStatistics(values []float64) FloatStatistics {
	if len(values) == 0 {
		return FloatStatistics{}
	}

	// ソート済みのコピーを作成
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	stats := FloatStatistics{
		Count: len(values),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	stats.Range = stats.Max - stats.Min
	stats.Sum = calc.sumFloat64(values)
	stats.Mean = stats.Sum / float64(stats.Count)
	stats.Median = calc.medianFloat64(sorted)
	stats.Mode = calc.modeFloat64(values)
	stats.Variance = calc.varianceFloat64(values, stats.Mean)
	stats.StdDev = math.Sqrt(stats.Variance)
	stats.Skewness = calc.skewnessFloat64(values, stats.Mean, stats.StdDev)
	stats.Kurtosis = calc.kurtosisFloat64(values, stats.Mean, stats.StdDev)
	stats.Percentiles = calc.percentilesFloat64(sorted)

	return stats
}

// CalculateIntStatistics は整数の包括統計を計算
func (calc *StatisticsCalculator) CalculateIntStatistics(values []int) IntStatistics {
	if len(values) == 0 {
		return IntStatistics{}
	}

	// ソート済みのコピーを作成
	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)

	stats := IntStatistics{
		Count: len(values),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	stats.Range = stats.Max - stats.Min
	stats.Sum = calc.sumInt(values)
	stats.Mean = float64(stats.Sum) / float64(stats.Count)
	stats.Median = calc.medianInt(sorted)
	stats.Mode = calc.modeInt(values)
	
	// 分散・標準偏差計算のため浮動小数点に変換
	floatValues := make([]float64, len(values))
	for i, v := range values {
		floatValues[i] = float64(v)
	}
	stats.Variance = calc.varianceFloat64(floatValues, stats.Mean)
	stats.StdDev = math.Sqrt(stats.Variance)
	stats.Percentiles = calc.percentilesInt(sorted)

	return stats
}

// CalculateDurationStatistics は時間の包括統計を計算
func (calc *StatisticsCalculator) CalculateDurationStatistics(values []time.Duration) DurationStatistics {
	if len(values) == 0 {
		return DurationStatistics{}
	}

	// ソート済みのコピーを作成
	sorted := make([]time.Duration, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	stats := DurationStatistics{
		Count: len(values),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	stats.Range = stats.Max - stats.Min
	stats.Sum = calc.sumDuration(values)
	stats.Mean = stats.Sum / time.Duration(stats.Count)
	stats.Median = calc.medianDuration(sorted)
	
	// 分散・標準偏差は秒単位で計算
	meanSeconds := stats.Mean.Seconds()
	variance := 0.0
	for _, v := range values {
		diff := v.Seconds() - meanSeconds
		variance += diff * diff
	}
	variance /= float64(len(values))
	
	stats.Variance = time.Duration(variance * float64(time.Second))
	stats.StdDev = time.Duration(math.Sqrt(variance) * float64(time.Second))
	stats.Percentiles = calc.percentilesDuration(sorted)

	return stats
}

// ヘルパーメソッド群

func (calc *StatisticsCalculator) sumFloat64(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

func (calc *StatisticsCalculator) sumInt(values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

func (calc *StatisticsCalculator) sumDuration(values []time.Duration) time.Duration {
	sum := time.Duration(0)
	for _, v := range values {
		sum += v
	}
	return sum
}

func (calc *StatisticsCalculator) medianFloat64(sortedValues []float64) float64 {
	n := len(sortedValues)
	if n%2 == 0 {
		return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	}
	return sortedValues[n/2]
}

func (calc *StatisticsCalculator) medianInt(sortedValues []int) float64 {
	n := len(sortedValues)
	if n%2 == 0 {
		return float64(sortedValues[n/2-1]+sortedValues[n/2]) / 2
	}
	return float64(sortedValues[n/2])
}

func (calc *StatisticsCalculator) medianDuration(sortedValues []time.Duration) time.Duration {
	n := len(sortedValues)
	if n%2 == 0 {
		return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	}
	return sortedValues[n/2]
}

func (calc *StatisticsCalculator) modeFloat64(values []float64) float64 {
	frequency := make(map[float64]int)
	for _, v := range values {
		frequency[v]++
	}
	
	maxCount := 0
	mode := 0.0
	for value, count := range frequency {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}
	return mode
}

func (calc *StatisticsCalculator) modeInt(values []int) int {
	frequency := make(map[int]int)
	for _, v := range values {
		frequency[v]++
	}
	
	maxCount := 0
	mode := 0
	for value, count := range frequency {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}
	return mode
}

func (calc *StatisticsCalculator) varianceFloat64(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	return variance / float64(len(values))
}

func (calc *StatisticsCalculator) skewnessFloat64(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 3 {
		return 0.0
	}
	
	skewness := 0.0
	n := float64(len(values))
	for _, v := range values {
		standardized := (v - mean) / stdDev
		skewness += math.Pow(standardized, 3)
	}
	
	return skewness / n
}

func (calc *StatisticsCalculator) kurtosisFloat64(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 4 {
		return 0.0
	}
	
	kurtosis := 0.0
	n := float64(len(values))
	for _, v := range values {
		standardized := (v - mean) / stdDev
		kurtosis += math.Pow(standardized, 4)
	}
	
	return (kurtosis / n) - 3.0 // 正規分布の尖度3を基準とした超過尖度
}

func (calc *StatisticsCalculator) percentilesFloat64(sortedValues []float64) Percentiles {
	return Percentiles{
		P10: calc.percentileFloat64(sortedValues, 0.10),
		P25: calc.percentileFloat64(sortedValues, 0.25),
		P50: calc.percentileFloat64(sortedValues, 0.50),
		P75: calc.percentileFloat64(sortedValues, 0.75),
		P90: calc.percentileFloat64(sortedValues, 0.90),
		P95: calc.percentileFloat64(sortedValues, 0.95),
		P99: calc.percentileFloat64(sortedValues, 0.99),
	}
}

func (calc *StatisticsCalculator) percentilesInt(sortedValues []int) Percentiles {
	return Percentiles{
		P10: calc.percentileInt(sortedValues, 0.10),
		P25: calc.percentileInt(sortedValues, 0.25),
		P50: calc.percentileInt(sortedValues, 0.50),
		P75: calc.percentileInt(sortedValues, 0.75),
		P90: calc.percentileInt(sortedValues, 0.90),
		P95: calc.percentileInt(sortedValues, 0.95),
		P99: calc.percentileInt(sortedValues, 0.99),
	}
}

func (calc *StatisticsCalculator) percentilesDuration(sortedValues []time.Duration) DurationPercentiles {
	return DurationPercentiles{
		P10: calc.percentileDuration(sortedValues, 0.10),
		P25: calc.percentileDuration(sortedValues, 0.25),
		P50: calc.percentileDuration(sortedValues, 0.50),
		P75: calc.percentileDuration(sortedValues, 0.75),
		P90: calc.percentileDuration(sortedValues, 0.90),
		P95: calc.percentileDuration(sortedValues, 0.95),
		P99: calc.percentileDuration(sortedValues, 0.99),
	}
}

func (calc *StatisticsCalculator) percentileFloat64(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0.0
	}
	if len(sortedValues) == 1 {
		return sortedValues[0]
	}
	
	index := p * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sortedValues[lower]
	}
	
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func (calc *StatisticsCalculator) percentileInt(sortedValues []int, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0.0
	}
	if len(sortedValues) == 1 {
		return float64(sortedValues[0])
	}
	
	index := p * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return float64(sortedValues[lower])
	}
	
	weight := index - float64(lower)
	return float64(sortedValues[lower])*(1-weight) + float64(sortedValues[upper])*weight
}

func (calc *StatisticsCalculator) percentileDuration(sortedValues []time.Duration, p float64) time.Duration {
	if len(sortedValues) == 0 {
		return 0
	}
	if len(sortedValues) == 1 {
		return sortedValues[0]
	}
	
	index := p * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sortedValues[lower]
	}
	
	weight := index - float64(lower)
	lowerVal := float64(sortedValues[lower])
	upperVal := float64(sortedValues[upper])
	
	return time.Duration(lowerVal*(1-weight) + upperVal*weight)
}

// OutlierDetection は外れ値検出の結果
type OutlierDetection struct {
	Method       string    `json:"method"`
	LowerBound   float64   `json:"lowerBound"`
	UpperBound   float64   `json:"upperBound"`
	Outliers     []float64 `json:"outliers"`
	OutlierCount int       `json:"outlierCount"`
}

// DetectOutliersIQR はIQR法による外れ値検出
func (calc *StatisticsCalculator) DetectOutliersIQR(values []float64) OutlierDetection {
	if len(values) < 4 {
		return OutlierDetection{Method: "IQR", Outliers: []float64{}}
	}
	
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	q1 := calc.percentileFloat64(sorted, 0.25)
	q3 := calc.percentileFloat64(sorted, 0.75)
	iqr := q3 - q1
	
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr
	
	var outliers []float64
	for _, v := range values {
		if v < lowerBound || v > upperBound {
			outliers = append(outliers, v)
		}
	}
	
	return OutlierDetection{
		Method:       "IQR",
		LowerBound:   lowerBound,
		UpperBound:   upperBound,
		Outliers:     outliers,
		OutlierCount: len(outliers),
	}
}

// DetectOutliersZScore はZ-Score法による外れ値検出
func (calc *StatisticsCalculator) DetectOutliersZScore(values []float64, threshold float64) OutlierDetection {
	if len(values) < 2 {
		return OutlierDetection{Method: "Z-Score", Outliers: []float64{}}
	}
	
	stats := calc.CalculateFloatStatistics(values)
	
	var outliers []float64
	for _, v := range values {
		zScore := math.Abs(v-stats.Mean) / stats.StdDev
		if zScore > threshold {
			outliers = append(outliers, v)
		}
	}
	
	lowerBound := stats.Mean - threshold*stats.StdDev
	upperBound := stats.Mean + threshold*stats.StdDev
	
	return OutlierDetection{
		Method:       "Z-Score",
		LowerBound:   lowerBound,
		UpperBound:   upperBound,
		Outliers:     outliers,
		OutlierCount: len(outliers),
	}
}

// TrendAnalysis はトレンド分析の結果
type TrendAnalysis struct {
	Slope            float64 `json:"slope"`
	Intercept        float64 `json:"intercept"`
	CorrelationCoeff float64 `json:"correlationCoeff"`
	Trend            string  `json:"trend"` // "increasing", "decreasing", "stable"
	Confidence       float64 `json:"confidence"`
}

// AnalyzeTrend は時系列データのトレンド分析
func (calc *StatisticsCalculator) AnalyzeTrend(values []float64) TrendAnalysis {
	n := len(values)
	if n < 2 {
		return TrendAnalysis{Trend: "insufficient_data"}
	}
	
	// X値を生成（インデックス）
	xValues := make([]float64, n)
	for i := range xValues {
		xValues[i] = float64(i)
	}
	
	// 線形回帰
	xMean := calc.sumFloat64(xValues) / float64(n)
	yMean := calc.sumFloat64(values) / float64(n)
	
	numerator := 0.0
	denominator := 0.0
	for i := 0; i < n; i++ {
		numerator += (xValues[i] - xMean) * (values[i] - yMean)
		denominator += (xValues[i] - xMean) * (xValues[i] - xMean)
	}
	
	slope := numerator / denominator
	intercept := yMean - slope*xMean
	
	// 相関係数
	correlation := calc.correlationCoefficient(xValues, values)
	
	// トレンド判定
	trend := "stable"
	confidence := math.Abs(correlation)
	
	if math.Abs(slope) > 0.1 && confidence > 0.5 {
		if slope > 0 {
			trend = "increasing"
		} else {
			trend = "decreasing"
		}
	}
	
	return TrendAnalysis{
		Slope:            slope,
		Intercept:        intercept,
		CorrelationCoeff: correlation,
		Trend:            trend,
		Confidence:       confidence,
	}
}

// correlationCoefficient は相関係数を計算
func (calc *StatisticsCalculator) correlationCoefficient(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}
	
	n := float64(len(x))
	xMean := calc.sumFloat64(x) / n
	yMean := calc.sumFloat64(y) / n
	
	numerator := 0.0
	xDenominator := 0.0
	yDenominator := 0.0
	
	for i := 0; i < len(x); i++ {
		xDiff := x[i] - xMean
		yDiff := y[i] - yMean
		
		numerator += xDiff * yDiff
		xDenominator += xDiff * xDiff
		yDenominator += yDiff * yDiff
	}
	
	denominator := math.Sqrt(xDenominator * yDenominator)
	if denominator == 0 {
		return 0.0
	}
	
	return numerator / denominator
}