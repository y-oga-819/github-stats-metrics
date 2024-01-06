import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import {QueryClient, QueryClientProvider} from '@tanstack/react-query'
import { BrowserRouter, RouterProvider } from 'react-router-dom'
import { App } from './App.tsx'

const queryClient: QueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
