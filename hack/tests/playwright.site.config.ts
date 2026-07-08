import path from 'node:path';
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e/site',
  testMatch: '**/TC*.ts',
  timeout: 30000,
  reporter: [['list']],
  outputDir: path.resolve(__dirname, '../../temp/e2e/site-test-results'),
  use: {
    baseURL: process.env.E2E_SITE_BASE_URL ?? process.env.E2E_BASE_URL ?? 'http://127.0.0.1:3000',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    video: 'off',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        channel: process.env.E2E_BROWSER_CHANNEL,
      },
    },
  ],
});
