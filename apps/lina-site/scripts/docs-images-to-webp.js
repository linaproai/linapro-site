#!/usr/bin/env node
'use strict';

const fs = require('node:fs/promises');
const path = require('node:path');

// 将 docs、blog 以及 i18n 内容中“实际被引用”的本地图片转换为 WebP，并同步改写引用路径。
// 转换成功且文档写入完成后，仅在全站已无原路径引用时删除原始图片。
let sharp;
try {
  sharp = require('sharp');
} catch (error) {
  console.error("Error: missing dependency 'sharp'. Run 'yarn install' first.");
  process.exit(1);
}

const ROOT_DIR = path.resolve(__dirname, '..');
const STATIC_DIR = path.join(ROOT_DIR, 'static');
const DOC_EXTS = new Set(['.md', '.mdx']);
const CONVERTIBLE_EXTS = new Set(['.png', '.jpg', '.jpeg', '.gif']);
const IMAGE_EXT_PATTERN = /\.(?:png|jpe?g|gif|webp)(?:[?#][^\s'"`<>)]*)?/i;
const FRONTMATTER_IMAGE_KEYS = ['image', 'cover', 'thumbnail'];
const SKIP_DIR_NAMES = new Set([
  '.docusaurus',
  '.git',
  '.turbo',
  '.yarn',
  'build',
  'dist',
  'node_modules',
]);
const SCANNABLE_EXTS = new Set([
  '.cjs',
  '.css',
  '.html',
  '.js',
  '.json',
  '.jsx',
  '.md',
  '.mdx',
  '.mjs',
  '.scss',
  '.ts',
  '.tsx',
  '.yaml',
  '.yml',
]);

// 默认使用无损 WebP，避免转换时降低图片质量；如需压缩体积，可显式设置 WEBP_LOSSLESS=0。
const args = process.argv.slice(2);
const dryRun = args.includes('--dry-run') || process.env.DRY_RUN === '1';
const includeStatic =
  args.includes('--include-static') || parseBoolean(process.env.WEBP_INCLUDE_STATIC || '0');
const lossless = parseBoolean(process.env.WEBP_LOSSLESS || '1');
const qualityArg = args.find((arg) => arg.startsWith('--quality='));
const quality = parseQuality(
  qualityArg ? qualityArg.slice('--quality='.length) : process.env.WEBP_QUALITY || '100',
);

function parseQuality(value) {
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed) || parsed < 1 || parsed > 100) {
    console.error(`Error: WEBP_QUALITY must be an integer from 1 to 100, got '${value}'.`);
    process.exit(1);
  }
  return parsed;
}

function parseBoolean(value) {
  return !['0', 'false', 'no', 'off'].includes(String(value).toLowerCase());
}

async function findContentDirs() {
  const dirs = [
    path.join(ROOT_DIR, 'docs'),
    path.join(ROOT_DIR, 'blog'),
    path.join(ROOT_DIR, 'versioned_docs'),
  ];
  const i18nDir = path.join(ROOT_DIR, 'i18n');

  if (await exists(i18nDir)) {
    const locales = await fs.readdir(i18nDir, { withFileTypes: true });
    for (const locale of locales) {
      if (!locale.isDirectory()) {
        continue;
      }

      const localeDir = path.join(i18nDir, locale.name);
      dirs.push(
        path.join(localeDir, 'docusaurus-plugin-content-docs'),
        path.join(localeDir, 'docusaurus-plugin-content-blog'),
      );
    }
  }

  const existingDirs = [];
  for (const dir of unique(dirs.map((item) => path.resolve(item)))) {
    if (await isDirectory(dir)) {
      existingDirs.push(dir);
    }
  }

  return existingDirs.sort();
}

