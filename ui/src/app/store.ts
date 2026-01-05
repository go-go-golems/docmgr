import { configureStore } from '@reduxjs/toolkit'

import { docmgrApi } from '../services/docmgrApi'
import { searchReducer } from '../features/search/searchSlice'
import { toastReducer } from '../features/toast/toastSlice'

export const store = configureStore({
  reducer: {
    search: searchReducer,
    toast: toastReducer,
    [docmgrApi.reducerPath]: docmgrApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(docmgrApi.middleware),
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
