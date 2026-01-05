import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export type WorkspaceStatus = {
  root: string
  configDir: string
  repoRoot: string
  configPath: string
  vocabularyPath: string
  indexedAt: string
  docsIndexed: number
  ftsAvailable: boolean
}

export type RefreshIndexResponse = {
  refreshed: boolean
  indexedAt: string
  docsIndexed: number
  ftsAvailable: boolean
}

export type WorkspaceSummaryStats = {
  ticketsTotal: number
  ticketsActive: number
  ticketsReview: number
  ticketsComplete: number
  ticketsDraft: number
}

export type WorkspaceTicketListItemStats = {
  docsTotal: number
  tasksTotal: number
  tasksDone: number
  relatedFilesTotal: number
}

export type WorkspaceTicketListItem = {
  ticket: string
  title: string
  status: string
  topics: string[]
  owners: string[]
  intent: string
  createdAt: string
  updatedAt: string
  ticketDir: string
  indexPath: string
  snippet: string
  stats: WorkspaceTicketListItemStats | null
}

export type WorkspaceRecentDocItem = {
  path: string
  ticket: string
  title: string
  docType: string
  status: string
  topics: string[]
  updatedAt: string
}

export type WorkspaceSummaryResponse = {
  root: string
  repoRoot: string
  indexedAt: string
  docsIndexed: number
  stats: WorkspaceSummaryStats
  recent: {
    tickets: WorkspaceTicketListItem[]
    docs: WorkspaceRecentDocItem[]
  }
}

export type WorkspaceTicketsQueryEcho = {
  q: string
  status: string
  ticket: string
  topics: string[]
  owners: string[]
  intent: string
  orderBy: string
  reverse: boolean
  includeArchived: boolean
  includeStats: boolean
  pageSize: number
  cursor: string
}

export type WorkspaceTicketsResponse = {
  query: WorkspaceTicketsQueryEcho
  total: number
  results: WorkspaceTicketListItem[]
  nextCursor: string
}

export type WorkspaceTicketsArgs = {
  q?: string
  status?: string
  ticket?: string
  topics?: string[]
  owners?: string[]
  intent?: string
  orderBy?: 'last_updated' | 'ticket' | 'title'
  reverse?: boolean
  includeArchived?: boolean
  includeStats?: boolean
  pageSize?: number
  cursor?: string
}

export type WorkspaceFacetsResponse = {
  statuses: string[]
  docTypes: string[]
  intents: string[]
  topics: string[]
  owners: string[]
}

export type WorkspaceRecentResponse = {
  tickets: WorkspaceTicketListItem[]
  docs: WorkspaceRecentDocItem[]
}

export type WorkspaceTopicListItem = {
  topic: string
  docsTotal: number
  ticketsTotal: number
  updatedAt: string
}

export type WorkspaceTopicsResponse = {
  total: number
  results: WorkspaceTopicListItem[]
}

export type WorkspaceTopicDetailResponse = {
  topic: string
  stats: WorkspaceSummaryStats
  tickets: WorkspaceTicketListItem[]
  docs: WorkspaceRecentDocItem[]
}

export type RelatedFile = {
  path: string
  note?: string
}

export type SearchDocResult = {
  ticket: string
  title: string
  docType: string
  status: string
  topics: string[]
  path: string
  lastUpdated?: string
  snippet: string
  relatedFiles: RelatedFile[]
  matchedFiles: string[]
  matchedNotes: string[]
}

export type SearchDocsQueryEcho = {
  query: string
  ticket: string
  topics: string[]
  docType: string
  status: string
  file: string
  dir: string
  externalSource: string
  since: string
  until: string
  createdSince: string
  updatedSince: string
  orderBy: string
  reverse: boolean
  pageSize: number
  cursor: string
}

export type SearchDocsResponse = {
  query: SearchDocsQueryEcho
  total: number
  results: SearchDocResult[]
  diagnostics: DiagnosticTaxonomy[] | null
  nextCursor: string
}

export type DiagnosticTaxonomy = {
  Tool?: string
  Stage?: string
  Symptom?: string
  Path?: string
  Severity?: string
  Context?: Record<string, unknown>
  Cause?: Record<string, unknown>
}

