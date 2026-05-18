#!/usr/bin/env node

const fs = require('node:fs/promises');
const path = require('node:path');

const siteDir = path.resolve(__dirname, '..');
const i18nDir = path.join(siteDir, 'i18n');
const docsPluginDir = 'docusaurus-plugin-content-docs';

async function exists(targetPath) {
  try {
    await fs.access(targetPath);
    return true;
  } catch {
    return false;
  }
}

function normalizeVersion(version) {
  return version.startsWith('version-') ? version.slice('version-'.length) : version;
}

async function getVersions() {
  const args = process.argv.slice(2).filter((arg) => arg.trim() !== '');
  if (args.length === 0) {
    throw new Error('Usage: node scripts/sync-versioned-i18n-docs.js <version>');
  }
  return args.map((arg) => normalizeVersion(arg.trim()));
}

async function copyCurrentDocsToVersion(locale, versionDirName) {
  const docsI18nDir = path.join(i18nDir, locale, docsPluginDir);
  const currentDir = path.join(docsI18nDir, 'current');
  const versionDir = path.join(docsI18nDir, versionDirName);

  if (!(await exists(currentDir))) {
    return false;
  }

  await fs.mkdir(path.dirname(versionDir), {recursive: true});
  await fs.rm(versionDir, {recursive: true, force: true});
  await fs.cp(currentDir, versionDir, {
    recursive: true,
  });

  console.log(
    `Synced ${locale} docs: ${path.relative(siteDir, currentDir)} -> ${path.relative(siteDir, versionDir)}`,
  );
  return true;
}

async function syncVersion(version) {
  const versionDirName = `version-${version}`;
  const sourceVersionDir = path.join(siteDir, 'versioned_docs', versionDirName);

  if (!(await exists(sourceVersionDir))) {
    throw new Error(
      `Missing ${path.relative(siteDir, sourceVersionDir)}. Run docs:version ${version} first.`,
    );
  }

  if (!(await exists(i18nDir))) {
    console.log('No i18n directory found; nothing to sync.');
    return;
  }

  const entries = await fs.readdir(i18nDir, {withFileTypes: true});
  let syncedCount = 0;

  for (const entry of entries) {
    if (!entry.isDirectory()) {
      continue;
    }

    const didSync = await copyCurrentDocsToVersion(entry.name, versionDirName);
    if (didSync) {
      syncedCount += 1;
    }
  }

  if (syncedCount === 0) {
    console.log(`No translated current docs found for ${version}.`);
  }
}

async function main() {
  const versions = await getVersions();
  if (versions.length === 0) {
    console.log('No versioned docs found; nothing to sync.');
    return;
  }

  for (const version of versions) {
    await syncVersion(version);
  }
}

main().catch((error) => {
  console.error(`Error: ${error.message}`);
  process.exit(1);
});
