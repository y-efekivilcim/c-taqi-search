import React from 'react';
import styles from './MetricsPanel.module.css';
import type { SearchResult } from '../types';

interface Props {
  result?: SearchResult | null;
}

const MetricsPanel = React.memo(({ result }: Props) => {
  if (!result || !result.math) return null;

  return (
    <div className={styles.pipelineMath}>
      <div className={styles.mathBreakdown}>
        <div className={styles.vectorNodeTitle}>
          BM25 Document Score
        </div>
        {Object.entries(result.math).map(([term, stats]) => (
          <div className={styles.mathRow} key={term}>
            <span className={styles.mathTerm}>λ("{term}")</span>
            <div className={styles.mathStats}>
              <span className={styles.statPill}>TF: {stats.tf}</span>
              <span className={styles.statPill}>DF: {stats.df}</span>
              <span className={styles.statPill}>IDF: {stats.idf.toFixed(3)}</span>
              <span className={`${styles.statPill} ${styles.highlight}`}>Score: {stats.term_score.toFixed(3)}</span>
            </div>
          </div>
        ))}
        <div className={styles.mathTotal}>
          Σ BM25 = {result.score.toFixed(3)}
        </div>
      </div>
    </div>
  );
});

MetricsPanel.displayName = 'MetricsPanel';

export default MetricsPanel;