export type SearchDocsArgs = {
  query?: string
  ticket?: string
  topics?: string[]
  docType?: string
  status?: string
  file?: string
  dir?: string
  orderBy?: string
  reverse?: boolean
  includeArchived?: boolean
  includeScripts?: boolean
  includeControlDocs?: boolean
  includeDiagnostics?: boolean
  pageSize?: number
  cursor?: string
}

export type FileSuggestion = {
  file: string
  source: string
  reason: string
}

export type SearchFilesResponse = {
  total: number
  results: FileSuggestion[]
}

export type SearchFilesArgs = {
  query?: string
  ticket?: string
  topics?: string[]
  limit?: number
}

export type DocumentMeta = {
  title: string
  ticket: string
  status: string
  topics: string[]
  docType: string
  intent: string
  owners: string[]
  relatedFiles: RelatedFile[]
  externalSources: string[]
  summary: string
  lastUpdated: string
  whatFor: string
  whenToUse: string
}

export type FileStats = {
  sizeBytes: number
  modTime: string
}

export type DocGetResponse = {
  path: string
  doc?: DocumentMeta
  relatedFiles: RelatedFile[]
  body: string
  stats: FileStats
  diagnostic?: DiagnosticTaxonomy
}

export type FileGetResponse = {
  path: string
  root: 'repo' | 'docs'
  language: string
  contentType: string
  truncated: boolean
  content: string
  stats: FileStats
}

export type TicketStats = {
  docsTotal: number
  tasksTotal: number
  tasksDone: number
  relatedFilesTotal: number
}

export type TicketGetResponse = {
  ticket: string
  title: string
  status: string
  intent: string
  owners: string[]
  topics: string[]
  createdAt: string
  updatedAt: string
  ticketDir: string
  indexPath: string
  stats: TicketStats
}

export type TicketDocItem = {
  path: string
  title: string
  docType: string
  status: string
  topics: string[]
  summary: string
  lastUpdated?: string
  relatedFiles: RelatedFile[]
}

export type TicketDocsResponse = {
  ticket: string
  total: number
  results: TicketDocItem[]
  nextCursor: string
}

export type TicketTasksItem = {
  id: number
  checked: boolean
  text: string
}

export type TicketTasksSection = {
  title: string
  items: TicketTasksItem[]
}

export type TicketTasksResponse = {
  ticket: string
  exists: boolean
  tasksPath: string
  stats: { total: number; done: number }
  sections: TicketTasksSection[]
}

export type TicketGraphResponse = {
  ticket: string
  direction: 'TD' | 'LR'
  mermaid: string
  stats: { nodes: number; edges: number }
}

