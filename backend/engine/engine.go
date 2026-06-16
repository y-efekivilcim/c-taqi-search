package engine

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Document struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Definitions []string `json:"definitions"`
	Examples    []string `json:"examples"`
	Length      int
}

type MathBreakdown struct {
	TF        int     `json:"tf"`
	DF        int     `json:"df"`
	IDF       float64 `json:"idf"`
	TermScore float64 `json:"term_score"`
}

type SearchResult struct {
	DocID       string                   `json:"doc_id"`
	Content     string                   `json:"content"`
	Definitions []string                 `json:"definitions"`
	Examples    []string                 `json:"examples"`
	Phonetic    string                   `json:"phonetic"`
	Score       float64                  `json:"score"`
	Math        map[string]MathBreakdown `json:"math"`
}

type SearchResponse struct {
	Results          []SearchResult    `json:"results"`
	DroppedStopWords []string          `json:"dropped_stop_words"`
	FuzzyCorrections map[string]string `json:"fuzzy_corrections"`
	TokensUsed       []string          `json:"tokens_used"`
}

type StorageBackend interface {
	AddDocument(doc *Document)
	GetDocument(id string) *Document
	GetPostingList(token string) map[string]int
	GetDocCount() int
	GetTotalTokens() int
	AddToInvertedIndex(token, docID string, tf int)
	AddVocab(word string)
	VocabExists(word string) bool
	SearchBKTree(word string, tolerance int) []string
}

type BKNode struct {
	Word     string
	Children map[int]*BKNode
}

type InMemoryStorage struct {
	mu            sync.RWMutex
	Documents     map[string]*Document
	InvertedIndex map[string]map[string]int
	Vocab         map[string]bool
	DocCount      int
	TotalTokens   int
	BKRoot        *BKNode
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Documents:     make(map[string]*Document),
		InvertedIndex: make(map[string]map[string]int),
		Vocab:         make(map[string]bool),
	}
}

func (s *InMemoryStorage) AddDocument(doc *Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Documents[doc.ID] = doc
	s.DocCount++
	s.TotalTokens += doc.Length
}

func (s *InMemoryStorage) GetDocument(id string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Documents[id]
}

func (s *InMemoryStorage) GetPostingList(token string) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pl := s.InvertedIndex[token]
	if pl == nil {
		return nil
	}
	res := make(map[string]int, len(pl))
	for k, v := range pl {
		res[k] = v
	}
	return res
}

func (s *InMemoryStorage) AddToInvertedIndex(token, docID string, tf int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.InvertedIndex[token] == nil {
		s.InvertedIndex[token] = make(map[string]int)
	}
	s.InvertedIndex[token][docID] += tf
}

func (s *InMemoryStorage) GetDocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.DocCount
}

func (s *InMemoryStorage) GetTotalTokens() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TotalTokens
}

func (s *InMemoryStorage) AddVocab(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Vocab[word] = true
	
	if s.BKRoot == nil {
		s.BKRoot = &BKNode{Word: word, Children: make(map[int]*BKNode)}
		return
	}
	
	curr := s.BKRoot
	for {
		dist := LevenshteinDistance(curr.Word, word)
		if dist == 0 {
			break
		}
		if child, exists := curr.Children[dist]; exists {
			curr = child
		} else {
			curr.Children[dist] = &BKNode{Word: word, Children: make(map[int]*BKNode)}
			break
		}
	}
}

func (s *InMemoryStorage) VocabExists(word string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Vocab[word]
}

func (s *InMemoryStorage) SearchBKTree(word string, tolerance int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []string
	if s.BKRoot == nil {
		return results
	}
	
	var search func(node *BKNode)
	search = func(node *BKNode) {
		dist := LevenshteinDistance(node.Word, word)
		if dist <= tolerance {
			results = append(results, node.Word)
		}
		
		for d, child := range node.Children {
			if d >= dist-tolerance && d <= dist+tolerance {
				search(child)
			}
		}
	}
	search(s.BKRoot)
	return results
}

type Engine struct {
	Storage          StorageBackend
	StopWords        map[string]bool
	PunctuationRegex *regexp.Regexp
}

func NewEngine(storage StorageBackend) *Engine {
	stopWordsList := []string{
		"the", "and", "is", "in", "to", "of", "a", "it", "for", "on", "with", "as", "by", "this", "an", "be", "that", "are",
	}

	stopMap := make(map[string]bool)
	for _, w := range stopWordsList {
		stopMap[w] = true
	}

	return &Engine{
		Storage:          storage,
		StopWords:        stopMap,
		PunctuationRegex: regexp.MustCompile(`[^\w\s]`),
	}
}

