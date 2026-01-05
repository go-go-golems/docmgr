import { Provider } from 'react-redux'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { store } from './app/store'
import { ToastHost } from './components/ToastHost'
import { DocViewerPage } from './features/doc/DocViewerPage'
import { FileViewerPage } from './features/file/FileViewerPage'
import { SearchPage } from './features/search/SearchPage'
import { TicketPage } from './features/ticket/TicketPage'
import { WorkspaceLayout } from './features/workspace/WorkspaceLayout'
import { WorkspaceHomePage } from './features/workspace/WorkspaceHomePage'
import { WorkspaceTicketsPage } from './features/workspace/WorkspaceTicketsPage'
import { WorkspaceTopicsPage } from './features/workspace/WorkspaceTopicsPage'
import { WorkspaceTopicDetailPage } from './features/workspace/WorkspaceTopicDetailPage'
import { WorkspaceRecentPage } from './features/workspace/WorkspaceRecentPage'
import './styles/design-system.css'
import './styles/search.css'

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <ToastHost />
        <Routes>
          <Route path="/" element={<Navigate to="/workspace" replace />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/workspace" element={<WorkspaceLayout />}>
            <Route index element={<WorkspaceHomePage />} />
            <Route path="tickets" element={<WorkspaceTicketsPage />} />
            <Route path="topics" element={<WorkspaceTopicsPage />} />
            <Route path="topics/:topic" element={<WorkspaceTopicDetailPage />} />
            <Route path="recent" element={<WorkspaceRecentPage />} />
          </Route>
          <Route path="/doc" element={<DocViewerPage />} />
          <Route path="/file" element={<FileViewerPage />} />
          <Route path="/ticket/:ticket" element={<TicketPage />} />
        </Routes>
      </BrowserRouter>
    </Provider>
  )
}

export default App
