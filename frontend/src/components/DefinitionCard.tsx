import React from 'react';
import styles from './DefinitionCard.module.css';
import type { SearchResult } from '../types';

interface Props {
  result?: SearchResult | null;
}

const DefinitionCard = React.memo(({ result }: Props) => {
  if (!result) return null;

  return (
    <div className={styles.pipelineDefinitions}>
      <div className={styles.pipelineWordContainer}>
        <h2 className={styles.pipelineWord}>{result.doc_id}</h2>
        {result.phonetic && (
          <span className={styles.pipelinePhonetic}>{result.phonetic}</span>
        )}
      </div>
      
      {result.definitions && result.definitions.length > 0 ? (
        <div className={styles.defList}>
          {result.definitions.map((def, idx) => {
            const hasExample = Array.isArray(result.examples) && idx < result.examples.length && result.examples[idx];
            return (
              <div key={idx} className={styles.defItem}>
                <p className={styles.defText}>• {def}</p>
                {hasExample && (
                  <p className={styles.defExample}>"{result.examples[idx]}"</p>
                )}
              </div>
            );
          })}
        </div>
      ) : (
        <div className={styles.defEmptyState}>
          <p>Definition available online only.</p>
          <p>No local definition found for this word.</p>
        </div>
      )}
    </div>
  );
});

DefinitionCard.displayName = 'DefinitionCard';

export default DefinitionCard;
