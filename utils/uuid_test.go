package utils

import "testing"

func TestUUID_ValueAlwaysDifferent(t *testing.T) {
	numUUIDs := 1000
	uuids := make(map[string]bool, numUUIDs)
	for i := 0; i < numUUIDs; i++ {
		uuid := UUID()
		if _, exists := uuids[uuid]; exists {
			t.Errorf("UUID() generated a duplicate UUID: %s", uuid)
		}
		uuids[uuid] = true
	}
}
