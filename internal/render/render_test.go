package render

import "testing"

func TestCodeToConditionMapsWMOCodes(t *testing.T) {
	cases := map[int]string{
		0:  "Clear",
		1:  "Clear",
		2:  "Clouds",
		3:  "Clouds",
		45: "Fog",
		48: "Fog",
		63: "Rain",
		80: "Rain",
		82: "Rain",
		71: "Snow",
		85: "Snow",
		95: "Thunderstorm",
		99: "Thunderstorm",
	}

	for code, want := range cases {
		if got := CodeToCondition(code); got != want {
			t.Errorf("CodeToCondition(%d) = %q, want %q", code, got, want)
		}
	}
}

func TestCodeToConditionFallsBackToClouds(t *testing.T) {
	if got := CodeToCondition(-1); got != "Clouds" {
		t.Errorf("CodeToCondition(-1) = %q, want %q", got, "Clouds")
	}
}
