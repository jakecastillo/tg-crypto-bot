package indicators

import "testing"

func TestComputeRSI(t *testing.T) {
	data := Series{44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28}
	result, err := ComputeRSI(data, 14)
	if err != nil {
		t.Fatalf("rsi failed: %v", err)
	}
	if diff := abs(result.Value - 70.532789); diff > 0.01 {
		t.Fatalf("unexpected rsi %.4f", result.Value)
	}
}

func TestComputeMACD(t *testing.T) {
	data := Series{22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24, 22.29, 22.15, 22.39, 22.38, 22.61, 23.36, 24.05, 23.75, 23.83, 23.95, 23.63, 23.82, 23.87, 23.65, 23.19, 23.10, 23.33, 22.68, 23.10, 22.40, 22.17}
	macd, err := ComputeMACD(data, 12, 26, 9)
	if err != nil {
		t.Fatalf("macd failed: %v", err)
	}
	if diff := abs(macd.MACD - (-0.219)); diff > 0.05 {
		t.Fatalf("unexpected macd %.4f", macd.MACD)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
