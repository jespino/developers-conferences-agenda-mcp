package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

var eventDataURL = "https://developers.events/all-events.json"

// Event represents a developer conference or event
type Event struct {
	Name           string    `json:"name"`
	DateTimestamps []int64   `json:"date"`
	URL            string    `json:"hyperlink"`
	Location       string    `json:"location"`
	City           string    `json:"city"`
	Country        string    `json:"country"`
	Misc           string    `json:"misc"`
	CFP            CFPInfo   `json:"cfp"`
	ClosedCaptions bool      `json:"closedCaptions"`
	Scholarship    bool      `json:"scholarship"`
	Status         string    `json:"status"`
	
	// Computed fields (not directly in JSON)
	StartDate   time.Time `json:"-"`
	EndDate     time.Time `json:"-"`
	CFPEndDate  time.Time `json:"-"`
}

// CFPInfo represents Call for Papers information
type CFPInfo struct {
	Link       string `json:"link"`
	Until      string `json:"until"`
	UntilDate  int64  `json:"untilDate"`
}

// EventData represents the collection of events
type EventData struct {
	Events []Event `json:"events"`
}

// SearchEventsArgs defines parameters for searching events
type SearchEventsArgs struct {
	Query       string `json:"query" jsonschema:"description=Search query for event name or description"`
	Location    string `json:"location" jsonschema:"description=Filter events by location"`
	FromDate    string `json:"fromDate" jsonschema:"description=Filter events starting from this date (YYYY-MM-DD)"`
	ToDate      string `json:"toDate" jsonschema:"description=Filter events up to this date (YYYY-MM-DD)"`
	HasOpenCFP  bool   `json:"hasOpenCFP" jsonschema:"description=Only show events with open CFPs (Call for Papers)"`
	CFPFromDate string `json:"cfpFromDate" jsonschema:"description=Filter events with CFP ending after this date (YYYY-MM-DD)"`
	CFPToDate   string `json:"cfpToDate" jsonschema:"description=Filter events with CFP ending before this date (YYYY-MM-DD)"`
	Limit       int    `json:"limit" jsonschema:"description=Maximum number of events to return"`
}

type LimitArgs struct {
	Limit int `json:"limit" jsonschema:"description=Maximum number of events to return"`
}

type DaysArgs struct {
	Days int `json:"date" jsonschema:"description=Filter events based on number of days before the end of the CFP"`
}

// FetchAndParseEvents retrieves the event data from the URL
// Exported for testing
func FetchAndParseEvents() ([]Event, error) {
	return fetchAndParseEvents()
}

// fetchAndParseEvents retrieves the event data from the URL
// millisToTime converts a millisecond timestamp to time.Time
func millisToTime(millis int64) time.Time {
	if millis == 0 {
		return time.Time{}
	}
	return time.Unix(0, millis*int64(time.Millisecond)).UTC()
}

func fetchAndParseEvents() ([]Event, error) {
	resp, err := http.Get(eventDataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		// If direct unmarshal fails, try with wrapper structure
		var wrappedData EventData
		if err := json.Unmarshal(body, &wrappedData); err != nil {
			return nil, fmt.Errorf("failed to parse event data: %w", err)
		}
		events = wrappedData.Events
	}

	// Process computed fields
	for i := range events {
		// Set start and end dates from the date array
		if len(events[i].DateTimestamps) > 0 {
			events[i].StartDate = millisToTime(events[i].DateTimestamps[0])
			
			// If there's more than one date, use the last one as end date
			if len(events[i].DateTimestamps) > 1 {
				events[i].EndDate = millisToTime(events[i].DateTimestamps[len(events[i].DateTimestamps)-1])
			} else {
				// If only one date, use it for both start and end
				events[i].EndDate = events[i].StartDate
			}
		}
		
		// Process CFP end date
		if events[i].CFP.UntilDate > 0 {
			events[i].CFPEndDate = millisToTime(events[i].CFP.UntilDate)
		}
	}

	return events, nil
}

