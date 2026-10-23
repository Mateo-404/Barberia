import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import viteCompression from 'vite-plugin-compression'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, import.meta.dirname, '')
  // connect-src sale del mismo VITE_API_BASE_URL que usa api-client.ts, así la CSP
  // nunca bloquea al backend configurado. En dev se suma el websocket de HMR.
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
      // Reemplaza el placeholder de index.html con los orígenes de connect-src configurados.
      {
        name: 'csp-connect-src',
        transformIndexHtml(html: string) {
          const out = html.replace('__CSP_CONNECT_SRC__', connectSrc)
          // Solo dev: el preámbulo de React Refresh es un <script> inline.
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
