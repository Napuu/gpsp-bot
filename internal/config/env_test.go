package config

import "testing"

func TestIsBackgroundFeature(t *testing.T) {
	if !IsBackgroundFeature("daymeme") {
		t.Fatal("expected daymeme to be a background feature")
	}
	if IsBackgroundFeature("ping") {
		t.Fatal("expected ping to be a user command feature")
	}
}
