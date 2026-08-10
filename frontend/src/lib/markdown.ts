import { marked } from 'marked';

// Custom extension for GitHub Alerts ([!TIP], [!NOTE], etc.)
marked.use({
  extensions: [{
    name: 'alert',
    level: 'block',
    start(src) { return src.match(/^> \[!/)?.index; },
    tokenizer(src) {
      const rule = /^> \[!(TIP|NOTE|IMPORTANT|WARNING|CAUTION)\][ \t]*\n((?:> .*(?:\n|$))*)/;
      const match = rule.exec(src);
      if (match) {
        return {
          type: 'alert',
          raw: match[0],
          alertType: match[1].toLowerCase(),
          text: match[2].replace(/^> /gm, '').trim()
        };
      }
    },
    renderer(token) {
      const icons: Record<string, string> = {
        tip: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A5 5 0 0 0 8 8c0 1.3.5 2.6 1.5 3.5.8.8 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>',
        note: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>',
        warning: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
        important: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
        caution: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>'
      };
      return `<div class="markdown-alert markdown-alert-${token.alertType}">
        <p class="markdown-alert-title">${icons[token.alertType] || icons.note}${token.alertType.toUpperCase()}</p>
        <div class="markdown-alert-content">${marked.parse(token.text)}</div>
      </div>`;
    }
  }]
});

export const parseMarkdown = async (text: string) => {
  return marked.parse(text, { gfm: true, breaks: true });
};

export { marked };
