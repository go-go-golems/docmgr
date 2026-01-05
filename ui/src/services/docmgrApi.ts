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

export const docmgrApi = createApi({
  reducerPath: 'docmgrApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api/v1' }),
  tagTypes: ['Workspace', 'Search'],
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
  }),
})

export const {
  useGetWorkspaceStatusQuery,
  useRefreshIndexMutation,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
  useGetDocQuery,
  useGetFileQuery,
} = docmgrApi
