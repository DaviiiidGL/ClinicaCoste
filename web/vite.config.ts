import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite' // <-- Importar

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(), // <-- Añadir el plugin
  ],
  server: {
    proxy: {
      "/hospital.v1.HospitalService": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
})