async function main() {
  const contentDirs = await findContentDirs();
  const contentFiles = (
    await Promise.all(contentDirs.map((dir) => walkIfExists(dir)))
  ).flat().filter((file) => DOC_EXTS.has(path.extname(file))).sort();
  const contents = new Map();
  const refs = [];
  const skippedStaticRefs = [];
  const skippedSiteRefs = [];

  // 扫描 Markdown/MDX 中的图片引用，跳过代码块和行内代码，避免误改示例代码。
  for (const file of contentFiles) {
    const content = await fs.readFile(file, 'utf8');
    const protectedRanges = collectProtectedRanges(content);
    const fileRefs = [
      ...collectMarkdownImageRefs(content, protectedRanges),
      ...collectHtmlImageRefs(content, protectedRanges),
      ...collectRequireRefs(content, protectedRanges),
      ...collectImportRefs(content, protectedRanges),
      ...collectFrontmatterRefs(content),
    ];

    contents.set(file, content);

    for (const ref of fileRefs) {
      const resolved = resolveImageRef(ref.value, file, contentDirs);
      if (!resolved) {
        continue;
      }

      const resolvedRef = {
        ...ref,
        docFile: file,
        sourcePath: resolved.sourcePath,
        ext: resolved.ext,
        scope: resolved.scope,
      };

      if (resolved.scope === 'static' && !includeStatic) {
        skippedStaticRefs.push(resolvedRef);
        continue;
      }

      if (resolved.scope !== 'content' && resolved.scope !== 'static') {
        skippedSiteRefs.push(resolvedRef);
        continue;
      }

      refs.push(resolvedRef);
    }
  }

  // 只处理能解析到仓库内真实文件的本地图片引用，远程 URL 和不存在的文件不会被转换。
  const missingRefs = [];
  const existingRefs = [];
  for (const ref of refs) {
    if (await exists(ref.sourcePath)) {
      existingRefs.push(ref);
    } else {
      missingRefs.push(ref);
    }
  }

  const sourcePaths = unique(
    existingRefs
      .filter((ref) => CONVERTIBLE_EXTS.has(ref.ext))
      .map((ref) => ref.sourcePath),
  );
  const targetPaths = await allocateTargets(sourcePaths);
  const converted = [];
  const failed = [];

  // 每个源图片只转换一次；多个文档引用同一图片时，后续统一指向同一个 WebP 文件。
  for (const sourcePath of sourcePaths) {
    const targetPath = targetPaths.get(sourcePath);
    try {
      if (!dryRun) {
        await convertToWebp(sourcePath, targetPath);
      }
      converted.push(sourcePath);
    } catch (error) {
      failed.push({ sourcePath, error });
    }
  }

  const convertedSet = new Set(converted);
  const replacementsByFile = new Map();

  // 根据图片引用原来的写法生成新路径，尽量保留相对路径、/static 绝对路径和 @site 风格。
  for (const ref of existingRefs) {
    if (!convertedSet.has(ref.sourcePath)) {
      continue;
    }

    const targetPath = targetPaths.get(ref.sourcePath);
    const nextValue = buildRefValue(ref.value, ref.docFile, targetPath);
    if (nextValue === ref.value) {
      continue;
    }

    if (!replacementsByFile.has(ref.docFile)) {
      replacementsByFile.set(ref.docFile, []);
    }
    replacementsByFile.get(ref.docFile).push({
      start: ref.start,
      end: ref.end,
      value: nextValue,
    });
  }

  let updatedDocCount = 0;
  let updatedRefCount = 0;
  const plannedContents = new Map(contents);
  // 先写回所有文档引用，再删除原图；这样可以避免文档仍指向已删除文件。
  for (const [file, replacements] of replacementsByFile) {
    const original = contents.get(file);
    const updated = applyReplacements(original, replacements);
    if (updated === original) {
      continue;
    }

    plannedContents.set(file, updated);
    updatedDocCount += 1;
    updatedRefCount += replacements.length;
    if (!dryRun) {
      await fs.writeFile(file, updated);
    }
  }

  const deleted = [];
  const deleteFailed = [];
  // 只删除本次成功转换的源文件；dry-run 只统计将要删除的文件，不做实际删除。
  const sourcePathsToDelete = converted.filter((sourcePath) => {
    const targetPath = targetPaths.get(sourcePath);
    return targetPath && sourcePath !== targetPath;
  });
  const remainingRefsBySource = await findRemainingRefs(
    sourcePathsToDelete,
    contentDirs,
    plannedContents,
  );
  const blockedDeletes = [];

  for (const sourcePath of sourcePathsToDelete) {
    const remainingRefs = remainingRefsBySource.get(sourcePath) || [];
    if (remainingRefs.length > 0) {
      blockedDeletes.push({ sourcePath, refs: remainingRefs });
      continue;
    }

    try {
      if (!dryRun) {
        await fs.rm(sourcePath);
      }
      deleted.push(sourcePath);
    } catch (error) {
      deleteFailed.push({ sourcePath, error });
    }
  }

  printSummary({
    contentDirs: contentDirs.length,
    contentFiles: contentFiles.length,
    localRefs: refs.length + skippedStaticRefs.length + skippedSiteRefs.length,
    skippedStaticRefs,
    skippedSiteRefs,
    existingRefs: existingRefs.length,
    missingRefs,
    alreadyWebpRefs: existingRefs.filter((ref) => ref.ext === '.webp').length,
    sourceCount: sourcePaths.length,
    convertedCount: converted.length,
    failed,
    deletedCount: deleted.length,
    blockedDeletes,
    deleteFailed,
    updatedDocCount,
    updatedRefCount,
  });

  if (failed.length > 0 || deleteFailed.length > 0) {
    process.exitCode = 1;
  }
}

