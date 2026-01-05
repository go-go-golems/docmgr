import type { RefObject } from 'react'

import type { SearchMode } from '../searchSlice'

export function SearchBar({
  mode,
  value,
  onChange,
  inputRef,
}: {
  mode: SearchMode
  value: string
  onChange: (next: string) => void
  inputRef: RefObject<HTMLInputElement | null>
}) {
  return (
    <div className="mb-3">
      <div className="input-group input-group-lg">
        <span className="input-group-text">Search</span>
        <input
          ref={inputRef}
          type="text"
          className="form-control search-input"
          placeholder={
            mode === 'reverse'
              ? 'Search by file path (e.g. backend/api/register.go)'
              : mode === 'files'
                ? 'Search for related files…'
                : 'Search docs…'
          }
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        <button className="btn btn-primary" type="submit">
          Search
        </button>
      </div>
      <div className="keyboard-hint">
        Press <kbd>/</kbd> focus • <kbd>?</kbd> shortcuts • <kbd>Ctrl/Cmd</kbd>+<kbd>R</kbd> refresh •{' '}
        <kbd>Esc</kbd> close
      </div>
    </div>
  )
}
