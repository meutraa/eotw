package score

import (
	"math"
	"testing"
	"time"
)

var result time.Duration

func BenchmarkDistance(b *testing.B) {
	s := DefaultScorer{}
	total := time.Millisecond * 0
	p, q := time.Millisecond*12456, time.Millisecond*13456
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		total += s.Distance(140, p, q)
	}

	result = total
}

func BenchmarkDistanceFloat(b *testing.B) {
	s := DefaultScorer{}
	total := time.Millisecond * 0
	p, q := time.Millisecond*12456, time.Millisecond*13456
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		total += s.DistanceFloat(140, p, q)
	}

	result = total
}

type distanceTest struct {
	Rate            uint16
	Error           time.Duration
	HitTime         time.Duration
	ExpectedHitTime time.Duration
}

func createDistanceTests() []distanceTest {
	tests := []distanceTest{}
	for i := -500; i < 500; i++ {
		if i == 0 {
			continue
		}
		rate := i
		if rate < 0 {
			rate = -rate
		}
		test := distanceTest{
			Rate:            uint16(rate),
			Error:           time.Duration(i) * time.Millisecond,
			ExpectedHitTime: abs(100 * time.Duration(i) * time.Millisecond),
		}
		// If the ExpectedHitTime was 2000ms
		// And the rate is 200, we expect the user to hit at 1000ms
		// If the user HitTime at 950ms, this is an Error of 50ms

		// All we need to do is divide 100*ExpectedHitTime by Rate
		test.HitTime = time.Duration(math.Round(100*float64(test.ExpectedHitTime)/float64(test.Rate))) - test.Error
		tests = append(tests, test)
	}
	return tests
}

func TestDistance(t *testing.T) {
	scorer := DefaultScorer{}
	tests := createDistanceTests()
	for _, test := range tests {
		err := scorer.Distance(test.Rate, test.ExpectedHitTime, test.HitTime)
		if err != test.Error {
			t.Log("           Rate:", test.Rate)
			t.Log("ExpectedHitTime:", test.ExpectedHitTime)
			t.Log("AdjExpectedTime:", time.Duration(math.Round(100*float64(test.ExpectedHitTime)/float64(test.Rate))))
			t.Log("  ActualHitTime:", test.HitTime)
			t.Log("Calculated Error", err)
			t.Log("  Expected Error", test.Error)
			t.Log("")
			t.Fail()
		}
	}
}

func TestDistanceFloat(t *testing.T) {
	scorer := DefaultScorer{}
	tests := createDistanceTests()
	for _, test := range tests {
		err := scorer.DistanceFloat(test.Rate, test.ExpectedHitTime, test.HitTime)
		if err != test.Error {
			t.Log("           Rate:", test.Rate)
			t.Log("ExpectedHitTime:", test.ExpectedHitTime)
			t.Log("AdjExpectedTime:", time.Duration(math.Round(100*float64(test.ExpectedHitTime)/float64(test.Rate))))
			t.Log("  ActualHitTime:", test.HitTime)
			t.Log("Calculated Error", err)
			t.Log("  Expected Error", test.Error)
			t.Log("")
			t.Fail()
		}
	}
}