async function walk(dir) {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(fullPath)));
    } else if (entry.isFile()) {
      files.push(fullPath);
    }
  }

  return files;
}

async function walkIfExists(dir) {
  if (!(await exists(dir))) {
    return [];
  }
  return walk(dir);
}

function collectProtectedRanges(content) {
  const fenceRanges = collectFenceRanges(content);
  const inlineRanges = collectInlineCodeRanges(content, fenceRanges);
  return [...fenceRanges, ...inlineRanges].sort((a, b) => a.start - b.start);
}

function collectFenceRanges(content) {
  const ranges = [];
  let offset = 0;
  let activeFence = null;

  while (offset < content.length) {
    const nextNewline = content.indexOf('\n', offset);
    const lineEnd = nextNewline === -1 ? content.length : nextNewline + 1;
    const line = content.slice(offset, lineEnd);
    const marker = getFenceMarker(line);

    if (marker) {
      if (!activeFence) {
        activeFence = { ...marker, start: offset };
      } else if (marker.char === activeFence.char && marker.length >= activeFence.length) {
        ranges.push({ start: activeFence.start, end: lineEnd });
        activeFence = null;
      }
    }

    offset = lineEnd;
  }

  if (activeFence) {
    ranges.push({ start: activeFence.start, end: content.length });
  }

  return ranges;
}

function getFenceMarker(line) {
  const match = line.match(/^ {0,3}(`{3,}|~{3,})/);
  if (!match) {
    return null;
  }
  return { char: match[1][0], length: match[1].length };
}

function collectInlineCodeRanges(content, fenceRanges) {
  const ranges = [];
  let index = 0;

  while (index < content.length) {
    const fence = rangeContaining(fenceRanges, index);
    if (fence) {
      index = fence.end;
      continue;
    }

    if (content[index] !== '`') {
      index += 1;
      continue;
    }

    const start = index;
    while (index < content.length && content[index] === '`') {
      index += 1;
    }
    const marker = content.slice(start, index);
    const end = content.indexOf(marker, index);
    if (end === -1) {
      continue;
    }

    ranges.push({ start, end: end + marker.length });
    index = end + marker.length;
  }

  return ranges;
}

