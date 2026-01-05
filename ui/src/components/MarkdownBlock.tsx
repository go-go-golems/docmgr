import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'

export function MarkdownBlock({
  markdown,
  enableHighlight,
}: {
  markdown: string
  enableHighlight?: boolean
}) {
  return (
    <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={enableHighlight ?? true ? [rehypeHighlight] : undefined}>
      {markdown}
    </ReactMarkdown>
  )
}
