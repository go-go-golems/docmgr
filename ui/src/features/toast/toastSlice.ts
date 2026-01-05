import { createSlice, nanoid, type PayloadAction } from '@reduxjs/toolkit'

export type ToastKind = 'success' | 'error' | 'info' | 'warning'

export type Toast = {
  id: string
  kind: ToastKind
  message: string
  timeoutMs: number
  createdAt: number
}

type ToastState = {
  toasts: Toast[]
}

const initialState: ToastState = {
  toasts: [],
}

export const toastSlice = createSlice({
  name: 'toast',
  initialState,
  reducers: {
    pushToast: {
      reducer(state, action: PayloadAction<Toast>) {
        state.toasts.push(action.payload)
      },
      prepare(args: { kind: ToastKind; message: string; timeoutMs?: number }) {
        return {
          payload: {
            id: nanoid(),
            kind: args.kind,
            message: args.message,
            timeoutMs: args.timeoutMs ?? 1600,
            createdAt: Date.now(),
          } satisfies Toast,
        }
      },
    },
    removeToast(state, action: PayloadAction<{ id: string }>) {
      state.toasts = state.toasts.filter((t) => t.id !== action.payload.id)
    },
    clearToasts(state) {
      state.toasts = []
    },
  },
})

export const { pushToast, removeToast, clearToasts } = toastSlice.actions
export const toastReducer = toastSlice.reducer

