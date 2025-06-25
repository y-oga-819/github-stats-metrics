package utils

import (
	"math"
	"testing"
	"time"
)

func TestStatisticsCalculator_CalculateFloatStatistics(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name     string
		values   []float64
		expected FloatStatistics
	}{
		{
			"空の配列",
			[]float64{},
			FloatStatistics{},
		},
		{
			"単一値",
			[]float64{5.0},
			FloatStatistics{
				Count:  1,
				Sum:    5.0,
				Mean:   5.0,
				Median: 5.0,
				Mode:   5.0,
				Min:    5.0,
				Max:    5.0,
				Range:  0.0,
			},
		},
		{
			"複数値",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			FloatStatistics{
				Count:  5,
				Sum:    15.0,
				Mean:   3.0,
				Median: 3.0,
				Min:    1.0,
				Max:    5.0,
				Range:  4.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateFloatStatistics(tt.values)
			
			if result.Count != tt.expected.Count {
				t.Errorf("Count = %v, want %v", result.Count, tt.expected.Count)
			}
			
			if len(tt.values) == 0 {
				return // 空の場合は他のフィールドはチェックしない
			}
			
			if math.Abs(result.Sum-tt.expected.Sum) > 0.001 {
				t.Errorf("Sum = %v, want %v", result.Sum, tt.expected.Sum)
			}
			
			if math.Abs(result.Mean-tt.expected.Mean) > 0.001 {
				t.Errorf("Mean = %v, want %v", result.Mean, tt.expected.Mean)
			}
			
			if math.Abs(result.Median-tt.expected.Median) > 0.001 {
				t.Errorf("Median = %v, want %v", result.Median, tt.expected.Median)
			}
			
			if math.Abs(result.Min-tt.expected.Min) > 0.001 {
				t.Errorf("Min = %v, want %v", result.Min, tt.expected.Min)
			}
			
			if math.Abs(result.Max-tt.expected.Max) > 0.001 {
				t.Errorf("Max = %v, want %v", result.Max, tt.expected.Max)
			}
			
			if math.Abs(result.Range-tt.expected.Range) > 0.001 {
				t.Errorf("Range = %v, want %v", result.Range, tt.expected.Range)
			}
		})
	}
}

func TestStatisticsCalculator_CalculateIntStatistics(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	values := []int{1, 2, 3, 4, 5}
	result := calc.CalculateIntStatistics(values)
	
	if result.Count != 5 {
		t.Errorf("Count = %v, want 5", result.Count)
	}
	
	if result.Sum != 15 {
		t.Errorf("Sum = %v, want 15", result.Sum)
	}
	
	if math.Abs(result.Mean-3.0) > 0.001 {
		t.Errorf("Mean = %v, want 3.0", result.Mean)
	}
	
	if math.Abs(result.Median-3.0) > 0.001 {
		t.Errorf("Median = %v, want 3.0", result.Median)
	}
	
	if result.Min != 1 {
		t.Errorf("Min = %v, want 1", result.Min)
	}
	
	if result.Max != 5 {
		t.Errorf("Max = %v, want 5", result.Max)
	}
	
	if result.Range != 4 {
		t.Errorf("Range = %v, want 4", result.Range)
	}
}

func TestStatisticsCalculator_CalculateDurationStatistics(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	durations := []time.Duration{
		1 * time.Hour,
		2 * time.Hour,
		3 * time.Hour,
		4 * time.Hour,
		5 * time.Hour,
	}
	
	result := calc.CalculateDurationStatistics(durations)
	
	if result.Count != 5 {
		t.Errorf("Count = %v, want 5", result.Count)
	}
	
	if result.Sum != 15*time.Hour {
		t.Errorf("Sum = %v, want %v", result.Sum, 15*time.Hour)
	}
	
	if result.Mean != 3*time.Hour {
		t.Errorf("Mean = %v, want %v", result.Mean, 3*time.Hour)
	}
	
	if result.Median != 3*time.Hour {
		t.Errorf("Median = %v, want %v", result.Median, 3*time.Hour)
	}
	
	if result.Min != 1*time.Hour {
		t.Errorf("Min = %v, want %v", result.Min, 1*time.Hour)
	}
	
	if result.Max != 5*time.Hour {
		t.Errorf("Max = %v, want %v", result.Max, 5*time.Hour)
	}
	
	if result.Range != 4*time.Hour {
		t.Errorf("Range = %v, want %v", result.Range, 4*time.Hour)
	}
}

