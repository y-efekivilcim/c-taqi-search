package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"taqi-search/engine"
	"time"
)

type FreeDictEntry struct {
	Word     string `json:"word"`
	Phonetic string `json:"phonetic"`
	Meanings []struct {
		Definitions []struct {
			Definition string `json:"definition"`
			Example    string `json:"example"`
		} `json:"definitions"`
	} `json:"meanings"`
}

var (
	httpClient = &http.Client{
		Timeout: 1500 * time.Millisecond,
	}
	searchEngine *engine.Engine
	logger       *slog.Logger
)

func enrichResultWithModernAPI(ctx context.Context, res *engine.SearchResult) {
	url := "https://api.dictionaryapi.dev/api/v2/entries/en/" + res.DocID
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request", "doc_id", res.DocID, "error", err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var entries []FreeDictEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return
	}

	if len(entries) == 0 {
		return
	}

	entry := entries[0]

	var newDefs []string
	var newExs []string

	for _, m := range entry.Meanings {
		for _, d := range m.Definitions {
			if d.Definition != "" {
				newDefs = append(newDefs, d.Definition)
			}
			if d.Example != "" {
				newExs = append(newExs, d.Example)
			}
		}
	}

	if len(newDefs) > 0 {
		res.Definitions = newDefs
		res.Examples = make([]string, 0, len(newExs))
		res.Examples = append(res.Examples, newExs...)
		res.Phonetic = entry.Phonetic
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	query := r.URL.Query().Get("q")
	
	if query == "" {
		http.Error(w, "missing query parameter 'q'", http.StatusBadRequest)
		logger.Warn("Bad Request: Missing query parameter")
		return
	}

	response := searchEngine.Search(query)

	if len(response.Results) > 0 {
		enrichResultWithModernAPI(r.Context(), &response.Results[0])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	logger.Info("Search Query Executed",
		slog.String("query", query),
		slog.Duration("latency", time.Since(start)),
		slog.Int("results_count", len(response.Results)),
	)
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("init engine")
	
	storage := engine.NewInMemoryStorage()
	searchEngine = engine.NewEngine(storage)

	logger.Info("loading corpus")
	if err := LoadDictionaryAsCorpus(searchEngine); err != nil {
		logger.Error("Failed to load dictionary as corpus", "error", err)
		os.Exit(1)
	}
	logger.Info("Boot complete", 
		slog.Int("total_documents", searchEngine.Storage.GetDocCount()),
		slog.Int("total_tokens", searchEngine.Storage.GetTotalTokens()),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/search", corsMiddleware(searchHandler))

	port := "8081"
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Info("Server listening", slog.String("port", port))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
