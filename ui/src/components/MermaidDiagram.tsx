import mermaid from 'mermaid'
import { useEffect, useId, useRef, useState } from 'react'

let configured = false
function ensureConfigured() {
  if (configured) return
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: 'strict',
    theme: 'default',
  })
  configured = true
}

export function MermaidDiagram({ code }: { code: string }) {
  const id = useId().replace(/[:]/g, '_')
  const ref = useRef<HTMLDivElement | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    ensureConfigured()

    const el = ref.current
    if (!el) return
    const input = (code ?? '').trim()
    if (!input) {
      el.innerHTML = ''
      return
    }

    let canceled = false
    mermaid
      .render(`mmd_${id}`, input)
      .then(({ svg, bindFunctions }) => {
        if (canceled) return
        setError(null)
        el.innerHTML = svg
        if (bindFunctions) bindFunctions(el)
      })
      .catch((e: unknown) => {
        if (canceled) return
        setError(e instanceof Error ? e.message : String(e))
        el.innerHTML = ''
      })

    return () => {
      canceled = true
    }
  }, [code, id])

  if (error) {
    return <div className="alert alert-danger mb-0">Mermaid render failed: {error}</div>
  }

  return <div ref={ref} />
}
