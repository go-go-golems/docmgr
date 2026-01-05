export async function copyToClipboard(text: string): Promise<void> {
  if (!navigator.clipboard) throw new Error('clipboard not available')
  await navigator.clipboard.writeText(text)
}