function collectMarkdownImageRefs(content, protectedRanges) {
  const refs = [];
  let searchFrom = 0;

  while (searchFrom < content.length) {
    const imageStart = content.indexOf('![', searchFrom);
    if (imageStart === -1) {
      break;
    }

    if (isProtected(protectedRanges, imageStart)) {
      searchFrom = imageStart + 2;
      continue;
    }

    const altEnd = content.indexOf(']', imageStart + 2);
    if (altEnd === -1 || content[altEnd + 1] !== '(') {
      searchFrom = imageStart + 2;
      continue;
    }

    const bodyStart = altEnd + 2;
    const bodyEnd = findClosingParen(content, bodyStart);
    if (bodyEnd === -1) {
      searchFrom = imageStart + 2;
      continue;
    }

    const body = content.slice(bodyStart, bodyEnd);
    const destination = parseMarkdownDestination(body);
    if (destination && IMAGE_EXT_PATTERN.test(destination.value)) {
      refs.push({
        start: bodyStart + destination.start,
        end: bodyStart + destination.end,
        value: destination.value,
      });
    }

    searchFrom = bodyEnd + 1;
  }

  return refs;
}

function findClosingParen(content, start) {
  let depth = 1;
  let inAngle = false;
  let escaped = false;

  for (let index = start; index < content.length; index += 1) {
    const char = content[index];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (char === '\\') {
      escaped = true;
      continue;
    }
    if (char === '<') {
      inAngle = true;
      continue;
    }
    if (char === '>') {
      inAngle = false;
      continue;
    }
    if (inAngle) {
      continue;
    }
    if (char === '(') {
      depth += 1;
      continue;
    }
    if (char === ')') {
      depth -= 1;
      if (depth === 0) {
        return index;
      }
    }
  }

  return -1;
}

