import { defineConfig } from 'vite';
import { resolve } from 'path';

// See FRONTEND_SCOPE.md Section 12.1 - Vite Configuration
export default defineConfig({
    // Phase 12: Load .env from project root (parent dir) so Clerk keys are accessible
    envDir: resolve(__dirname, '..'),
    resolve: {
        alias: {
            '@': resolve(__dirname, './src'),
        },
    },
    server: {
        port: 3000,
        host: true,
        proxy: {
            '/api': {
                target: 'http://localhost:8081',
                changeOrigin: true,
            },
        },
    },
    build: {
        outDir: 'dist',
        emptyOutDir: true,
        target: 'esnext',
        modulePreload: { polyfill: false },
    },
});
