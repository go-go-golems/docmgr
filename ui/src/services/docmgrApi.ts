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
  diagnostics: unknown[]
  nextCursor: string
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
  }),
})

export const {
  useGetWorkspaceStatusQuery,
  useRefreshIndexMutation,
  useLazySearchDocsQuery,
  useLazySearchFilesQuery,
} = docmgrApi
