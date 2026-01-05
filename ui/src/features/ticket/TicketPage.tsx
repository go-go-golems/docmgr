import { useMemo, useState } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'

import {
  useAddTicketTaskMutation,
  useCheckTicketTasksMutation,
  useGetDocQuery,
  useGetTicketDocsQuery,
  useGetTicketGraphQuery,
  useGetTicketQuery,
  useGetTicketTasksQuery,
  type TicketDocItem,
} from '../../services/docmgrApi'

import { ApiErrorAlert } from '../../components/ApiErrorAlert'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { copyToClipboard } from '../../lib/clipboard'

import { TicketHeader } from './components/TicketHeader'
import { TicketTabs, type TicketTabKey } from './components/TicketTabs'
import { TicketChangelogTab } from './tabs/TicketChangelogTab'
import { TicketDocumentsTab } from './tabs/TicketDocumentsTab'
import { TicketGraphTab } from './tabs/TicketGraphTab'
import { TicketOverviewTab } from './tabs/TicketOverviewTab'
import { TicketTasksTab } from './tabs/TicketTasksTab'

function normalizeTab(raw: string | null): TicketTabKey {
  const t = (raw ?? '').trim().toLowerCase()
  if (t === 'documents' || t === 'tasks' || t === 'graph' || t === 'changelog') return t
  return 'overview'
}

