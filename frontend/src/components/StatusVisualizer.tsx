import React from 'react';
import styles from './StatusVisualizer.module.css';

interface Props {
  state: 'calm' | 'loading' | 'sad' | 'excited' | 'corrected';
}

const StatusVisualizer = React.memo(({ state }: Props) => {
  return (
    <div className={styles.statusContainer}>
      {state === 'loading' && (
        <div className={styles.wrapper}>
          {[...Array(6)].map((_, i) => (
            <div key={i} className={styles.miniCircle} style={{ animationDelay: `${i * 0.3}s` }}></div>
          ))}
        </div>
      )}
      {state === 'corrected' && (
        <div className={styles.wrapper}>
          <div className={`${styles.sweep} ${styles.sweepLeft}`}></div>
          <div className={`${styles.sweep} ${styles.sweepRight}`}></div>
        </div>
      )}
      {state === 'excited' && (
        <div className={styles.wrapper}>
          <div className={styles.curlyWaveContainer}>
            <svg className={`${styles.curlyWave} ${styles.wave1}`} viewBox="0 0 200 60" preserveAspectRatio="none">
              <path d="M0,30 Q25,50 50,30 T100,30 T150,30 T200,30" fill="none" stroke="var(--accent-orange)" strokeWidth="3" strokeLinecap="round" />
            </svg>
            <svg className={`${styles.curlyWave} ${styles.wave2}`} viewBox="0 0 200 60" preserveAspectRatio="none">
              <path d="M0,30 Q25,10 50,30 T100,30 T150,30 T200,30" fill="none" stroke="var(--accent-orange)" strokeWidth="2" strokeLinecap="round" />
            </svg>
          </div>
        </div>
      )}
      {state === 'sad' && (
        <div className={styles.wrapper}>
          <div className={styles.singularSweep}></div>
        </div>
      )}
      {state === 'calm' && (
        <div className={styles.wrapper}>
          <div className={styles.calmSweep}></div>
        </div>
      )}
    </div>
  );
});

StatusVisualizer.displayName = 'StatusVisualizer';

export default StatusVisualizer;
