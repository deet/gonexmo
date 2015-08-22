package nexmo

import (
	"strings"
	"testing"
	"time"
)

func TestSearchForAvailableNumber(t *testing.T) {
	time.Sleep(1 * time.Second) // Sleep 1 second due to API limitation

	nexmo, err := NewClientFromAPI(API_KEY, API_SECRET)
	if err != nil {
		t.Error("Failed to create Client with error:", err)
	}

	resp, err := nexmo.Numbers.SearchAvailable("US")
	if err != nil {
		t.Error("Unexpected number search error:", err)
	} else {
		if resp.Count == 0 {
			t.Error("Should have found at least one number")
		}
		if len(resp.Numbers) == 0 {
			t.Error("Should have data for at least one number")
		}
		for _, numInfo := range resp.Numbers {
			if numInfo.Country == "" {
				t.Error("Available number should have country set")
			}
			if numInfo.MSISDN == "" {
				t.Error("Available number should have MSISDN set")
			}
			if numInfo.Type == "" {
				t.Error("Available number should have type set")
			}
			if numInfo.Cost < 0.0001 {
				t.Error("Available should have cost set, and numbers are not free")
			}
		}
	}

	// This test could break if Nexmo's 1985 supply is exhausted
	pattern := "1985"
	resp2, err := nexmo.Numbers.SearchAvailableWithOptions("US", NumberSearchOptions{Pattern: pattern, SearchPattern: "0"})
	if err != nil {
		t.Error("Unexpected number search error:", err)
	} else {
		if resp2.Count == 0 {
			t.Error("Should have found at least one number in", pattern)
		}
		if len(resp2.Numbers) == 0 {
			t.Error("Should have data for at least one number in", pattern)
		}
		for _, numInfo := range resp2.Numbers {
			if !strings.Contains(numInfo.MSISDN, pattern) {
				t.Error("Available number should have MSISDN contain the pattern", pattern, "but was:", numInfo.MSISDN)
			}
		}
	}
}

func TestBuyAvailableNumber(t *testing.T) {

}