func TestStatisticsCalculator_medianFloat64(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"奇数個", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
		{"偶数個", []float64{1.0, 2.0, 3.0, 4.0}, 2.5},
		{"単一値", []float64{42.0}, 42.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.medianFloat64(tt.values)
			
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("medianFloat64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatisticsCalculator_percentileFloat64(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	
	tests := []struct {
		name       string
		percentile float64
		expected   float64
	}{
		{"10パーセンタイル", 0.10, 1.9},
		{"25パーセンタイル", 0.25, 3.25},
		{"50パーセンタイル", 0.50, 5.5},
		{"75パーセンタイル", 0.75, 7.75},
		{"90パーセンタイル", 0.90, 9.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.percentileFloat64(values, tt.percentile)
			
			if math.Abs(result-tt.expected) > 0.1 {
				t.Errorf("percentileFloat64(%v) = %v, want %v", tt.percentile, result, tt.expected)
			}
		})
	}
}

func TestStatisticsCalculator_DetectOutliersIQR(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name         string
		values       []float64
		expectedCount int
	}{
		{
			"外れ値なし",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			0,
		},
		{
			"外れ値あり",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 100.0}, // 100.0が外れ値
			1,
		},
		{
			"複数の外れ値",
			[]float64{-50.0, 1.0, 2.0, 3.0, 4.0, 5.0, 100.0}, // -50.0と100.0が外れ値
			2,
		},
		{
			"データが少ない",
			[]float64{1.0, 2.0},
			0, // 4未満のデータでは外れ値検出しない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.DetectOutliersIQR(tt.values)
			
			if result.OutlierCount != tt.expectedCount {
				t.Errorf("OutlierCount = %v, want %v", result.OutlierCount, tt.expectedCount)
			}
			
			if len(result.Outliers) != tt.expectedCount {
				t.Errorf("len(Outliers) = %v, want %v", len(result.Outliers), tt.expectedCount)
			}
			
			if result.Method != "IQR" {
				t.Errorf("Method = %v, want IQR", result.Method)
			}
		})
	}
}

func TestStatisticsCalculator_DetectOutliersZScore(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 100.0} // 100.0が明らかに外れ値
	threshold := 2.0
	
	result := calc.DetectOutliersZScore(values, threshold)
	
	if result.OutlierCount == 0 {
		t.Error("Expected to detect at least one outlier")
	}
	
	if result.Method != "Z-Score" {
		t.Errorf("Method = %v, want Z-Score", result.Method)
	}
	
	// 100.0が外れ値として検出されているかチェック
	found := false
	for _, outlier := range result.Outliers {
		if outlier == 100.0 {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected 100.0 to be detected as outlier")
	}
}

func TestStatisticsCalculator_AnalyzeTrend(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name     string
		values   []float64
		expected string
	}{
		{
			"増加トレンド",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			"increasing",
		},
		{
			"減少トレンド",
			[]float64{5.0, 4.0, 3.0, 2.0, 1.0},
			"decreasing",
		},
		{
			"安定トレンド",
			[]float64{3.0, 3.1, 2.9, 3.0, 3.1},
			"stable",
		},
		{
			"データ不足",
			[]float64{1.0},
			"insufficient_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.AnalyzeTrend(tt.values)
			
			if result.Trend != tt.expected {
				t.Errorf("Trend = %v, want %v", result.Trend, tt.expected)
			}
			
			// 相関係数の範囲チェック
			if len(tt.values) >= 2 && (result.CorrelationCoeff < -1.0 || result.CorrelationCoeff > 1.0) {
				t.Errorf("CorrelationCoeff = %v, want between -1.0 and 1.0", result.CorrelationCoeff)
			}
			
			// 信頼度の範囲チェック
			if len(tt.values) >= 2 && (result.Confidence < 0.0 || result.Confidence > 1.0) {
				t.Errorf("Confidence = %v, want between 0.0 and 1.0", result.Confidence)
			}
		})
	}
}

