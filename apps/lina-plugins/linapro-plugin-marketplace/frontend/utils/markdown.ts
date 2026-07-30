import hljs from "highlight.js/lib/common";
import MarkdownIt from "markdown-it";

// Marketplace documentation is authored by plugin publishers. Keep HTML
// disabled so Markdown rendering cannot inject script tags into the workbench.
const marketplaceMarkdown = new MarkdownIt({
  breaks: false,
  highlight(code, language) {
    const lang = (language || "").trim().split(/\s+/u)[0] || "";
    if (lang.toLowerCase() === "mermaid") {
      // Mermaid fences are handled by the custom fence rule below.
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
        // Fall through to escaped plain code.
      }
    }
    return `<pre class="hljs marketplace-code-block"><code class="hljs">${escapeHtml(code)}</code></pre>`;
  },
  html: false,
  linkify: true,
  typographer: false,
});

// Default markdown-it validateLink blocks every data: URL. Allow only
// data:image/* so docs can embed small inline images while other data:
// schemes stay rejected (and javascript: remains blocked by the default rule).
const defaultValidateLink = marketplaceMarkdown.validateLink.bind(
  marketplaceMarkdown,
);
marketplaceMarkdown.validateLink = (url) => {
  const value = String(url || "")
    .trim()
    .toLowerCase();
  if (value.startsWith("data:image/")) {
    return true;
  }
  return defaultValidateLink(url);
};

const defaultFence =
  marketplaceMarkdown.renderer.rules.fence ||
  function fence(tokens, idx, options, _env, self) {
    return self.renderToken(tokens, idx, options);
  };

marketplaceMarkdown.renderer.rules.fence = (tokens, idx, options, env, self) => {
  const token = tokens[idx];
  const info = (token.info || "").trim();
  const lang = info.split(/\s+/u)[0]?.toLowerCase() || "";
  if (lang === "mermaid") {
    const source = token.content.replace(/\n$/u, "");
    // Keep the source as text content only (html:false path). Mermaid reads
    // textContent after mount; data-processed guards re-runs.
    return `<div class="marketplace-mermaid-wrap"><pre class="mermaid marketplace-mermaid">${escapeHtml(source)}</pre></div>\n`;
  }
  return defaultFence(tokens, idx, options, env, self);
};

// Make images responsive and safe (no javascript: URLs).
const defaultImage =
  marketplaceMarkdown.renderer.rules.image ||
  function image(tokens, idx, options, _env, self) {
    return self.renderToken(tokens, idx, options);
  };

marketplaceMarkdown.renderer.rules.image = (tokens, idx, options, env, self) => {
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

function isUnsafeImageSrc(src: string) {
  const value = src.trim().toLowerCase();
  if (!value) {
    return true;
  }
  if (value.startsWith("javascript:") || value.startsWith("vbscript:")) {
    return true;
  }
  if (value.startsWith("data:") && !value.startsWith("data:image/")) {
    return true;
  }
  return false;
}

function escapeHtml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function escapeHtmlAttr(value: string) {
  return escapeHtml(value).replaceAll("`", "&#96;");
}

export function renderMarketplaceMarkdown(source: string | null | undefined) {
  const markdown = typeof source === "string" ? source : "";
  if (!markdown.trim()) {
    return "";
  }
  return marketplaceMarkdown.render(markdown);
}

/**
 * Render Mermaid diagrams inside a Markdown body container after v-html mount.
 * Failures degrade in-place: the pre.mermaid source remains visible as code.
 */
export async function enhanceMarketplaceMarkdown(
  root: HTMLElement | null | undefined,
  options?: { dark?: boolean },
) {
  if (!root || typeof window === "undefined") {
    return;
  }
  const nodes = root.querySelectorAll<HTMLElement>(
    "pre.mermaid:not([data-processed])",
  );
  if (nodes.length === 0) {
    return;
  }

  try {
    const mermaidModule = await import("mermaid");
    const mermaid = mermaidModule.default;
    mermaid.initialize({
      // Publishers author diagrams; keep strict to avoid loose HTML injection.
      securityLevel: "strict",
      startOnLoad: false,
      theme: options?.dark ? "dark" : "default",
    });
    await mermaid.run({ nodes: Array.from(nodes) });
  } catch {
    // Leave the escaped mermaid source visible as a code-like block.
    for (const node of nodes) {
      node.classList.add("marketplace-mermaid-error");
      node.setAttribute("data-processed", "error");
    }
  }
}

export function resolveRelativeMarkdownPath(
  currentPath: string,
  href: string,
): null | string {
  const cleaned = href.trim().split("#")[0]?.split("?")[0] ?? "";
  if (
    !cleaned ||
    cleaned.includes("://") ||
    cleaned.startsWith("//") ||
    cleaned.startsWith("/") ||
    cleaned.startsWith("mailto:") ||
    cleaned.startsWith("data:")
  ) {
    return null;
  }
  if (!cleaned.toLowerCase().endsWith(".md")) {
    return null;
  }
  const baseDir = currentPath.includes("/")
    ? currentPath.slice(0, Math.max(0, currentPath.lastIndexOf("/") + 1))
    : "";
  const joined = `${baseDir}${cleaned}`.replaceAll("\\", "/");
  const segments = joined.split("/");
  const resolved: string[] = [];
  for (const segment of segments) {
    if (!segment || segment === ".") {
      continue;
    }
    if (segment === "..") {
      if (resolved.length === 0) {
        return null;
      }
      resolved.pop();
      continue;
    }
    resolved.push(segment);
  }
  return resolved.join("/") || null;
}