export const docmgrApi = createApi({
  reducerPath: 'docmgrApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api/v1' }),
  tagTypes: ['Workspace', 'Search', 'Ticket'],
  endpoints: (builder) => ({
    healthz: builder.query<{ ok: boolean }, void>({
      query: () => '/healthz',
    }),
    getWorkspaceStatus: builder.query<WorkspaceStatus, void>({
      query: () => '/workspace/status',
      providesTags: ['Workspace'],
    }),
    refreshIndex: builder.mutation<RefreshIndexResponse, void>({
      query: () => ({ url: '/index/refresh', method: 'POST' }),
      invalidatesTags: ['Workspace', 'Search'],
    }),
    getWorkspaceSummary: builder.query<WorkspaceSummaryResponse, void>({
      query: () => '/workspace/summary',
      providesTags: ['Workspace'],
    }),
    getWorkspaceTickets: builder.query<WorkspaceTicketsResponse, WorkspaceTicketsArgs>({
      query: (args) => ({
        url: '/workspace/tickets',
        params: {
          q: args.q ?? '',
          status: args.status ?? '',
          ticket: args.ticket ?? '',
          topics: (args.topics ?? []).join(','),
          owners: (args.owners ?? []).join(','),
          intent: args.intent ?? '',
          orderBy: args.orderBy ?? 'last_updated',
          reverse: args.reverse ?? false,
          includeArchived: args.includeArchived ?? true,
          includeStats: args.includeStats ?? false,
          pageSize: args.pageSize ?? 200,
          cursor: args.cursor ?? '',
        },
      }),
      serializeQueryArgs: ({ queryArgs }) => {
        const { cursor, ...rest } = queryArgs ?? {}
        void cursor
        return rest
      },
      merge: (currentCache, newData, { arg }) => {
        const cursor = arg?.cursor ?? ''

        if (!cursor) {
          currentCache.query = newData.query
          currentCache.total = newData.total
          currentCache.results = newData.results
          currentCache.nextCursor = newData.nextCursor
          return
        }

        currentCache.query = newData.query
        currentCache.total = newData.total
        currentCache.nextCursor = newData.nextCursor

        const seen = new Set(currentCache.results.map((r) => r.ticket))
        for (const r of newData.results) {
          if (seen.has(r.ticket)) continue
          currentCache.results.push(r)
          seen.add(r.ticket)
        }
      },
      forceRefetch: ({ currentArg, previousArg }) =>
        (currentArg?.cursor ?? '') !== (previousArg?.cursor ?? ''),
      providesTags: ['Workspace'],
    }),
    getWorkspaceFacets: builder.query<WorkspaceFacetsResponse, { includeArchived?: boolean } | undefined>({
      query: (args) => ({
        url: '/workspace/facets',
        params: { includeArchived: args?.includeArchived ?? true },
      }),
      providesTags: ['Workspace'],
    }),
    getWorkspaceRecent: builder.query<
      WorkspaceRecentResponse,
      { ticketsLimit?: number; docsLimit?: number; includeArchived?: boolean } | undefined
    >({
      query: (args) => ({
        url: '/workspace/recent',
        params: {
          ticketsLimit: args?.ticketsLimit ?? 20,
          docsLimit: args?.docsLimit ?? 20,
          includeArchived: args?.includeArchived ?? true,
        },
      }),
      providesTags: ['Workspace'],
    }),
    getWorkspaceTopics: builder.query<WorkspaceTopicsResponse, { includeArchived?: boolean } | undefined>({
      query: (args) => ({
        url: '/workspace/topics',
        params: { includeArchived: args?.includeArchived ?? true },
      }),
      providesTags: ['Workspace'],
    }),
    getWorkspaceTopic: builder.query<
      WorkspaceTopicDetailResponse,
      { topic: string; includeArchived?: boolean; docsLimit?: number }
    >({
      query: (args) => ({
        url: '/workspace/topics/get',
        params: {
          topic: args.topic,
          includeArchived: args.includeArchived ?? true,
          docsLimit: args.docsLimit ?? 20,
        },
      }),
      providesTags: ['Workspace'],
    }),
    searchDocs: builder.query<SearchDocsResponse, SearchDocsArgs>({
      query: (args) => ({
        url: '/search/docs',
        params: {
          query: args.query ?? '',
          ticket: args.ticket ?? '',
          topics: (args.topics ?? []).join(','),
          docType: args.docType ?? '',
          status: args.status ?? '',
          file: args.file ?? '',
          dir: args.dir ?? '',
          orderBy: args.orderBy ?? '',
          reverse: args.reverse ?? false,
          includeArchived: args.includeArchived ?? true,
          includeScripts: args.includeScripts ?? true,
          includeControlDocs: args.includeControlDocs ?? true,
          includeDiagnostics: args.includeDiagnostics ?? true,
          pageSize: args.pageSize ?? 200,
          cursor: args.cursor ?? '',
        },
      }),
      serializeQueryArgs: ({ queryArgs }) => {
        const { cursor, ...rest } = queryArgs ?? {}
        void cursor
        return rest
      },
      merge: (currentCache, newData, { arg }) => {
        const cursor = arg?.cursor ?? ''

        if (!cursor) {
          currentCache.query = newData.query
          currentCache.total = newData.total
          currentCache.results = newData.results
          currentCache.diagnostics = newData.diagnostics
          currentCache.nextCursor = newData.nextCursor
          return
        }

        currentCache.query = newData.query
        currentCache.total = newData.total
        currentCache.nextCursor = newData.nextCursor
        if (newData.diagnostics != null) currentCache.diagnostics = newData.diagnostics

        const seen = new Set(currentCache.results.map((r) => `${r.ticket}:${r.path}`))
        for (const r of newData.results) {
          const key = `${r.ticket}:${r.path}`
          if (seen.has(key)) continue
          currentCache.results.push(r)
          seen.add(key)
        }
      },
      forceRefetch: ({ currentArg, previousArg }) =>
        (currentArg?.cursor ?? '') !== (previousArg?.cursor ?? ''),
      providesTags: ['Search'],
    }),
    searchFiles: builder.query<SearchFilesResponse, SearchFilesArgs>({
      query: (args) => ({
        url: '/search/files',
        params: {
          query: args.query ?? '',
          ticket: args.ticket ?? '',
          topics: (args.topics ?? []).join(','),
          limit: args.limit ?? 200,
        },
      }),
    }),
    getDoc: builder.query<DocGetResponse, { path: string }>({
      query: (args) => ({
        url: '/docs/get',
        params: { path: args.path },
      }),
    }),
    getFile: builder.query<FileGetResponse, { path: string; root?: 'repo' | 'docs' }>({
      query: (args) => ({
        url: '/files/get',
        params: { path: args.path, root: args.root ?? 'repo' },
      }),
    }),

    getTicket: builder.query<TicketGetResponse, { ticket: string }>({
      query: (args) => ({ url: '/tickets/get', params: { ticket: args.ticket } }),
      providesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),

    getTicketDocs: builder.query<
      TicketDocsResponse,
      {
        ticket: string
        pageSize?: number
        cursor?: string
        orderBy?: 'path' | 'last_updated'
        includeArchived?: boolean
        includeScripts?: boolean
        includeControlDocs?: boolean
      }
    >({
      query: (args) => ({
        url: '/tickets/docs',
        params: {
          ticket: args.ticket,
          pageSize: args.pageSize ?? 200,
          cursor: args.cursor ?? '',
          orderBy: args.orderBy ?? 'path',
          includeArchived: args.includeArchived ?? true,
          includeScripts: args.includeScripts ?? true,
          includeControlDocs: args.includeControlDocs ?? true,
        },
      }),
      providesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),

    getTicketTasks: builder.query<TicketTasksResponse, { ticket: string }>({
      query: (args) => ({ url: '/tickets/tasks', params: { ticket: args.ticket } }),
      providesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),

    checkTicketTasks: builder.mutation<{ ok: boolean }, { ticket: string; ids: number[]; checked: boolean }>({
      query: (args) => ({
        url: '/tickets/tasks/check',
        method: 'POST',
        body: { ticket: args.ticket, ids: args.ids, checked: args.checked },
      }),
      invalidatesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),

    addTicketTask: builder.mutation<{ ok: boolean }, { ticket: string; section: string; text: string }>({
      query: (args) => ({
        url: '/tickets/tasks/add',
        method: 'POST',
        body: { ticket: args.ticket, section: args.section, text: args.text },
      }),
      invalidatesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),

    getTicketGraph: builder.query<
      TicketGraphResponse,
      {
        ticket: string
        direction?: 'TD' | 'LR'
        includeArchived?: boolean
        includeScripts?: boolean
        includeControlDocs?: boolean
      }
    >({
      query: (args) => ({
        url: '/tickets/graph',
        params: {
          ticket: args.ticket,
          direction: args.direction ?? 'TD',
          includeArchived: args.includeArchived ?? false,
          includeScripts: args.includeScripts ?? false,
          includeControlDocs: args.includeControlDocs ?? true,
        },
      }),
      providesTags: (_r, _e, args) => [{ type: 'Ticket', id: args.ticket }],
    }),
  }),
})

export const {
  useGetWorkspaceStatusQuery,
  useGetWorkspaceSummaryQuery,
  useGetWorkspaceTicketsQuery,
  useGetWorkspaceFacetsQuery,
  useGetWorkspaceRecentQuery,
  useGetWorkspaceTopicsQuery,
  useGetWorkspaceTopicQuery,
  useRefreshIndexMutation,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
  useGetDocQuery,
  useGetFileQuery,
  useGetTicketQuery,
  useGetTicketDocsQuery,
  useGetTicketTasksQuery,
  useCheckTicketTasksMutation,
  useAddTicketTaskMutation,
  useGetTicketGraphQuery,
} = docmgrApi
