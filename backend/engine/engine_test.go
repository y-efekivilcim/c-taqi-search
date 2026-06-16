package engine

import (
	"math"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"", "abc", 3},
		{"abc", "", 3},
		{"same", "same", 0},
	}

	for _, tt := range tests {
		result := LevenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestTokenize(t *testing.T) {
	storage := NewInMemoryStorage()
	eng := NewEngine(storage)

	text := "The quick brown fox is jumping over a lazy dog!"
	valid, dropped := eng.Tokenize(text)

	expectedValid := []string{"quick", "brown", "fox", "jumping", "over", "lazy", "dog"}
	expectedDropped := []string{"the", "is", "a"}

	if len(valid) != len(expectedValid) {
		t.Fatalf("Expected %d valid tokens, got %d", len(expectedValid), len(valid))
	}
	for i, v := range expectedValid {
		if valid[i] != v {
			t.Errorf("Expected valid token %q at %d, got %q", v, i, valid[i])
		}
	}

	if len(dropped) != len(expectedDropped) {
		t.Fatalf("Expected %d dropped tokens, got %d", len(expectedDropped), len(dropped))
	}
	for i, v := range expectedDropped {
		if dropped[i] != v {
			t.Errorf("Expected dropped token %q at %d, got %q", v, i, dropped[i])
		}
	}
}

func TestBM25Scoring(t *testing.T) {
	storage := NewInMemoryStorage()
	eng := NewEngine(storage)

	// Add exact match document
	eng.AddDocument("apple", "apple", []string{"A fruit"}, nil)
	// Add partial match document
	eng.AddDocument("banana", "banana apple", []string{"Also a fruit"}, nil)

	// Since it's a single word dictionary, length normalisation should not overly punish the slightly longer document
	res := eng.Search("apple")

	if len(res.Results) == 0 {
		t.Fatal("Expected results for 'apple', got 0")
	}

	// Exact match "apple" should outscore "banana" significantly because of the exact match multiplier (* 5.0)
	if res.Results[0].DocID != "apple" {
		t.Errorf("Expected top result to be 'apple', got %q", res.Results[0].DocID)
	}
	
	// Verify math contains the query term
	if _, exists := res.Results[0].Math["apple"]; !exists {
		t.Error("Expected math breakdown to contain 'apple' token")
	}
	
	score := res.Results[0].Score
	if math.IsNaN(score) || score <= 0 {
		t.Errorf("Invalid score generated: %f", score)
	}
}

func TestFuzzyMatch(t *testing.T) {
	storage := NewInMemoryStorage()
	eng := NewEngine(storage)

	// Add vocab
	eng.AddDocument("hello", "hello", nil, nil)
	eng.AddDocument("world", "world", nil, nil)
	eng.AddDocument("dictionary", "dictionary", nil, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},       // exact match
		{"helo", "hello"},        // 1 edit distance
		{"dictionry", "dictionary"}, // 1 edit distance
		{"wrold", "world"},       // 2 edit distance (swap)
		{"xyz", "xyz"},           // no match within tolerance, returns input
	}

	for _, tt := range tests {
		result := eng.FuzzyMatch(tt.input)
		if result != tt.expected {
			t.Errorf("FuzzyMatch(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	storage := NewInMemoryStorage()
	eng := NewEngine(storage)

	eng.AddDocument("test", "test", nil, nil)

	// Empty string
	res := eng.Search("")
	if len(res.Results) != 0 {
		t.Errorf("Expected 0 results for empty string, got %d", len(res.Results))
	}

	// Punctuation only
	res = eng.Search("!!! ???")
	if len(res.Results) != 0 {
		t.Errorf("Expected 0 results for punctuation only, got %d", len(res.Results))
	}

	// Case insensitivity
	res = eng.Search("TeSt")
	if len(res.Results) == 0 || res.Results[0].DocID != "test" {
		t.Errorf("Expected result 'test' for query 'TeSt', but did not match")
	}
}