function formatDate(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function asArray<T>(v: T[] | null | undefined): T[] {
  return Array.isArray(v) ? v : []
}

function groupByDocType(items: TicketDocItem[]): Record<string, TicketDocItem[]> {
  const out: Record<string, TicketDocItem[]> = {}
  for (const it of items) {
    const k = it.docType?.trim() || 'unknown'
    out[k] ||= []
    out[k].push(it)
  }
  return out
}

export function TicketPage() {
  const params = useParams()
  const ticket = (params.ticket ?? '').trim()
  const [searchParams, setSearchParams] = useSearchParams()
  const [toast, setToast] = useState<{ kind: 'success' | 'error'; message: string } | null>(null)

  const tab = normalizeTab(searchParams.get('tab'))
  const selectedDoc = (searchParams.get('doc') ?? '').trim()

  const { data: t, error: ticketError, isLoading: ticketLoading } = useGetTicketQuery(
    { ticket },
    { skip: ticket === '' },
  )

  const { data: docsData, error: docsError, isLoading: docsLoading } = useGetTicketDocsQuery(
    { ticket, pageSize: 500, orderBy: 'path' },
    { skip: ticket === '' || (tab !== 'documents' && tab !== 'overview') },
  )

  const { data: tasksData, error: tasksError, isLoading: tasksLoading } = useGetTicketTasksQuery(
    { ticket },
    { skip: ticket === '' || (tab !== 'tasks' && tab !== 'overview') },
  )

  const { data: graphData, error: graphError, isLoading: graphLoading } = useGetTicketGraphQuery(
    { ticket, direction: 'TD' },
    { skip: ticket === '' || tab !== 'graph' },
  )

  const indexPath = (t?.indexPath ?? '').trim()
  const { data: indexDocData, error: indexDocError } = useGetDocQuery(
    { path: indexPath },
    { skip: indexPath === '' || tab !== 'overview' },
  )

  const [checkTask, checkTaskState] = useCheckTicketTasksMutation()
  const [addTask, addTaskState] = useAddTicketTaskMutation()

  async function onCopyPath(text: string) {
    try {
      await copyToClipboard(text)
      setToast({ kind: 'success', message: 'Copied' })
      setTimeout(() => setToast(null), 1200)
    } catch (e) {
      setToast({ kind: 'error', message: `Copy failed: ${String(e)}` })
      setTimeout(() => setToast(null), 2500)
    }
  }

  const selectedDocItem = useMemo(() => {
    const list = docsData?.results ?? []
    if (!selectedDoc) return null
    return list.find((d) => d.path === selectedDoc) ?? null
  }, [docsData, selectedDoc])

  const docsByType = useMemo(() => groupByDocType(docsData?.results ?? []), [docsData])
  const docTypeKeys = useMemo(() => Object.keys(docsByType).sort(), [docsByType])

  const keyDocs = useMemo(() => {
    const list = docsData?.results ?? []
    return list.filter((d) => d.docType !== 'index').slice(0, 6)
  }, [docsData])

  const openTasks = useMemo(() => {
    const secs = tasksData?.sections ?? []
    const out: { id: number; text: string; checked: boolean }[] = []
    for (const sec of secs) {
      for (const it of asArray(sec.items)) {
        if (!it.checked) out.push({ id: it.id, text: it.text, checked: it.checked })
      }
    }
    return out.slice(0, 10)
  }, [tasksData])

  function setTab(next: TicketTabKey) {
    const sp = new URLSearchParams(searchParams)
    sp.set('tab', next)
    if (next !== 'documents') sp.delete('doc')
    setSearchParams(sp, { replace: true })
  }

  function selectDoc(path: string) {
    const sp = new URLSearchParams(searchParams)
    sp.set('tab', 'documents')
    sp.set('doc', path)
    setSearchParams(sp, { replace: true })
  }

  function clearSelectedDoc() {
    const sp = new URLSearchParams(searchParams)
    sp.delete('doc')
    setSearchParams(sp, { replace: true })
  }

  return (
    <div className="container py-4">
      <TicketHeader ticket={ticket} title={t?.title} ticketDir={t?.ticketDir} />

      {toast ? (
        <div className={`alert ${toast.kind === 'success' ? 'alert-success' : 'alert-danger'} py-2`}>
          {toast.message}
        </div>
      ) : null}

      {ticket === '' ? <div className="alert alert-info">Missing ticket id.</div> : null}
      {ticketError ? <ApiErrorAlert title="Failed to load ticket" error={ticketError} /> : null}
      {ticketLoading ? <LoadingSpinner /> : null}

      <TicketTabs tab={tab} onTabChange={setTab} />

      {tab === 'overview' ? (
        <TicketOverviewTab
          ticket={ticket}
          ticketData={t}
          docsLoading={docsLoading}
          docsError={docsError}
          keyDocs={keyDocs}
          tasksLoading={tasksLoading}
          tasksError={tasksError}
          openTasks={openTasks}
          indexDocError={indexDocError}
          indexBody={indexDocData?.body}
          onSetTab={setTab}
          onCheckTask={(args) => checkTask(args).unwrap()}
          checkTaskLoading={checkTaskState.isLoading}
          formatDate={formatDate}
        />
      ) : null}

      {tab === 'documents' ? (
        <TicketDocumentsTab
          ticket={ticket}
          docsError={docsError}
          docsLoading={docsLoading}
          docTypeKeys={docTypeKeys}
          docsByType={docsByType}
          selectedDoc={selectedDoc}
          selectedDocItem={selectedDocItem}
          onSelectDoc={selectDoc}
          onClearSelectedDoc={clearSelectedDoc}
          onCopyPath={(p) => void onCopyPath(p)}
          formatDate={formatDate}
        />
      ) : null}

      {tab === 'tasks' ? (
        <TicketTasksTab
          ticket={ticket}
          tasksData={tasksData}
          tasksError={tasksError}
          tasksLoading={tasksLoading}
          checkTask={(args) => checkTask(args).unwrap()}
          checkTaskLoading={checkTaskState.isLoading}
          addTask={(args) => addTask(args).unwrap()}
          addTaskLoading={addTaskState.isLoading}
          addTaskError={addTaskState.error}
        />
      ) : null}

      {tab === 'graph' ? (
        <TicketGraphTab graphData={graphData} graphError={graphError} graphLoading={graphLoading} />
      ) : null}

      {tab === 'changelog' ? <TicketChangelogTab ticketDir={t?.ticketDir} /> : null}

      {/* TODO: Tab errors are rendered inline per tab; toast is page-scoped for now. */}
    </div>
  )
}
