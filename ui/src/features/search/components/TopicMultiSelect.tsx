import { useState } from 'react'

export function TopicMultiSelect({
  topics,
  onChange,
}: {
  topics: string[]
  onChange: (topics: string[]) => void
}) {
  const [value, setValue] = useState('')

  const add = () => {
    const next = value.trim()
    if (!next) return
    if (topics.includes(next)) {
      setValue('')
      return
    }
    onChange([...topics, next])
    setValue('')
  }

  return (
    <div>
      <div className="d-flex flex-wrap gap-1 mb-2">
        {topics.map((t) => (
          <span key={t} className="badge text-bg-secondary">
            {t}{' '}
            <button
              type="button"
              className="btn btn-sm btn-link p-0 ms-1 text-white"
              style={{ textDecoration: 'none' }}
              onClick={() => onChange(topics.filter((x) => x !== t))}
            >
              ×
            </button>
          </span>
        ))}
      </div>
      <div className="input-group input-group-sm">
        <input
          className="form-control"
          placeholder="Add topic and press Enter"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              add()
            }
          }}
        />
        <button className="btn btn-outline-secondary" type="button" onClick={add}>
          Add
        </button>
      </div>
    </div>
  )
}

