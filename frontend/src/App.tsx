import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import './App.css';

import DotsGrid from './components/DotsGrid';
import SearchBar from './components/SearchBar';
import StatusVisualizer from './components/StatusVisualizer';
import DefinitionCard from './components/DefinitionCard';
import MetricsPanel from './components/MetricsPanel';
import ErrorBoundary from './components/ErrorBoundary';
import type { SearchResponse } from './types';

const HIGH_CONFIDENCE_THRESHOLD = 20;

function AppContent() {
  const [searchQuery, setSearchQuery] = useState('');
  const [hasSearched, setHasSearched] = useState(false);

  const { data, error, isLoading, isFetching } = useQuery<SearchResponse, Error>({
    queryKey: ['search', searchQuery],
    queryFn: async () => {
      if (!searchQuery) return { results: [], dropped_stop_words: [], fuzzy_corrections: {}, tokens_used: [] };
      const apiUrl = '/api';
      const response = await fetch(`${apiUrl}/search?q=${encodeURIComponent(searchQuery)}`);
      if (!response.ok) throw new Error('Network response was not ok');
      return response.json();
    },
    enabled: !!searchQuery,
  });

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    setHasSearched(true);
  }, []);

  const isNetworkLoading = isLoading || isFetching;

  let emotion: 'calm' | 'loading' | 'sad' | 'excited' | 'corrected' = 'calm';
  if (isNetworkLoading) {
    emotion = 'loading';
  } else if (data) {
    if (!data.results || data.results.length === 0) {
      emotion = 'sad';
    } else if (Object.keys(data.fuzzy_corrections || {}).length > 0) {
      emotion = 'corrected';
    } else {
      emotion = 'excited';
    }
  }

  return (
    <div className={`app-container ${hasSearched ? 'split' : 'centered'}`}>
      
      <div className="left-panel">
        <DotsGrid />
        <div className="left-content-wrapper">
          <div className="branding">
            <h1>Taqi Search</h1>
          </div>
          <SearchBar onSearch={handleSearch} />
        </div>
      </div>

      <div className="right-panel">
        {isNetworkLoading ? (
          <div className="math-results" style={{ justifyContent: 'center', height: '100%' }}>
            <StatusVisualizer state="loading" />
          </div>
        ) : error ? (
          <div className="math-results">
            <StatusVisualizer state="sad" />
            <div className="search-status-wrapper" style={{ marginTop: '1.5rem' }}>
              <div className="status-header no-match">Network Disconnected</div>
            </div>
          </div>
        ) : data ? (
          <div className="math-results">
            
            {data.results && data.results.length > 0 && (
              <div className="lexical-pipeline-container">
                <StatusVisualizer state={emotion} />
                
                <div className="search-status-wrapper">
                  {!data.results || data.results.length === 0 ? (
                    <div className="status-header no-match">No Match</div>
                  ) : Object.keys(data.fuzzy_corrections || {}).length > 0 ? (
                    <div className="status-header corrected">Taqi Correction: {data.tokens_used.join(" ")}</div>
                  ) : (
                    <div className="status-header exact">Input: {data.tokens_used.join(" ")}</div>
                  )}
                </div>

                <DefinitionCard result={data.results[0]} />
                <MetricsPanel result={data.results[0]} />
              </div>
            )}

            {(!data.results || data.results.length === 0) && (
              <div style={{ paddingTop: '1rem' }}>
                <StatusVisualizer state="sad" />
                <div className="search-status-wrapper" style={{ marginTop: '1.5rem' }}>
                  <div className="status-header no-match">No Match</div>
                </div>
              </div>
            )}
            
          </div>
        ) : (
          <div className="empty-state"></div>
        )}
      </div>

    </div>
  );
}

export default function App() {
  return (
    <ErrorBoundary>
      <AppContent />
    </ErrorBoundary>
  );
}
