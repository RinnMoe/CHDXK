package canteen

import "testing"

func TestStoreWeightedScore(t *testing.T) {
	weights := DefaultStoreWeights
	cases := []struct {
		name string
		dims Dimensions
		want float64
	}{
		{name: "all five", dims: Dimensions{5, 5, 5, 5, 5}, want: 5},
		{name: "mixed", dims: Dimensions{3, 4, 2, 5, 1}, want: 3.25},
		{name: "minimum", dims: Dimensions{1, 1, 1, 1, 1}, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := weights.Score(tc.dims); got != tc.want {
				t.Fatalf("Score(%v) = %v, want %v", tc.dims, got, tc.want)
			}
		})
	}
}

func TestDishWeightedScore(t *testing.T) {
	got := DefaultDishWeights.Score(DishDimensions{Taste: 4, Portion: 3, Value: 5})
	// 0.5*4 + 0.25*3 + 0.25*5 = 4.0
	if got != 4 {
		t.Fatalf("Score = %v, want 4", got)
	}
}

func TestRecentWeights(t *testing.T) {
	if got := RecentWeights(0); got != nil {
		t.Fatalf("RecentWeights(0) = %v, want nil", got)
	}
	if got := RecentWeights(1); len(got) != 1 || got[0] != 0.6 {
		t.Fatalf("RecentWeights(1) = %v", got)
	}
	if got := RecentWeights(2); len(got) != 2 || got[0] != 0.6 || got[1] != 0.3 {
		t.Fatalf("RecentWeights(2) = %v", got)
	}
	if got := RecentWeights(9); len(got) != MaxRecentEvents {
		t.Fatalf("RecentWeights(9) length = %d, want %d", len(got), MaxRecentEvents)
	}
}

func TestWeightedScore(t *testing.T) {
	cases := []struct {
		name    string
		scores  []float64
		weights []float64
		want    float64
	}{
		{name: "single event keeps its score", scores: []float64{4}, weights: RecentWeights(1), want: 4},
		{name: "two events renormalize", scores: []float64{5, 3}, weights: RecentWeights(2), want: 4.33},
		{name: "three events", scores: []float64{5, 4, 3}, weights: RecentWeights(3), want: 4.5},
		{name: "extra scores ignored", scores: []float64{5, 4, 3, 1}, weights: RecentWeights(3), want: 4.5},
		{name: "empty", scores: nil, weights: nil, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WeightedScore(tc.scores, tc.weights); got != tc.want {
				t.Fatalf("WeightedScore = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBayesScore(t *testing.T) {
	cases := []struct {
		name      string
		avg       float64
		count     int
		globalAvg float64
		prior     int
		want      float64
	}{
		{name: "no ratings", avg: 0, count: 0, globalAvg: 3, prior: 5, want: 0},
		{name: "no prior", avg: 4.2, count: 3, globalAvg: 3, prior: 0, want: 4.2},
		{name: "shrunk toward global", avg: 4, count: 9, globalAvg: 3, prior: 5, want: 3.64},
		{name: "single rating shrunk", avg: 5, count: 1, globalAvg: 0, prior: 5, want: 0.83},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BayesScore(tc.avg, tc.count, tc.globalAvg, tc.prior); got != tc.want {
				t.Fatalf("BayesScore = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRatingConfigNormalized(t *testing.T) {
	got := RatingConfig{}.Normalized()
	if got.PriorCount != DefaultPriorCount || got.RefreshCron != DefaultRatingConfig.RefreshCron {
		t.Fatalf("Normalized() = %+v", got)
	}
	custom := RatingConfig{PriorCount: 8, RefreshCron: "0 1 * * *"}.Normalized()
	if custom.PriorCount != 8 || custom.RefreshCron != "0 1 * * *" {
		t.Fatalf("Normalized() = %+v", custom)
	}
}
