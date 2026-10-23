import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import viteCompression from 'vite-plugin-compression'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, import.meta.dirname, '')
  // connect-src comes from the same VITE_API_BASE_URL api-client.ts uses, so the
  // CSP never blocks the configured backend. Dev adds the HMR websocket.
  const apiOrigin = new URL(env.VITE_API_BASE_URL || 'http://localhost:8080').origin
  const connectSrc = [
    apiOrigin,
    mode === 'development' ? 'ws://localhost:5173' : '',
    env.VITE_CSP_CONNECT_SRC ?? '',
  ].filter(Boolean).join(' ')

  return {
    plugins: [
      react(),
      tailwindcss(),
      (viteCompression as any)({ algorithm: 'brotliCompress', ext: '.br', deleteOriginFile: false }),
      (viteCompression as any)({ algorithm: 'gzip', ext: '.gz', deleteOriginFile: false }),
      // Replaces the index.html placeholder with the configured connect-src origins.
      {
        name: 'csp-connect-src',
        transformIndexHtml(html: string) {
          const out = html.replace('__CSP_CONNECT_SRC__', connectSrc)
          // Dev only: the React Refresh preamble is an inline <script>.
          return mode === 'development'
            ? out.replace("default-src 'self';", "default-src 'self'; script-src 'self' 'unsafe-inline';")
            : out
        },
      },
    ],
    resolve: {
      alias: {
        '@': path.resolve(import.meta.dirname, 'src'),
      },
    },
  }
})
