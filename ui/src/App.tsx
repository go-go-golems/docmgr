import { Provider } from 'react-redux'
import { BrowserRouter, Route, Routes } from 'react-router-dom'

import { store } from './app/store'
import { ToastHost } from './components/ToastHost'
import { DocViewerPage } from './features/doc/DocViewerPage'
import { FileViewerPage } from './features/file/FileViewerPage'
import { SearchPage } from './features/search/SearchPage'
import { TicketPage } from './features/ticket/TicketPage'
import './styles/design-system.css'
import './styles/search.css'

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <ToastHost />
        <Routes>
          <Route path="/" element={<SearchPage />} />
          <Route path="/doc" element={<DocViewerPage />} />
          <Route path="/file" element={<FileViewerPage />} />
          <Route path="/ticket/:ticket" element={<TicketPage />} />
        </Routes>
      </BrowserRouter>
    </Provider>
  )
}

export default App
