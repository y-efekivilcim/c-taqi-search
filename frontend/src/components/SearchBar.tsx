import React from 'react';
import styles from './SearchBar.module.css';

interface Props {
  onSearch: (query: string) => void;
}

const SearchBar = React.memo(({ onSearch }: Props) => {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const query = formData.get('query');
    if (typeof query === 'string') {
      const trimmed = query.trim();
      if (trimmed) {
        onSearch(trimmed);
      }
    }
  };

  return (
    <form className={styles.searchWrapper} onSubmit={handleSubmit}>
      <input
        type="text"
        name="query"
        className={styles.searchInput}
        placeholder="Search..."
        autoComplete="off"
      />
      <button type="submit" className={styles.searchBtnIcon} title="Search" aria-label="Search">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="11" cy="11" r="8"></circle>
          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
        </svg>
      </button>
    </form>
  );
});

SearchBar.displayName = 'SearchBar';

export default SearchBar;
