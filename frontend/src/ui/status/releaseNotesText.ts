/**
 * frontend/src/ui/status/releaseNotesText.ts
 *
 * Converts a GitHub release body (markdown) into clean plain text for the update
 * tooltip preview: strips the common markers (headings, bold, inline code,
 * strikethrough, links, images), drops horizontal rules, and normalizes list
 * bullets to "•". Intentionally lightweight — not a markdown parser; the fully
 * rendered notes are one click away on the release page. No HTML is produced.
 */
export const toPlainReleaseNotes = (markdown: string): string => {
  if (!markdown) {
    return '';
  }

  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const transformed = lines.map((line) => {
    let text = line;
    // Images have no place in a text preview — drop them entirely.
    text = replaceMarkdownMedia(text, '![', false);
    // Links [text](url) -> text.
    text = replaceMarkdownMedia(text, '[', true);
    // Horizontal rules (---, ***, ___) become a blank line.
    if (/^\s*([-*_])\1{2,}\s*$/.test(text)) {
      return '';
    }
    // Leading list marker (-, *, +) -> bullet, preserving indentation.
    text = text.replace(/^(\s*)[-*+]\s+/, '$1• ');
    // Leading heading (#…) / blockquote (>) markers -> drop marker, keep text.
    text = text.replace(/^\s*#{1,6}\s+/, '');
    text = text.replace(/^\s*>\s?/, '');
    // Inline emphasis / code / strikethrough markers.
    text = text.replace(/\*\*|__|~~|`/g, '');
    return text.trimEnd();
  });

  // Collapse runs of blank lines so stripped headings/rules don't leave gaps,
  // then drop leading/trailing blank lines (but keep the first line's indentation
  // so a leading nested bullet stays nested).
  const collapsed = transformed.join('\n').replace(/\n{3,}/g, '\n\n');
  // Scan rather than match: `/^\n+|\n+$/` backtracks quadratically on a long
  // newline run, and only newlines may go — the first line's indentation is
  // load-bearing, so trimStart() would flatten a leading nested bullet.
  let start = 0;
  while (start < collapsed.length && collapsed[start] === '\n') {
    start += 1;
  }
  let end = collapsed.length;
  while (end > start && collapsed[end - 1] === '\n') {
    end -= 1;
  }
  return collapsed.slice(start, end);
};
const replaceMarkdownMedia = (input: string, marker: '![' | '[', keepLabel: boolean): string => {
  let output = '';
  let cursor = 0;
  while (cursor < input.length) {
    const markerStart = input.indexOf(marker, cursor);
    if (markerStart === -1) {
      output += input.slice(cursor);
      break;
    }
    const labelStart = markerStart + marker.length;
    const labelEnd = input.indexOf(']', labelStart);
    if (labelEnd === -1) {
      output += input.slice(cursor);
      break;
    }
    if (input[labelEnd + 1] !== '(') {
      output += input.slice(cursor, labelStart);
      cursor = labelStart;
      continue;
    }
    const targetEnd = input.indexOf(')', labelEnd + 2);
    if (targetEnd === -1) {
      output += input.slice(cursor);
      break;
    }

    output += input.slice(cursor, markerStart);
    if (keepLabel) {
      output += input.slice(labelStart, labelEnd);
    }
    cursor = targetEnd + 1;
  }
  return output;
};
