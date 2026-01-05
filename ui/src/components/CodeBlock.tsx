export function CodeBlock({
  html,
  language,
  className,
}: {
  html: string
  language?: string
  className?: string
}) {
  return (
    <pre className={['docmgr-code', className].filter(Boolean).join(' ')}>
      <code
        className={`hljs ${language ? `language-${language}` : ''}`.trim()}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </pre>
  )
}