func (e *Engine) Tokenize(text string) ([]string, []string) {
	text = strings.ToLower(text)
	text = e.PunctuationRegex.ReplaceAllString(text, "")
	words := strings.Fields(text)

	var valid []string
	var dropped []string

	for _, w := range words {
		if e.StopWords[w] {
			dropped = append(dropped, w)
		} else {
			valid = append(valid, w)
		}
	}
	return valid, dropped
}

func LevenshteinDistance(s, t string) int {
	m, n := len(s), len(t)
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}
	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)
		}
	}
	return d[m][n]
}

func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}

func (e *Engine) FuzzyMatch(word string) string {
	if e.Storage.VocabExists(word) {
		return word
	}
	
	candidates := e.Storage.SearchBKTree(word, 2)
	if len(candidates) == 0 {
		return word
	}

	bestMatch := word
	minDist := math.MaxInt32
	maxPrefix := -1

	lcp := func(a, b string) int {
		n := len(a)
		if len(b) < n {
			n = len(b)
		}
		for i := 0; i < n; i++ {
			if a[i] != b[i] {
				return i
			}
		}
		return n
	}

	for _, v := range candidates {
		dist := LevenshteinDistance(word, v)
		if dist < minDist {
			minDist = dist
			bestMatch = v
			maxPrefix = lcp(word, v)
		} else if dist == minDist && dist > 0 {
			p := lcp(word, v)
			if p > maxPrefix {
				maxPrefix = p
				bestMatch = v
			}
		}
	}

	if minDist <= 2 {
		return bestMatch
	}
	return word
}

func (e *Engine) AddDocument(id, content string, definitions, examples []string) {
	tokens, _ := e.Tokenize(content)
	doc := &Document{
		ID:          id,
		Content:     content,
		Definitions: definitions,
		Examples:    examples,
		Length:      len(tokens),
	}
	e.Storage.AddDocument(doc)

	for _, token := range tokens {
		e.Storage.AddToInvertedIndex(token, id, 1)
	}
	
	idToken := strings.ToLower(id)
	e.Storage.AddVocab(idToken)
}

func (e *Engine) Search(query string) SearchResponse {
	tokens, dropped := e.Tokenize(query)
	var finalTokens []string
	fuzzyCorrections := make(map[string]string)

	for _, t := range tokens {
		corrected := e.FuzzyMatch(t)
		if corrected != t {
			fuzzyCorrections[t] = corrected
		}
		finalTokens = append(finalTokens, corrected)
	}

	k1 := 1.5 // TODO: tune BM25 params later
	b := 0.0
	docCount := e.Storage.GetDocCount()
	N := float64(docCount)
	avgDocLength := 1.0
	if docCount > 0 {
		avgDocLength = float64(e.Storage.GetTotalTokens()) / float64(docCount)
	}

	docScores := make(map[string]float64)
	docMath := make(map[string]map[string]MathBreakdown)

	for _, token := range finalTokens {
		postingList := e.Storage.GetPostingList(token)
		df := float64(len(postingList))

		if df == 0 {
			continue
		}

		idf := math.Log((N-df+0.5)/(df+0.5) + 1.0)
		if idf < 0 {
			idf = 0
		}

		for docID, tfRaw := range postingList {
			tf := float64(tfRaw)
			
			docObj := e.Storage.GetDocument(docID)
			if docObj == nil {
				continue
			}
			docLen := float64(docObj.Length)

			numerator := tf * (k1 + 1)
			denominator := tf + k1*(1-b+b*(docLen/avgDocLength))
			termScore := idf * (numerator / denominator)

			if token == strings.ToLower(docID) {
				termScore *= 5.0
			}

			docScores[docID] += termScore

			if docMath[docID] == nil {
				docMath[docID] = make(map[string]MathBreakdown)
			}
			docMath[docID][token] = MathBreakdown{
				TF:        tfRaw,
				DF:        int(df),
				IDF:       math.Round(idf*1000) / 1000,
				TermScore: math.Round(termScore*1000) / 1000,
			}
		}
	}

	var results []SearchResult
	for docID, score := range docScores {
		docObj := e.Storage.GetDocument(docID)
		if docObj == nil {
			continue
		}
		results = append(results, SearchResult{
			DocID:       docID,
			Content:     docObj.Content,
			Definitions: docObj.Definitions,
			Examples:    docObj.Examples,
			Score:       math.Round(score*1000) / 1000,
			Math:        docMath[docID],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > 10 {
		results = results[:10]
	}

	droppedSet := make(map[string]bool)
	var uniqueDropped []string
	for _, d := range dropped {
		if !droppedSet[d] {
			droppedSet[d] = true
			uniqueDropped = append(uniqueDropped, d)
		}
	}

	return SearchResponse{
		Results:          results,
		DroppedStopWords: uniqueDropped,
		FuzzyCorrections: fuzzyCorrections,
		TokensUsed:       finalTokens,
	}
}