func main() {
	done := make(chan struct{})

	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	// Register tool to search for events
	err := server.RegisterTool("search_events", "Search for developer conferences and events", func(args SearchEventsArgs) (*mcp_golang.ToolResponse, error) {
		events, err := fetchAndParseEvents()
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error fetching events: %s", err))), nil
		}

		now := time.Now()

		// Parse date filters if provided
		var fromDate, toDate, cfpFromDate, cfpToDate time.Time
		if args.FromDate != "" {
			fromDate, err = time.Parse("2006-01-02", args.FromDate)
			if err != nil {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Invalid fromDate format: %s", err))), nil
			}
		}
		if args.ToDate != "" {
			toDate, err = time.Parse("2006-01-02", args.ToDate)
			if err != nil {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Invalid toDate format: %s", err))), nil
			}
		}
		if args.CFPFromDate != "" {
			cfpFromDate, err = time.Parse("2006-01-02", args.CFPFromDate)
			if err != nil {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Invalid cfpFromDate format: %s", err))), nil
			}
		}
		if args.CFPToDate != "" {
			cfpToDate, err = time.Parse("2006-01-02", args.CFPToDate)
			if err != nil {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Invalid cfpToDate format: %s", err))), nil
			}
		}

		// Filter and search events
		var filteredEvents []Event
		for _, event := range events {
			// Apply filters
			if args.Query != "" && !contains(event.Name+event.Location+event.City+event.Country+event.Misc, args.Query) {
				continue
			}
			if args.Location != "" && !contains(event.Location, args.Location) {
				continue
			}
			if !fromDate.IsZero() && event.StartDate.Before(fromDate) {
				continue
			}
			if !toDate.IsZero() && event.StartDate.After(toDate) {
				continue
			}

			// CFP filters
			if args.HasOpenCFP && (!event.CFPEndDate.After(now) || event.CFP.Link == "") {
				continue
			}
			if !cfpFromDate.IsZero() && event.CFPEndDate.Before(cfpFromDate) {
				continue
			}
			if !cfpToDate.IsZero() && event.CFPEndDate.After(cfpToDate) {
				continue
			}

			filteredEvents = append(filteredEvents, event)

			// Respect limit if set
			if args.Limit > 0 && len(filteredEvents) >= args.Limit {
				break
			}
		}

		// Convert to JSON for response
		eventJSON, err := json.MarshalIndent(filteredEvents, "", "  ")
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error encoding events: %s", err))), nil
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(eventJSON))), nil
	})
	if err != nil {
		panic(err)
	}

	// Register tool for events with open CFPs
	err = server.RegisterTool("open_cfps", "Get events with open CFP (Call for Papers)", func(args LimitArgs) (*mcp_golang.ToolResponse, error) {
		events, err := fetchAndParseEvents()
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error fetching events: %s", err))), nil
		}

		now := time.Now()
		var openCFPEvents []Event

		for _, event := range events {
			// Only include events with open CFPs (CFP deadline in the future and has CFP link)
			if event.CFPEndDate.After(now) && event.CFP.Link != "" {
				openCFPEvents = append(openCFPEvents, event)

				if args.Limit > 0 && len(openCFPEvents) >= args.Limit {
					break
				}
			}
		}

		eventJSON, err := json.MarshalIndent(openCFPEvents, "", "  ")
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error encoding events: %s", err))), nil
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(eventJSON))), nil
	})
	if err != nil {
		panic(err)
	}

	// Register resource for accessing all events
	err = server.RegisterResource("events://all", "all_events", "All developer conferences and events", "application/json", func() (*mcp_golang.ResourceResponse, error) {
		events, err := fetchAndParseEvents()
		if err != nil {
			return nil, err
		}

		eventJSON, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp_golang.NewResourceResponse(mcp_golang.NewTextEmbeddedResource("events://all", string(eventJSON), "application/json")), nil
	})
	if err != nil {
		panic(err)
	}

	// Register resource for accessing events with open CFPs
	err = server.RegisterResource("events://open-cfps", "open_cfps", "Events with open Call for Papers", "application/json", func() (*mcp_golang.ResourceResponse, error) {
		events, err := fetchAndParseEvents()
		if err != nil {
			return nil, err
		}

		now := time.Now()
		var openCFPEvents []Event

		for _, event := range events {
			if event.CFPEndDate.After(now) && event.CFP.Link != "" {
				openCFPEvents = append(openCFPEvents, event)
			}
		}

		eventJSON, err := json.MarshalIndent(openCFPEvents, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp_golang.NewResourceResponse(mcp_golang.NewTextEmbeddedResource("events://open-cfps", string(eventJSON), "application/json")), nil
	})
	if err != nil {
		panic(err)
	}

	// Register tool to get upcoming events
	err = server.RegisterTool("upcoming_events", "Get upcoming developer conferences and events", func(args LimitArgs) (*mcp_golang.ToolResponse, error) {
		events, err := fetchAndParseEvents()
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error fetching events: %s", err))), nil
		}

		now := time.Now()
		var upcomingEvents []Event

		for _, event := range events {
			if event.StartDate.After(now) {
				upcomingEvents = append(upcomingEvents, event)

				if args.Limit > 0 && len(upcomingEvents) >= args.Limit {
					break
				}
			}
		}

		eventJSON, err := json.MarshalIndent(upcomingEvents, "", "  ")
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error encoding events: %s", err))), nil
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(eventJSON))), nil
	})
	if err != nil {
		panic(err)
	}

	// Register tool to get CFP deadlines soon
	err = server.RegisterTool("cfp_deadlines_soon", "Get events with CFP deadlines approaching within days", func(args DaysArgs) (*mcp_golang.ToolResponse, error) {
		if args.Days <= 0 {
			args.Days = 30 // Default to 30 days if not specified
		}

		events, err := fetchAndParseEvents()
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error fetching events: %s", err))), nil
		}

		now := time.Now()
		deadline := now.AddDate(0, 0, args.Days)
		var approachingCFPs []Event

		for _, event := range events {
			// Include events where CFP is still open but deadline is approaching
			if event.CFPEndDate.After(now) && event.CFPEndDate.Before(deadline) && event.CFP.Link != "" {
				approachingCFPs = append(approachingCFPs, event)
			}
		}

		eventJSON, err := json.MarshalIndent(approachingCFPs, "", "  ")
		if err != nil {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Error encoding events: %s", err))), nil
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(eventJSON))), nil
	})
	if err != nil {
		panic(err)
	}

	err = server.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
