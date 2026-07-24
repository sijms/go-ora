package types

import (
	"testing"
	"time"
)

func TestIsEqualLoc(t *testing.T) {
	testScenarios := []struct {
		description string
		loc1        *time.Location
		loc2        *time.Location
		expectEqual bool
	}{
		{
			description: "Equal locations",
			loc1:        time.FixedZone("UTC", 0),
			loc2:        time.FixedZone("UTC", 0),
			expectEqual: true,
		},
		{
			description: "Different locations",
			loc1:        time.FixedZone("UTC", 0),
			loc2:        time.FixedZone("America/New_York", -5*60*60),
			expectEqual: false,
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			result := isEqualLoc(scenario.loc1, scenario.loc2)
			if result != scenario.expectEqual {
				t.Errorf("expected %v, got %v", scenario.expectEqual, result)
			}
		})
	}
}
