import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // Forward API calls to the Go server so the browser only talks to one origin (no CORS needed).
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
