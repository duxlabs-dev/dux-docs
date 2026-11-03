import React, {useCallback, useEffect, useId, useRef, useState} from 'react';
import {useLocation} from '@docusaurus/router';
import SearchBar from '@theme/SearchBar';
import {Close, Search} from 'material-symbols/rounded';
import {IconButton} from '@/components/ui/icon-button';
import styles from './styles.module.css';

export default function DocsSearch() {
  const dialog = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const location = useLocation();
  const [shortcut, setShortcut] = useState('Ctrl K');
  const open = useCallback(() => {
    if (!dialog.current) return;
    if (!dialog.current.open) dialog.current.showModal();
    const input = dialog.current.querySelector('input');
    input?.focus();
    input?.select();
  }, []);

  useEffect(() => {
    if (navigator.platform.includes('Mac')) setShortcut('⌘ K');
    const handleKey = (event: KeyboardEvent) => {
      if (!(event.ctrlKey || event.metaKey) || event.key.toLowerCase() !== 'k') return;
      event.preventDefault();
      event.stopPropagation();
      open();
    };
    document.addEventListener('keydown', handleKey, true);
    return () => document.removeEventListener('keydown', handleKey, true);
  }, [open]);

  useEffect(() => {
    dialog.current?.close();
  }, [location.key]);

  return (
    <>
      <button type="button" className={styles.trigger} onClick={open} aria-label="Search documentation" aria-haspopup="dialog" aria-keyshortcuts="Control+k Meta+k">
        <Search size={20} />
        <span>Search</span><kbd>{shortcut}</kbd>
      </button>
      <dialog ref={dialog} className={styles.dialog} aria-labelledby={titleId} onClick={event => {
        if (event.target === event.currentTarget) dialog.current?.close();
      }}>
        <div className={styles.panel}>
          <div className={styles.header}>
            <h2 id={titleId}>Search documentation</h2>
            <IconButton label="Close search" variant="text" onClick={() => dialog.current?.close()}>
              <Close />
            </IconButton>
          </div>
          <SearchBar />
        </div>
      </dialog>
    </>
  );
}
