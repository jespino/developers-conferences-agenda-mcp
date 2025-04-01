package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Use millisToTime from main package

func TestContains(t *testing.T) {
	testCases := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Go Conference", "conference", true},
		{"Berlin, Germany", "germany", true},
		{"DevOps Summit", "frontend", false},
		{"", "test", false},
		{"Test", "", true},
		{"UPPERCASE", "uppercase", true},
		{"lowercase", "LOWERCASE", true},
	}

	for _, tc := range testCases {
		result := contains(tc.s, tc.substr)
		if result != tc.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", 
				tc.s, tc.substr, result, tc.expected)
		}
	}
}

func TestFetchAndParseEvents(t *testing.T) {
	// Save original URL and restore after tests
	originalURL := eventDataURL
	defer func() { 
		// Restore the original URL after all tests in this function
		eventDataURL = originalURL 
	}()

	t.Run("Successfully parses direct array JSON", func(t *testing.T) {
		// Create test server with mock response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events := []map[string]interface{}{
				{
					"name":      "TestConf 2025",
					"hyperlink": "https://testconf.example.com",
					"date":      []int64{1762752000000}, // Oct 15, 2025
					"location":  "Virtual",
					"city":      "Virtual",
					"country":   "Online",
					"cfp":       map[string]interface{}{},
					"status":    "open",
				},
				{
					"name":      "DevTest Summit",
					"hyperlink": "https://devtest.example.com",
					"date":      []int64{1764048000000, 1764220800000}, // Nov 1-3, 2025
					"location":  "Berlin, Germany",
					"city":      "Berlin",
					"country":   "Germany",
					"cfp": map[string]interface{}{
						"link":       "https://devtest.example.com/cfp",
						"until":      "15-August-2025",
						"untilDate":  1754323200000, // Aug 15, 2025
					},
					"status":    "open",
				},
			}
			json.NewEncoder(w).Encode(events)
		}))
		defer server.Close()

		// Set test server URL for the test
		eventDataURL = server.URL

		// Test function
		events, err := FetchAndParseEvents()
		if err != nil {
			t.Fatalf("fetchAndParseEvents returned error: %v", err)
		}

		// Assertions
		if len(events) != 2 {
			t.Errorf("Expected 2 events, got %d", len(events))
		}
		if events[0].Name != "TestConf 2025" {
			t.Errorf("Expected first event name to be 'TestConf 2025', got '%s'", events[0].Name)
		}
		if events[1].CFP.Link != "https://devtest.example.com/cfp" {
			t.Errorf("Expected second event CFP URL to be 'https://devtest.example.com/cfp', got '%s'", events[1].CFP.Link)
		}
		// Check the computed fields
		expectedStartDate := millisToTime(1764048000000)
		if !events[1].StartDate.Equal(expectedStartDate) {
			t.Errorf("Expected second event start date to be %v, got %v", expectedStartDate, events[1].StartDate)
		}
	})

	t.Run("Successfully parses wrapped JSON", func(t *testing.T) {
		// Create test server with mock response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events := map[string]interface{}{
				"events": []map[string]interface{}{
					{
						"name":      "TestConf 2025",
						"hyperlink": "https://testconf.example.com",
						"date":      []int64{1762752000000}, // Oct 15, 2025
						"location":  "Virtual",
						"city":      "Virtual",
						"country":   "Online",
						"cfp":       map[string]interface{}{},
						"status":    "open",
					},
				},
			}
			json.NewEncoder(w).Encode(events)
		}))
		defer server.Close()

		// Set test server URL for the test
		eventDataURL = server.URL

		// Test function
		events, err := FetchAndParseEvents()
		if err != nil {
			t.Fatalf("fetchAndParseEvents returned error: %v", err)
		}

		// Assertions
		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}
		if events[0].Name != "TestConf 2025" {
			t.Errorf("Expected event name to be 'TestConf 2025', got '%s'", events[0].Name)
		}
	})

	t.Run("Handles HTTP errors", func(t *testing.T) {
		// Create test server with error response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		// Set test server URL for the test
		eventDataURL = server.URL

		// Test function
		_, err := FetchAndParseEvents()
		if err == nil {
			t.Fatal("Expected error for HTTP 500 status, got nil")
		}
	})

	t.Run("Handles malformed JSON", func(t *testing.T) {
		// Create test server with malformed JSON
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"events": [{"name": "Test", "broken json`))
		}))
		defer server.Close()

		// Set test server URL for the test
		eventDataURL = server.URL

		// Test function
		_, err := FetchAndParseEvents()
		if err == nil {
			t.Fatal("Expected error for malformed JSON, got nil")
		}
	})
}

func TestMillisToTime(t *testing.T) {
	testCases := []struct {
		millis   int64
		expected time.Time
	}{
		{0, time.Time{}},
		{1554323200000, time.Date(2019, 4, 3, 20, 26, 40, 0, time.UTC)},
		{1762752000000, time.Date(2025, 11, 10, 5, 20, 0, 0, time.UTC)},
	}

	for _, tc := range testCases {
		result := millisToTime(tc.millis)
		// Compare Unix timestamps since time zones might differ
		if result.Unix() != tc.expected.Unix() {
			t.Errorf("millisToTime(%d) = %v, want %v", tc.millis, result, tc.expected)
		}
	}
}

func TestFetchRealEvents(t *testing.T) {
	// This test fetches data from the real endpoint
	events, err := FetchAndParseEvents()
	if err != nil {
		t.Fatalf("Error fetching real events: %v", err)
	}
	
	// Verify we have events
	if len(events) == 0 {
		t.Error("Expected to receive events, but got empty array")
	}
	
	// Count events with missing fields
	emptyLocationCount := 0
	
	// Verify basic data structure
	for i, event := range events {
		if event.Name == "" {
			t.Errorf("Event %d has empty name", i)
		}
		if event.URL == "" {
			t.Errorf("Event %d has empty URL", i)
		}
		if event.StartDate.IsZero() {
			t.Errorf("Event %d has zero StartDate", i)
		}
		if event.EndDate.IsZero() {
			t.Errorf("Event %d has zero EndDate", i)
		}
		if event.Location == "" {
			// Count but don't fail test
			emptyLocationCount++
		}
	}
	
	if emptyLocationCount > 0 {
		t.Logf("Warning: %d events have empty Location field", emptyLocationCount)
	}
	
	t.Logf("Successfully fetched and parsed %d events from real endpoint", len(events))
}
