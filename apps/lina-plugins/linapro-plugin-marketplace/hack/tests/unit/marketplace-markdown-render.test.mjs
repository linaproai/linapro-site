/**
 * Runtime unit tests for the marketplace Markdown render pipeline.
 * Imports the same production deps declared in frontend/package.json.
 */
import assert from "node:assert/strict";
import { createRequire } from "node:module";
import path from "node:path";
import { describe, it } from "node:test";
import { fileURLToPath } from "node:url";

const pluginRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../..",
);
const require = createRequire(
  path.join(pluginRoot, "frontend/package.json"),
);

const MarkdownIt = require("markdown-it");
const hljs = require("highlight.js/lib/common");

function escapeHtml(value) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function escapeHtmlAttr(value) {
  return escapeHtml(value).replaceAll("`", "&#96;");
}

function isUnsafeImageSrc(src) {
  const value = src.trim().toLowerCase();
  if (!value) return true;
  if (value.startsWith("javascript:") || value.startsWith("vbscript:")) {
    return true;
  }
  if (value.startsWith("data:") && !value.startsWith("data:image/")) {
    return true;
  }
  return false;
}

/**
 * Mirror of frontend/utils/markdown.ts render pipeline for Node unit isolation.
 * Keep in sync when changing highlight/mermaid/image rules.
 */
function createMarketplaceMarkdown() {
  const md = new MarkdownIt({
    breaks: false,
    highlight(code, language) {
      const lang = (language || "").trim().split(/\s+/u)[0] || "";
      if (lang.toLowerCase() === "mermaid") {
        return "";
      }
      if (lang && hljs.getLanguage(lang)) {
        try {
          const highlighted = hljs.highlight(code, {
            ignoreIllegals: true,
            language: lang,
          }).value;
          return `<pre class="hljs marketplace-code-block"><code class="hljs language-${escapeHtmlAttr(lang)}">${highlighted}</code></pre>`;
        } catch {
          // fall through
        }
      }
      return `<pre class="hljs marketplace-code-block"><code class="hljs">${escapeHtml(code)}</code></pre>`;
    },
    html: false,
    linkify: true,
    typographer: false,
  });

  const defaultValidateLink = md.validateLink.bind(md);
  md.validateLink = (url) => {
    const value = String(url || "")
      .trim()
      .toLowerCase();
    if (value.startsWith("data:image/")) {
      return true;
    }
    return defaultValidateLink(url);
  };

  const defaultFence =
    md.renderer.rules.fence ||
    function fence(tokens, idx, options, _env, self) {
      return self.renderToken(tokens, idx, options);
    };

  md.renderer.rules.fence = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    const info = (token.info || "").trim();
    const lang = info.split(/\s+/u)[0]?.toLowerCase() || "";
    if (lang === "mermaid") {
      const source = token.content.replace(/\n$/u, "");
      return `<div class="marketplace-mermaid-wrap"><pre class="mermaid marketplace-mermaid">${escapeHtml(source)}</pre></div>\n`;
    }
    return defaultFence(tokens, idx, options, env, self);
  };

  const defaultImage =
    md.renderer.rules.image ||
    function image(tokens, idx, options, _env, self) {
      return self.renderToken(tokens, idx, options);
    };

  md.renderer.rules.image = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    const srcIndex = token.attrIndex("src");
    if (srcIndex >= 0) {
      const src = token.attrs?.[srcIndex]?.[1] || "";
      if (isUnsafeImageSrc(src)) {
        const alt = token.content || "";
        return `<span class="marketplace-md-image-blocked" title="blocked image">${escapeHtml(alt || "image")}</span>`;
      }
    }
    token.attrSet("loading", "lazy");
    token.attrSet("class", "marketplace-md-image");
    return defaultImage(tokens, idx, options, env, self);
  };

  return md;
}

describe("marketplace markdown render pipeline", () => {
  const md = createMarketplaceMarkdown();

  it("highlights fenced code with language classes", () => {
    const html = md.render("```ts\nconst x: number = 1;\n```\n");
    assert.match(html, /pre class="hljs marketplace-code-block"/);
    assert.match(html, /language-ts/);
    assert.match(html, /hljs-/);
    assert.doesNotMatch(html, /```ts/);
  });

  it("renders tables and images", () => {
    const html = md.render(
      "| A | B |\n| --- | --- |\n| 1 | 2 |\n\n![logo](https://example.com/a.png)\n\n![pixel](data:image/png;base64,abc)\n",
    );
    assert.match(html, /<table>/);
    assert.match(html, /<th>/);
    assert.match(html, /class="marketplace-md-image"/);
    assert.match(html, /src="https:\/\/example\.com\/a\.png"/);
    assert.match(html, /src="data:image\/png;base64,abc"/);
    assert.match(html, /loading="lazy"/);
  });

  it("turns mermaid fences into mermaid pre blocks", () => {
    const html = md.render("```mermaid\nflowchart LR\n  A --> B\n```\n");
    assert.match(html, /class="marketplace-mermaid-wrap"/);
    assert.match(html, /class="mermaid marketplace-mermaid"/);
    assert.match(html, /flowchart LR/);
    assert.match(html, /A --&gt; B|A --> B/);
    assert.doesNotMatch(html, /```mermaid/);
  });

  it("blocks unsafe image sources and raw HTML injection", () => {
    // markdown-it validateLink already rejects javascript: for images, so the
    // fence stays plain text rather than an <img>. Extra isUnsafeImageSrc
    // covers non-image data: schemes if a token is still produced.
    const unsafeImage = md.render("![x](javascript:alert(1))\n");
    assert.doesNotMatch(unsafeImage, /<img/i);
    assert.doesNotMatch(unsafeImage, /src=["']javascript:/i);

    const blockedData = md.render("![x](data:text/html;base64,PHNjcmlwdD4=)\n");
    assert.match(blockedData, /marketplace-md-image-blocked|<p>!\[x\]/);
    assert.doesNotMatch(blockedData, /<img[^>]+src=["']data:text\/html/i);

    const rawHtml = md.render("<script>alert(1)</script>\n\n**ok**\n");
    assert.doesNotMatch(rawHtml, /<script>/i);
    assert.match(rawHtml, /<strong>ok<\/strong>/);
  });
});
