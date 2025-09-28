import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import UmamiApp from './umami-app'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <UmamiApp />
  </StrictMode>,
)
