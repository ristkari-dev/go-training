package dedup

import "testing"

// TestSeen is a SKELETON. Cover: first sight false then repeat true;
// bounded eviction (oldest evicted past capacity — probe ONE key, since
// Seen mutates state); concurrent Seen under -race.
func TestSeen(t *testing.T) {
	// TODO
	_ = New
}
