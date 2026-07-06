import { useMemo } from 'react'
import type { AnchorHTMLAttributes, ImgHTMLAttributes } from 'react'
import ReactMarkdown from 'react-markdown'
import type { Components } from 'react-markdown'
import { Link } from 'react-router-dom'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'

import { MermaidDiagram } from './MermaidDiagram'

// Minimal structural view of the hast nodes react-markdown hands to
// component overrides (kept local to avoid depending on hast types).
type HastNode = {
  type?: string
  value?: string
  tagName?: string
  properties?: { className?: unknown }
  children?: HastNode[]
}

function hastText(node: HastNode | undefined): string {
  if (!node) return ''
  if (typeof node.value === 'string') return node.value
  return (node.children ?? []).map(hastText).join('')
}

function hastClassNames(node: HastNode | undefined): string[] {
  const cn = node?.properties?.className
  if (Array.isArray(cn)) return cn.map(String)
  if (typeof cn === 'string') return cn.split(/\s+/)
  return []
}

/** Returns the mermaid source when the <pre> wraps a ```mermaid fence. */
function mermaidSource(preNode: HastNode | undefined): string | null {
  const code = (preNode?.children ?? []).find((c) => c.tagName === 'code')
  if (!code) return null
  if (!hastClassNames(code).includes('language-mermaid')) return null
  return hastText(code)
}

/** Resolve `rel` against `baseDir` (both slash-separated); null when the
 * result escapes the root ("../" beyond the top). */
function resolveRelative(baseDir: string, rel: string): string | null {
  const parts = baseDir ? baseDir.split('/').filter(Boolean) : []
  for (const seg of rel.split('/')) {
    if (!seg || seg === '.') continue
    if (seg === '..') {
      if (parts.length === 0) return null
      parts.pop()
    } else {
      parts.push(seg)
    }
  }
  return parts.join('/')
}

function isExternalUrl(href: string): boolean {
  return /^[a-z][a-z0-9+.-]*:/i.test(href)
}

function dirOf(path: string): string {
  const idx = path.lastIndexOf('/')
  return idx >= 0 ? path.slice(0, idx) : ''
}

type MarkdownAnchorProps = AnchorHTMLAttributes<HTMLAnchorElement> & { node?: unknown }
type MarkdownImgProps = ImgHTMLAttributes<HTMLImageElement> & { node?: unknown }

function MarkdownLink({ docDir, ...props }: MarkdownAnchorProps & { docDir: string }) {
  const { href, children, node, ...rest } = props
  void node
  const target = (href ?? '').trim()

  if (target === '' || target.startsWith('#')) {
    return (
      <a href={href} {...rest}>
        {children}
      </a>
    )
  }

  if (isExternalUrl(target)) {
    const external = /^https?:/i.test(target)
    return (
      <a
        href={href}
        {...rest}
        {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
      >
        {children}
      </a>
    )
  }

  // Absolute paths are treated as repo file paths.
  if (target.startsWith('/')) {
    const repoPath = target.replace(/^\/+/, '')
    return (
      <Link to={`/file?root=repo&path=${encodeURIComponent(repoPath)}`} {...rest}>
        {children}
      </Link>
    )
  }

  // Relative paths resolve against the current doc's directory (docs root).
  const [pathPart, hash] = target.split('#', 2)
  const resolved = resolveRelative(docDir, pathPart)
  if (resolved === null || resolved === '') {
    return (
      <a href={href} {...rest}>
        {children}
      </a>
    )
  }
  if (/\.md$/i.test(resolved)) {
    const suffix = hash ? `#${hash}` : ''
    return (
      <Link to={`/doc?path=${encodeURIComponent(resolved)}${suffix}`} {...rest}>
        {children}
      </Link>
    )
  }
  return (
    <Link to={`/file?root=docs&path=${encodeURIComponent(resolved)}`} {...rest}>
      {children}
    </Link>
  )
}

function MarkdownImage({ docDir, ...props }: MarkdownImgProps & { docDir: string }) {
  const { src, alt, node, ...rest } = props
  void node
  const raw = typeof src === 'string' ? src.trim() : ''

  let resolvedSrc: string | undefined = typeof src === 'string' ? src : undefined
  if (raw !== '' && !isExternalUrl(raw)) {
    if (raw.startsWith('/')) {
      resolvedSrc = `/api/v1/files/raw?root=repo&path=${encodeURIComponent(raw.replace(/^\/+/, ''))}`
    } else {
      const resolved = resolveRelative(docDir, raw)
      if (resolved !== null && resolved !== '') {
        resolvedSrc = `/api/v1/files/raw?root=docs&path=${encodeURIComponent(resolved)}`
      }
    }
  }

  return <img src={resolvedSrc} alt={alt ?? ''} style={{ maxWidth: '100%' }} {...rest} />
}

export function MarkdownBlock({
  markdown,
  enableHighlight,
  docPath,
}: {
  markdown: string
  enableHighlight?: boolean
  /** Docs-root-relative path of the doc being rendered. Enables relative
   * link/image resolution; when omitted, links/images pass through. */
  docPath?: string
}) {
  const docDir = docPath ? dirOf(docPath) : ''

  const components: Components = useMemo(
    () => ({
      pre: (props) => {
        const { node, children, ...rest } = props
        const source = mermaidSource(node as HastNode | undefined)
        if (source !== null) {
          return <MermaidDiagram code={source} />
        }
        return <pre {...rest}>{children}</pre>
      },
      a: (props) => <MarkdownLink docDir={docDir} {...(props as MarkdownAnchorProps)} />,
      img: (props) => <MarkdownImage docDir={docDir} {...(props as MarkdownImgProps)} />,
    }),
    [docDir],
  )

  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={enableHighlight ?? true ? [rehypeHighlight] : undefined}
      components={components}
    >
      {markdown}
    </ReactMarkdown>
  )
}
