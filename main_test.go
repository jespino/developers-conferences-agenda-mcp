package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
			events := []Event{
				{
					Name:      "TestConf 2025",
					URL:       "https://testconf.example.com",
					StartDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
					EndDate:   time.Date(2025, 10, 17, 0, 0, 0, 0, time.UTC),
					Location:  "Virtual",
				},
				{
					Name:       "DevTest Summit",
					URL:        "https://devtest.example.com",
					StartDate:  time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
					EndDate:    time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
					Location:   "Berlin, Germany",
					CFPEndDate: time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
					CFPUrl:     "https://devtest.example.com/cfp",
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
		if events[1].CFPUrl != "https://devtest.example.com/cfp" {
			t.Errorf("Expected second event CFP URL to be 'https://devtest.example.com/cfp', got '%s'", events[1].CFPUrl)
		}
	})

	t.Run("Successfully parses wrapped JSON", func(t *testing.T) {
		// Create test server with mock response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events := EventData{
				Events: []Event{
					{
						Name:      "TestConf 2025",
						URL:       "https://testconf.example.com",
						StartDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
						EndDate:   time.Date(2025, 10, 17, 0, 0, 0, 0, time.UTC),
						Location:  "Virtual",
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
