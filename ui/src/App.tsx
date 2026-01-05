import { Provider } from 'react-redux'
import { BrowserRouter, Route, Routes } from 'react-router-dom'

import { store } from './app/store'
import { DocViewerPage } from './features/doc/DocViewerPage'
import { FileViewerPage } from './features/file/FileViewerPage'
import { SearchPage } from './features/search/SearchPage'
import './App.css'

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<SearchPage />} />
          <Route path="/doc" element={<DocViewerPage />} />
          <Route path="/file" element={<FileViewerPage />} />
        </Routes>
      </BrowserRouter>
    </Provider>
  )
}

export default App
