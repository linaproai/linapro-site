import React from 'react';
import TOC from '@theme-original/TOC';
import styles from './styles.module.css';

export default function TOCWrapper(props) {
  return (
    <div className={styles.tocContainer}>
      <TOC {...props} />
    </div>
  );
}
