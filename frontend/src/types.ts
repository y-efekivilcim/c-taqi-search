export interface MathBreakdown {
  tf: number;
  df: number;
  idf: number;
  term_score: number;
}

export interface SearchResult {
  doc_id: string;
  content: string;
  definitions: string[];
  examples: string[];
  phonetic: string;
  score: number;
  math: Record<string, MathBreakdown>;
}

export interface SearchResponse {
  results: SearchResult[];
  dropped_stop_words: string[];
  fuzzy_corrections: Record<string, string>;
  tokens_used: string[];
}