function parseMarkdownDestination(body) {
  const leadingWhitespace = body.match(/^\s*/)[0].length;
  if (leadingWhitespace >= body.length) {
    return null;
  }

  if (body[leadingWhitespace] === '<') {
    const close = body.indexOf('>', leadingWhitespace + 1);
    if (close === -1) {
      return null;
    }
    return {
      start: leadingWhitespace + 1,
      end: close,
      value: body.slice(leadingWhitespace + 1, close),
    };
  }

  const rest = body.slice(leadingWhitespace);
  const match = rest.match(/^(.+?\.(?:png|jpe?g|gif|webp)(?:[?#][^\s'"<>)]*)?)(?=\s+(?:"|'|\()|$)/i);
  if (match) {
    return {
      start: leadingWhitespace,
      end: leadingWhitespace + match[1].length,
      value: match[1],
    };
  }

  const token = rest.match(/^\S+/);
  if (!token) {
    return null;
  }

  return {
    start: leadingWhitespace,
    end: leadingWhitespace + token[0].length,
    value: token[0],
  };
}

function collectHtmlImageRefs(content, protectedRanges) {
  const refs = [];
  const tagRegex = /<img\b[\s\S]*?>/gi;
  let tagMatch;

  while ((tagMatch = tagRegex.exec(content)) !== null) {
    if (isProtected(protectedRanges, tagMatch.index)) {
      continue;
    }

    const tag = tagMatch[0];
    const attrMatch = /\bsrc\s*=\s*(?:"([^"]+)"|'([^']+)'|{\s*["']([^"']+)["']\s*})/i.exec(tag);
    if (!attrMatch) {
      continue;
    }

    const value = attrMatch[1] || attrMatch[2] || attrMatch[3];
    if (!IMAGE_EXT_PATTERN.test(value)) {
      continue;
    }

    const valueStartInAttr = attrMatch[0].indexOf(value);
    refs.push({
      start: tagMatch.index + attrMatch.index + valueStartInAttr,
      end: tagMatch.index + attrMatch.index + valueStartInAttr + value.length,
      value,
    });
  }

  return refs;
}

function collectRequireRefs(content, protectedRanges) {
  const refs = [];
  const regex = /require\(\s*(['"`])([^'"`]+)\1\s*\)/g;
  let match;

  while ((match = regex.exec(content)) !== null) {
    if (isProtected(protectedRanges, match.index) || !IMAGE_EXT_PATTERN.test(match[2])) {
      continue;
    }

    const valueStart = match.index + match[0].indexOf(match[2]);
    refs.push({
      start: valueStart,
      end: valueStart + match[2].length,
      value: match[2],
    });
  }

  return refs;
}

function collectImportRefs(content, protectedRanges) {
  const refs = [];
  const regex = /import\s+[^;]+?\s+from\s*(['"`])([^'"`]+\.(?:png|jpe?g|gif|webp)(?:[?#][^'"`]*)?)\1/g;
  let match;

  while ((match = regex.exec(content)) !== null) {
    if (isProtected(protectedRanges, match.index)) {
      continue;
    }

    const valueStart = match.index + match[0].indexOf(match[2]);
    refs.push({
      start: valueStart,
      end: valueStart + match[2].length,
      value: match[2],
    });
  }

  return refs;
}

function collectFrontmatterRefs(content) {
  const match = /^---\r?\n([\s\S]*?)\r?\n---/.exec(content);
  if (!match) {
    return [];
  }

  const frontmatter = match[1];
  const frontmatterStart = match[0].indexOf(frontmatter);
  const keyPattern = FRONTMATTER_IMAGE_KEYS.join('|');
  const regex = new RegExp(
    `^(\\s*(?:${keyPattern})\\s*:\\s*)(['"]?)([^'"\\r\\n]+?\\.(?:png|jpe?g|gif|webp)(?:[?#][^'"\\s\\r\\n]*)?)\\2\\s*$`,
    'gim',
  );
  const refs = [];
  let lineMatch;

  while ((lineMatch = regex.exec(frontmatter)) !== null) {
    const value = lineMatch[3];
    const valueStart = frontmatterStart + lineMatch.index + lineMatch[0].indexOf(value);
    refs.push({
      start: valueStart,
      end: valueStart + value.length,
      value,
    });
  }

  return refs;
}

function resolveImageRef(value, docFile, contentDirs) {
  if (isExternalRef(value)) {
    return null;
  }

  const { pathname } = splitSuffix(value);
  const ext = path.extname(pathname).toLowerCase();
  if (![...CONVERTIBLE_EXTS, '.webp'].includes(ext)) {
    return null;
  }

  const decodedPathname = safeDecodeUri(pathname);
  let sourcePath;

  // Docusaurus 中 /xxx 指向 static/xxx；相对路径则按当前文档所在目录解析。
  if (decodedPathname.startsWith('@site/')) {
    sourcePath = path.resolve(ROOT_DIR, decodedPathname.slice('@site/'.length));
  } else if (decodedPathname.startsWith('/')) {
    sourcePath = path.resolve(STATIC_DIR, decodedPathname.slice(1));
  } else {
    sourcePath = path.resolve(path.dirname(docFile), decodedPathname);
  }

  if (!isInsideDir(sourcePath, ROOT_DIR)) {
    return null;
  }

  return { sourcePath, ext, scope: getPathScope(sourcePath, contentDirs) };
}

function getPathScope(sourcePath, contentDirs) {
  if (contentDirs.some((dir) => isInsideDir(sourcePath, dir))) {
    return 'content';
  }

  if (isInsideDir(sourcePath, STATIC_DIR)) {
    return 'static';
  }

  return 'site';
}

function isExternalRef(value) {
  return (
    /^(?:[a-z][a-z0-9+.-]*:)?\/\//i.test(value) ||
    /^(?:data|mailto|tel):/i.test(value) ||
    value.startsWith('#')
  );
}

function splitSuffix(value) {
  const suffixIndex = value.search(/[?#]/);
  if (suffixIndex === -1) {
    return { pathname: value, suffix: '' };
  }
  return {
    pathname: value.slice(0, suffixIndex),
    suffix: value.slice(suffixIndex),
  };
}

function safeDecodeUri(value) {
  try {
    return decodeURI(value);
  } catch (error) {
    return value;
  }
}

function isInsideDir(file, dir) {
  const relative = path.relative(dir, file);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}

async function exists(file) {
  try {
    await fs.access(file);
    return true;
  } catch (error) {
    return false;
  }
}

async function isDirectory(dir) {
  try {
    return (await fs.stat(dir)).isDirectory();
  } catch (error) {
    return false;
  }
}

function unique(values) {
  return [...new Set(values)];
}

async function allocateTargets(sourcePaths) {
  const groups = new Map();
  for (const sourcePath of sourcePaths) {
    const ext = path.extname(sourcePath);
    const baseTarget = path.join(path.dirname(sourcePath), `${path.basename(sourcePath, ext)}.webp`);
    if (!groups.has(baseTarget)) {
      groups.set(baseTarget, []);
    }
    groups.get(baseTarget).push(sourcePath);
  }

  const allocated = new Map();
  const usedTargets = new Set();

  // 处理同目录同名不同后缀的冲突，例如 image.png 和 image.jpg 不能都写成 image.webp。
  for (const [baseTarget, group] of groups) {
    if (group.length === 1) {
      allocated.set(group[0], await reserveAvailableTarget(baseTarget, usedTargets));
      continue;
    }

    for (const sourcePath of group) {
      const ext = path.extname(sourcePath).slice(1).toLowerCase();
      const stem = path.basename(sourcePath, path.extname(sourcePath));
      const candidate = path.join(path.dirname(sourcePath), `${stem}-${ext}.webp`);
      allocated.set(sourcePath, await reserveAvailableTarget(candidate, usedTargets));
    }
  }

  return allocated;
}

async function reserveAvailableTarget(baseTarget, usedTargets) {
  let candidate = baseTarget;
  let index = 2;
  const ext = path.extname(baseTarget);
  const stem = path.basename(baseTarget, ext);
  const dir = path.dirname(baseTarget);

  while (usedTargets.has(candidate) || (await exists(candidate))) {
    candidate = path.join(dir, `${stem}-${index}${ext}`);
    index += 1;
  }

  usedTargets.add(candidate);
  return candidate;
}

async function convertToWebp(sourcePath, targetPath) {
  await fs.mkdir(path.dirname(targetPath), { recursive: true });
  const tempPath = `${targetPath}.tmp-${process.pid}`;
  const animated = path.extname(sourcePath).toLowerCase() === '.gif';
  const webpOptions = lossless
    ? { lossless: true, effort: 4 }
    : { quality, effort: 4 };

  try {
    // 先写临时文件，再原子替换为目标文件，避免转换中断时留下损坏的 .webp。
    await sharp(sourcePath, { animated, limitInputPixels: false })
      .webp(webpOptions)
      .toFile(tempPath);
    await fs.rename(tempPath, targetPath);
  } catch (error) {
    await fs.rm(tempPath, { force: true }).catch(() => {});
    throw error;
  }
}

function buildRefValue(originalValue, docFile, targetPath) {
  const { suffix } = splitSuffix(originalValue);
  const { pathname } = splitSuffix(originalValue);
  const decodedPathname = safeDecodeUri(pathname);

  if (decodedPathname.startsWith('@site/')) {
    return `@site/${toPosix(path.relative(ROOT_DIR, targetPath))}${suffix}`;
  }

  if (decodedPathname.startsWith('/') && isInsideDir(targetPath, STATIC_DIR)) {
    return `/${toPosix(path.relative(STATIC_DIR, targetPath))}${suffix}`;
  }

  let relative = toPosix(path.relative(path.dirname(docFile), targetPath));
  if (!relative.startsWith('.')) {
    if (decodedPathname.startsWith('./')) {
      relative = `./${relative}`;
    }
  }

  return `${relative}${suffix}`;
}

function applyReplacements(content, replacements) {
  const ordered = [...replacements].sort((a, b) => b.start - a.start);
  let updated = content;

  for (const replacement of ordered) {
    updated =
      updated.slice(0, replacement.start) +
      replacement.value +
      updated.slice(replacement.end);
  }

  return updated;
}

async function findRemainingRefs(sourcePaths, contentDirs, plannedContents) {
  const remainingRefs = new Map(sourcePaths.map((sourcePath) => [sourcePath, []]));
  if (sourcePaths.length === 0) {
    return remainingRefs;
  }

  const sourceSet = new Set(sourcePaths);
  const siteFiles = await walkScannableFiles(ROOT_DIR);

  for (const file of siteFiles) {
    let content;
    if (plannedContents.has(file)) {
      content = plannedContents.get(file);
    } else {
      try {
        content = await fs.readFile(file, 'utf8');
      } catch (error) {
        continue;
      }
    }

    const refs = collectSafetyRefs(content, file);
    const seen = new Set();
    for (const ref of refs) {
      const resolved = resolveImageRef(ref.value, file, contentDirs);
      if (!resolved || !sourceSet.has(resolved.sourcePath)) {
        continue;
      }

      const key = `${file}\0${ref.value}`;
      if (seen.has(key)) {
        continue;
      }
      seen.add(key);

      remainingRefs.get(resolved.sourcePath).push({
        file,
        value: ref.value,
      });
    }
  }

  return remainingRefs;
}

async function walkScannableFiles(dir) {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (SKIP_DIR_NAMES.has(entry.name)) {
        continue;
      }
      files.push(...(await walkScannableFiles(fullPath)));
    } else if (entry.isFile() && SCANNABLE_EXTS.has(path.extname(entry.name))) {
      files.push(fullPath);
    }
  }

  return files;
}

function collectSafetyRefs(content, file) {
  const protectedRanges = DOC_EXTS.has(path.extname(file)) ? collectProtectedRanges(content) : [];
  const refs = [
    ...collectHtmlImageRefs(content, protectedRanges),
    ...collectRequireRefs(content, protectedRanges),
    ...collectImportRefs(content, protectedRanges),
    ...collectGenericImageRefs(content, protectedRanges),
  ];

  if (DOC_EXTS.has(path.extname(file))) {
    refs.push(
      ...collectMarkdownImageRefs(content, protectedRanges),
      ...collectFrontmatterRefs(content),
    );
  }

  return refs;
}

function collectGenericImageRefs(content, protectedRanges) {
  return [
    ...collectQuotedImageRefs(content, protectedRanges),
    ...collectCssUrlRefs(content, protectedRanges),
  ];
}

function collectQuotedImageRefs(content, protectedRanges) {
  const refs = [];
  const regex = /(['"`])([^'"`\r\n]+?\.(?:png|jpe?g|gif|webp)(?:[?#][^'"`\r\n]*)?)\1/gi;
  let match;

  while ((match = regex.exec(content)) !== null) {
    if (isProtected(protectedRanges, match.index)) {
      continue;
    }

    const valueStart = match.index + match[0].indexOf(match[2]);
    refs.push({
      start: valueStart,
      end: valueStart + match[2].length,
      value: match[2],
    });
  }

  return refs;
}

function collectCssUrlRefs(content, protectedRanges) {
  const refs = [];
  const regex = /\burl\(\s*(?:"([^"]+)"|'([^']+)'|([^'")\s]+))\s*\)/gi;
  let match;

  while ((match = regex.exec(content)) !== null) {
    if (isProtected(protectedRanges, match.index)) {
      continue;
    }

    const value = match[1] || match[2] || match[3];
    if (!IMAGE_EXT_PATTERN.test(value)) {
      continue;
    }

    const valueStart = match.index + match[0].indexOf(value);
    refs.push({
      start: valueStart,
      end: valueStart + value.length,
      value,
    });
  }

  return refs;
}

function rangeContaining(ranges, index) {
  return ranges.find((range) => index >= range.start && index < range.end);
}

function isProtected(ranges, index) {
  return Boolean(rangeContaining(ranges, index));
}

function toPosix(value) {
  return value.split(path.sep).join('/');
}

function printSummary(summary) {
  const mode = dryRun ? 'dry run' : 'write';
  console.log(`content image conversion (${mode})`);
  console.log(`- webp mode: ${lossless ? 'lossless' : `quality ${quality}`}`);
  console.log(`- static image refs: ${includeStatic ? 'included' : 'skipped'}`);
  console.log(`- content roots scanned: ${summary.contentDirs}`);
  console.log(`- content files scanned: ${summary.contentFiles}`);
  console.log(`- local image refs found: ${summary.localRefs}`);
  console.log(`- static image refs skipped: ${summary.skippedStaticRefs.length}`);
  console.log(`- other site refs skipped: ${summary.skippedSiteRefs.length}`);
  console.log(`- existing local refs: ${summary.existingRefs}`);
  console.log(`- already webp refs: ${summary.alreadyWebpRefs}`);
  console.log(`- source images to convert: ${summary.sourceCount}`);
  console.log(`- converted images: ${summary.convertedCount}`);
  console.log(`- original images ${dryRun ? 'to delete' : 'deleted'}: ${summary.deletedCount}`);
  console.log(`- original images kept due to remaining refs: ${summary.blockedDeletes.length}`);
  console.log(`- content files updated: ${summary.updatedDocCount}`);
  console.log(`- refs updated: ${summary.updatedRefCount}`);

  if (summary.skippedStaticRefs.length > 0) {
    console.log('- skipped static refs:');
    for (const ref of summary.skippedStaticRefs.slice(0, 10)) {
      console.log(`  - ${toPosix(path.relative(ROOT_DIR, ref.docFile))}: ${ref.value}`);
    }
    if (summary.skippedStaticRefs.length > 10) {
      console.log(`  - ... ${summary.skippedStaticRefs.length - 10} more`);
    }
  }

  if (summary.missingRefs.length > 0) {
    console.log(`- missing image refs: ${summary.missingRefs.length}`);
    for (const ref of summary.missingRefs.slice(0, 10)) {
      console.log(`  - ${toPosix(path.relative(ROOT_DIR, ref.docFile))}: ${ref.value}`);
    }
    if (summary.missingRefs.length > 10) {
      console.log(`  - ... ${summary.missingRefs.length - 10} more`);
    }
  }

  if (summary.failed.length > 0) {
    console.log(`- failed conversions: ${summary.failed.length}`);
    for (const failure of summary.failed.slice(0, 10)) {
      console.log(`  - ${toPosix(path.relative(ROOT_DIR, failure.sourcePath))}: ${failure.error.message}`);
    }
    if (summary.failed.length > 10) {
      console.log(`  - ... ${summary.failed.length - 10} more`);
    }
  }

  if (summary.blockedDeletes.length > 0) {
    console.log('- originals kept because references remain:');
    for (const blocked of summary.blockedDeletes.slice(0, 10)) {
      console.log(`  - ${toPosix(path.relative(ROOT_DIR, blocked.sourcePath))}`);
      for (const ref of blocked.refs.slice(0, 3)) {
        console.log(`    - ${toPosix(path.relative(ROOT_DIR, ref.file))}: ${ref.value}`);
      }
      if (blocked.refs.length > 3) {
        console.log(`    - ... ${blocked.refs.length - 3} more refs`);
      }
    }
    if (summary.blockedDeletes.length > 10) {
      console.log(`  - ... ${summary.blockedDeletes.length - 10} more`);
    }
  }

  if (summary.deleteFailed.length > 0) {
    console.log(`- failed deletions: ${summary.deleteFailed.length}`);
    for (const failure of summary.deleteFailed.slice(0, 10)) {
      console.log(`  - ${toPosix(path.relative(ROOT_DIR, failure.sourcePath))}: ${failure.error.message}`);
    }
    if (summary.deleteFailed.length > 10) {
      console.log(`  - ... ${summary.deleteFailed.length - 10} more`);
    }
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
