import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

export type SearchMode = 'docs' | 'reverse' | 'files'

export type SearchFilters = {
  ticket: string
  topics: string[]
  docType: string
  status: string
  file: string
  dir: string
  orderBy: string
  includeArchived: boolean
  includeScripts: boolean
  includeControlDocs: boolean
}

export type SearchState = {
  mode: SearchMode
  query: string
  filters: SearchFilters
}

const initialState: SearchState = {
  mode: 'docs',
  query: '',
  filters: {
    ticket: '',
    topics: [],
    docType: '',
    status: '',
    file: '',
    dir: '',
    orderBy: 'rank',
    includeArchived: true,
    includeScripts: true,
    includeControlDocs: true,
  },
}

const searchSlice = createSlice({
  name: 'search',
  initialState,
  reducers: {
    setMode(state, action: PayloadAction<SearchMode>) {
      state.mode = action.payload
    },
    setQuery(state, action: PayloadAction<string>) {
      state.query = action.payload
    },
    setFilter(
      state,
      action: PayloadAction<{
        key: keyof SearchFilters
        value: SearchFilters[keyof SearchFilters]
      }>,
    ) {
      state.filters[action.payload.key] = action.payload.value as never
    },
    clearFilters(state) {
      state.query = ''
      state.filters.ticket = ''
      state.filters.topics = []
      state.filters.docType = ''
      state.filters.status = ''
      state.filters.file = ''
      state.filters.dir = ''
      state.filters.orderBy = 'rank'
      state.filters.includeArchived = true
      state.filters.includeScripts = true
      state.filters.includeControlDocs = true
    },
  },
})

export const { setMode, setQuery, setFilter, clearFilters } = searchSlice.actions
export const searchReducer = searchSlice.reducer
