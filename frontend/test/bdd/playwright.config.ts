import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig, devices } from '@playwright/test'
import { defineBddConfig } from 'playwright-bdd'

const configDir = path.dirname(fileURLToPath(import.meta.url))

const testDir = defineBddConfig({
  features: path.resolve(configDir, '../../../features/**/*.feature'),
  featuresRoot: path.resolve(configDir, '../../../features'),
  steps: './steps/**/*.ts',
  outputDir: '.features-gen',
  tags: '@browser',
})

export default defineConfig({
  testDir,
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: 'http://127.0.0.1:4173',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'npm run build && npm run preview -- --host 127.0.0.1',
    cwd: '../..',
    port: 4173,
    reuseExistingServer: !process.env.CI,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
