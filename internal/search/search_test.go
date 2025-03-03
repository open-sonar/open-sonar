package search

import (
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    // Set test mode
    os.Setenv("TEST_MODE", "true")

    // Run tests
    exitCode := m.Run()

    // Exit
    os.Exit(exitCode)
}

func TestSearch(t *testing.T) {
    tests := []struct {
        name        string
        query       string
        domains     []string
        expectCount int
    }{
        {
            name:        "Basic search",
            query:       "test query",
            domains:     nil,
            expectCount: 10,
        },
        {
            name:        "Empty search",
            query:       "",
            domains:     nil,
            expectCount: 0,
        },
        {
            name:        "With domain filter",
            query:       "government data",
            domains:     []string{".gov"},
            expectCount: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results := RunSearch(tt.query, 3, 1, tt.domains)

            t.Logf("Test %q returned %d results", tt.name, len(results))

            for i, r := range results {
                t.Logf("  Result %d: URL=%s", i+1, r.URL)
            }

            if len(results) <tt.expectCount {
                t.Errorf("Expected %d results, got %d", tt.expectCount, len(results))
            }
        })
    }
}
