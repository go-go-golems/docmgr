import { useEffect, useState } from 'react'

export function useIsMobile(breakpointPx = 992): boolean {
  const [isMobile, setIsMobile] = useState(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return false
    return window.matchMedia(`(max-width: ${breakpointPx}px)`).matches
  })

  useEffect(() => {
    if (!window.matchMedia) return
    const m = window.matchMedia(`(max-width: ${breakpointPx}px)`)
    const onChange = () => setIsMobile(m.matches)
    onChange()
    m.addEventListener('change', onChange)
    return () => m.removeEventListener('change', onChange)
  }, [breakpointPx])

  return isMobile
}
