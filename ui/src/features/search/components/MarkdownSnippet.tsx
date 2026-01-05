import { cloneElement, isValidElement, useMemo } from 'react'
import type { JSX as ReactJSX, ReactElement, ReactNode } from 'react'

import ReactMarkdown from 'react-markdown'
import type { Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function extractHighlightTerms(raw: string): string[] {
  const q = raw.trim()
  if (!q) return []

  const terms: string[] = []
  const used = new Set<string>()

  const quoted = q.matchAll(/"([^"]+)"/g)
  for (const m of quoted) {
    const v = (m[1] ?? '').trim()
    if (!v) continue
    const key = v.toLowerCase()
    if (used.has(key)) continue
    used.add(key)
    terms.push(v)
  }

  const cleaned = q.replace(/"[^"]*"/g, ' ')
  for (const tok of cleaned.split(/\s+/g)) {
    const t = tok.trim().replace(/^[^a-zA-Z0-9]+|[^a-zA-Z0-9]+$/g, '')
    if (!t) continue
    const lower = t.toLowerCase()
    if (lower === 'and' || lower === 'or' || lower === 'not') continue
    if (used.has(lower)) continue
    used.add(lower)
    terms.push(t)
  }

  return terms.slice(0, 8)
}

function highlightReactNode(node: ReactNode, re: RegExp | null, inCode = false): ReactNode {
  if (!re) return node

  if (node == null || typeof node === 'boolean') return node

  if (Array.isArray(node)) {
    return node.map((child) => highlightReactNode(child, re, inCode))
  }

  if (typeof node === 'string' || typeof node === 'number') {
    if (inCode) return node
    const parts = String(node).split(re)
    if (parts.length <= 1) return node
    return parts.map((p, idx) => (idx % 2 === 1 ? <mark key={idx}>{p}</mark> : p))
  }

  if (isValidElement(node)) {
    const el = node as ReactElement
    const type = el.type
    const nextInCode = inCode || type === 'code' || type === 'pre'
    const props = el.props as { children?: ReactNode }
    const nextChildren = highlightReactNode(props.children, re, nextInCode)
    return cloneElement(el, undefined, nextChildren)
  }

  return node
}

export function MarkdownSnippet({ markdown, query }: { markdown: string; query: string }) {
  const terms = useMemo(() => extractHighlightTerms(query), [query])
  const re = useMemo(() => {
    if (terms.length === 0) return null
    const pattern = `(${terms.map(escapeRegExp).join('|')})`
    return new RegExp(pattern, 'gi')
  }, [terms])

  type MarkdownElementProps<T extends keyof ReactJSX.IntrinsicElements> = ReactJSX.IntrinsicElements[T] & {
    node?: unknown
    children?: ReactNode
  }

  const components: Components = useMemo(
    () => ({
      p: ({ children, node, ...props }: MarkdownElementProps<'p'>) => {
        void node
        return (
          <p {...props} className={props.className} style={{ marginBottom: 0 }}>
            {highlightReactNode(children, re)}
          </p>
        )
      },
      li: ({ children, node, ...props }: MarkdownElementProps<'li'>) => {
        void node
        return <li {...props}>{highlightReactNode(children, re)}</li>
      },
      h1: ({ children, node, ...props }: MarkdownElementProps<'h1'>) => {
        void node
        return <h1 {...props}>{highlightReactNode(children, re)}</h1>
      },
      h2: ({ children, node, ...props }: MarkdownElementProps<'h2'>) => {
        void node
        return <h2 {...props}>{highlightReactNode(children, re)}</h2>
      },
      h3: ({ children, node, ...props }: MarkdownElementProps<'h3'>) => {
        void node
        return <h3 {...props}>{highlightReactNode(children, re)}</h3>
      },
      h4: ({ children, node, ...props }: MarkdownElementProps<'h4'>) => {
        void node
        return <h4 {...props}>{highlightReactNode(children, re)}</h4>
      },
      h5: ({ children, node, ...props }: MarkdownElementProps<'h5'>) => {
        void node
        return <h5 {...props}>{highlightReactNode(children, re)}</h5>
      },
      h6: ({ children, node, ...props }: MarkdownElementProps<'h6'>) => {
        void node
        return <h6 {...props}>{highlightReactNode(children, re)}</h6>
      },
      blockquote: ({ children, node, ...props }: MarkdownElementProps<'blockquote'>) => {
        void node
        return <blockquote {...props}>{highlightReactNode(children, re)}</blockquote>
      },
    }),
    [re],
  )

  return (
    <div className="snippet-markdown">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {markdown}
      </ReactMarkdown>
    </div>
  )
}

