import React from 'react';
import styles from './DotsGrid.module.css';

const DotsGrid = React.memo(() => {
  return <div className={styles.dotsGridContainer} />;
});

DotsGrid.displayName = 'DotsGrid';

export default DotsGrid;