func TestStatisticsCalculator_correlationCoefficient(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name     string
		x        []float64
		y        []float64
		expected float64
		tolerance float64
	}{
		{
			"完全正相関",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			[]float64{2.0, 4.0, 6.0, 8.0, 10.0},
			1.0,
			0.001,
		},
		{
			"完全負相関",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			[]float64{5.0, 4.0, 3.0, 2.0, 1.0},
			-1.0,
			0.001,
		},
		{
			"無相関",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			[]float64{3.0, 3.0, 3.0, 3.0, 3.0},
			0.0,
			0.001,
		},
		{
			"長さ不一致",
			[]float64{1.0, 2.0, 3.0},
			[]float64{1.0, 2.0},
			0.0,
			0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.correlationCoefficient(tt.x, tt.y)
			
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("correlationCoefficient() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatisticsCalculator_varianceFloat64(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mean := 3.0
	expected := 2.0 // ((1-3)² + (2-3)² + (3-3)² + (4-3)² + (5-3)²) / 5 = (4+1+0+1+4)/5 = 2.0
	
	result := calc.varianceFloat64(values, mean)
	
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("varianceFloat64() = %v, want %v", result, expected)
	}
}

func TestStatisticsCalculator_skewnessFloat64(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	tests := []struct {
		name     string
		values   []float64
		mean     float64
		stdDev   float64
		expected float64 // 大まかな期待値
	}{
		{
			"対称分布",
			[]float64{1.0, 2.0, 3.0, 4.0, 5.0},
			3.0,
			math.Sqrt(2.0),
			0.0, // 対称分布なので0に近い
		},
		{
			"右に歪んだ分布",
			[]float64{1.0, 1.0, 1.0, 1.0, 10.0},
			2.8,
			3.6,
			1.0, // 正の歪度
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.skewnessFloat64(tt.values, tt.mean, tt.stdDev)
			
			// 歪度は概算値なので、符号が正しいかをチェック
			if tt.expected > 0 && result <= 0 {
				t.Errorf("Expected positive skewness, got %v", result)
			} else if tt.expected < 0 && result >= 0 {
				t.Errorf("Expected negative skewness, got %v", result)
			} else if tt.expected == 0 && math.Abs(result) > 1.0 {
				t.Errorf("Expected near-zero skewness, got %v", result)
			}
		})
	}
}

func TestStatisticsCalculator_kurtosisFloat64(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	// 正規分布に近いデータ
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mean := 3.0
	stdDev := math.Sqrt(2.0)
	
	result := calc.kurtosisFloat64(values, mean, stdDev)
	
	// 尖度は正規分布で0になるように調整済み（超過尖度）
	// 完全な正規分布ではないが、-3から3の範囲内であることをチェック
	if result < -3.0 || result > 3.0 {
		t.Errorf("Kurtosis %v is outside reasonable range [-3, 3]", result)
	}
}

func TestStatisticsCalculator_EdgeCases(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	t.Run("空の配列での外れ値検出", func(t *testing.T) {
		result := calc.DetectOutliersIQR([]float64{})
		
		if result.OutlierCount != 0 {
			t.Errorf("Expected 0 outliers for empty array, got %d", result.OutlierCount)
		}
	})
	
	t.Run("同じ値の配列での統計計算", func(t *testing.T) {
		values := []float64{5.0, 5.0, 5.0, 5.0, 5.0}
		result := calc.CalculateFloatStatistics(values)
		
		if result.Mean != 5.0 {
			t.Errorf("Mean = %v, want 5.0", result.Mean)
		}
		
		if result.Variance != 0.0 {
			t.Errorf("Variance = %v, want 0.0", result.Variance)
		}
		
		if result.StdDev != 0.0 {
			t.Errorf("StdDev = %v, want 0.0", result.StdDev)
		}
	})
	
	t.Run("単一値での相関計算", func(t *testing.T) {
		result := calc.correlationCoefficient([]float64{1.0}, []float64{2.0})
		
		if result != 0.0 {
			t.Errorf("Expected 0.0 correlation for single values, got %v", result)
		}
	})
}

func TestPercentiles_Calculation(t *testing.T) {
	calc := NewStatisticsCalculator()
	
	// 1から100までの連続した値
	values := make([]float64, 100)
	for i := 0; i < 100; i++ {
		values[i] = float64(i + 1)
	}
	
	result := calc.CalculateFloatStatistics(values)
	
	// パーセンタイルの妥当性チェック
	if result.Percentiles.P25 < 20 || result.Percentiles.P25 > 30 {
		t.Errorf("P25 = %v, expected around 25", result.Percentiles.P25)
	}
	
	if result.Percentiles.P50 < 45 || result.Percentiles.P50 > 55 {
		t.Errorf("P50 = %v, expected around 50", result.Percentiles.P50)
	}
	
	if result.Percentiles.P75 < 70 || result.Percentiles.P75 > 80 {
		t.Errorf("P75 = %v, expected around 75", result.Percentiles.P75)
	}
	
	if result.Percentiles.P95 < 90 || result.Percentiles.P95 > 100 {
		t.Errorf("P95 = %v, expected around 95", result.Percentiles.P95)
	}
}