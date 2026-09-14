import { useCallback, useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { PrismLight as SyntaxHighlighter } from 'react-syntax-highlighter';
import { coldarkDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import bash from 'react-syntax-highlighter/dist/esm/languages/prism/bash';
import c from 'react-syntax-highlighter/dist/esm/languages/prism/c';
import cpp from 'react-syntax-highlighter/dist/esm/languages/prism/cpp';
import csharp from 'react-syntax-highlighter/dist/esm/languages/prism/csharp';
import css from 'react-syntax-highlighter/dist/esm/languages/prism/css';
import go from 'react-syntax-highlighter/dist/esm/languages/prism/go';
import java from 'react-syntax-highlighter/dist/esm/languages/prism/java';
import javascript from 'react-syntax-highlighter/dist/esm/languages/prism/javascript';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import jsx from 'react-syntax-highlighter/dist/esm/languages/prism/jsx';
import markdown from 'react-syntax-highlighter/dist/esm/languages/prism/markdown';
import python from 'react-syntax-highlighter/dist/esm/languages/prism/python';
import rust from 'react-syntax-highlighter/dist/esm/languages/prism/rust';
import sql from 'react-syntax-highlighter/dist/esm/languages/prism/sql';
import tsx from 'react-syntax-highlighter/dist/esm/languages/prism/tsx';
import typescript from 'react-syntax-highlighter/dist/esm/languages/prism/typescript';
import yaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml';
import {
  IconChevronRight,
  IconChevronDown,
  IconCheck,
  IconCopy,
} from '@tabler/icons-react';
import { formatElapsed } from './ChatActivity';

// The full Prism export eagerly bundles hundreds of language grammars into the
// startup-critical Chat route. Workspace/chat output overwhelmingly uses this
// focused set; unknown fences still render safely as plain code.
SyntaxHighlighter.registerLanguage('bash', bash);
SyntaxHighlighter.registerLanguage('c', c);
SyntaxHighlighter.registerLanguage('cpp', cpp);
SyntaxHighlighter.registerLanguage('csharp', csharp);
SyntaxHighlighter.registerLanguage('css', css);
SyntaxHighlighter.registerLanguage('go', go);
SyntaxHighlighter.registerLanguage('java', java);
SyntaxHighlighter.registerLanguage('javascript', javascript);
SyntaxHighlighter.registerLanguage('json', json);
SyntaxHighlighter.registerLanguage('jsx', jsx);
SyntaxHighlighter.registerLanguage('markdown', markdown);
SyntaxHighlighter.registerLanguage('python', python);
SyntaxHighlighter.registerLanguage('rust', rust);
SyntaxHighlighter.registerLanguage('sql', sql);
SyntaxHighlighter.registerLanguage('tsx', tsx);
SyntaxHighlighter.registerLanguage('typescript', typescript);
SyntaxHighlighter.registerLanguage('yaml', yaml);

const CODE_LANGUAGE_ALIASES: Record<string, string> = {
  cs: 'csharp',
  js: 'javascript',
  md: 'markdown',
  py: 'python',
  sh: 'bash',
  shell: 'bash',
  ts: 'typescript',
  yml: 'yaml',
};

// Reasoning-capable models emit their hidden chain-of-thought wrapped in one
// of several tag conventions. We detect them all and render as collapsible
// "Thoughts" / "Reasoning" / "Reflection" blocks.
//
// Coverage (case-insensitive on the tag name, attributes allowed for robustness):
//   <think>      DeepSeek R1 / R1-distill, QwQ, Marco-o1, R1-style fine-tunes
//   <thinking>   Some Claude prompting patterns, certain Llama fine-tunes
//   <thought>    Hermes thinking variants, certain agentic fine-tunes
//   <reasoning>  Various open-source reasoning models
//   <reflection> Reflection-70B and derivatives

type ReasoningKind =
  | 'think'
  | 'thinking'
  | 'thought'
  | 'reasoning'
  | 'reflection';

type Segment =
  | { kind: 'text'; content: string }
  | {
      kind: 'reasoning';
      tag: ReasoningKind;
      content: string;
      complete: boolean;
    };

const REASONING_TAGS: ReasoningKind[] = [
  'think',
  'thinking',
  'thought',
  'reasoning',
  'reflection',
];

const LABELS_STREAMING: Record<ReasoningKind, string> = {
  think: 'Thinking',
  thinking: 'Thinking',
  thought: 'Thinking',
  reasoning: 'Reasoning',
  reflection: 'Reflecting',
};

const LABELS_COMPLETE: Record<ReasoningKind, string> = {
  think: 'Thoughts',
  thinking: 'Thoughts',
  thought: 'Thoughts',
  reasoning: 'Reasoning',
  reflection: 'Reflection',
};

// Single regex that matches an opening tag (any of the configured tag names)
// — case-insensitive, attributes tolerated (e.g. `<think type="x">`).
const OPEN_TAG_RE = new RegExp(
  `<(${REASONING_TAGS.join('|')})(?:\\s[^>]*)?>`,
  'i',
);

function buildCloseRe(tag: string): RegExp {
  return new RegExp(`</${tag}\\s*>`, 'i');
}

function parseSegments(text: string): Segment[] {
  const segments: Segment[] = [];
  let cursor = 0;

  while (cursor < text.length) {
    const remaining = text.slice(cursor);
    const openMatch = remaining.match(OPEN_TAG_RE);
    if (!openMatch || openMatch.index === undefined) {
      const tail = remaining;
      if (tail.length > 0) segments.push({ kind: 'text', content: tail });
      break;
    }

    const openOffset = openMatch.index;
    const openLen = openMatch[0].length;
    const tag = openMatch[1].toLowerCase() as ReasoningKind;

    if (openOffset > 0) {
      segments.push({ kind: 'text', content: remaining.slice(0, openOffset) });
    }

    const contentStart = cursor + openOffset + openLen;
    const closeRe = buildCloseRe(tag);
    const tailFromContent = text.slice(contentStart);
    const closeMatch = tailFromContent.match(closeRe);
    if (!closeMatch || closeMatch.index === undefined) {
      // Unclosed tag — we're mid-stream inside a reasoning block.
      segments.push({
        kind: 'reasoning',
        tag,
        content: tailFromContent,
        complete: false,
      });
      cursor = text.length;
    } else {
      const closeOffset = closeMatch.index;
      const closeLen = closeMatch[0].length;
      segments.push({
        kind: 'reasoning',
        tag,
        content: tailFromContent.slice(0, closeOffset),
        complete: true,
      });
      cursor = contentStart + closeOffset + closeLen;
    }
  }

  return segments;
}

const ThinkingContainer = styled.div`
  border-left: 2px solid rgba(33, 220, 143, 0.35);
  margin: 6px 0 10px;
  padding: 2px 0 2px 12px;
`;

const ThinkingHeader = styled.button`
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.55);
  font-size: 0.92em;
  font-weight: 500;
  letter-spacing: 0.3px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  user-select: none;

  &:hover {
    color: rgba(255, 255, 255, 0.85);
  }

  &:focus {
    outline: none;
  }
`;

const EmptyTag = styled.span`
  margin-left: 6px;
  padding: 1px 6px;
  font-size: 0.78em;
  font-weight: 500;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.5);
  background: rgba(255, 255, 255, 0.08);
  border-radius: 3px;
`;

// Tabular numerals keep the header from jittering as the count ticks.
const ThinkingElapsed = styled.span`
  margin-left: 6px;
  font-size: 0.82em;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.35);
  font-variant-numeric: tabular-nums;
`;

const ThinkingDots = styled.span`
  display: inline-block;
  margin-left: 4px;
  &::after {
    content: '…';
    display: inline-block;
    animation: thinkingDots 1.4s infinite steps(4, end);
    overflow: hidden;
    vertical-align: bottom;
    width: 0;
  }
  @keyframes thinkingDots {
    0% {
      width: 0;
    }
    33% {
      width: 0.4em;
    }
    66% {
      width: 0.8em;
    }
    100% {
      width: 1.2em;
    }
  }
`;

const ThinkingBody = styled.div<{ $hidden: boolean }>`
  display: ${(p) => (p.$hidden ? 'none' : 'block')};
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.55);
  font-size: 0.95em;
  font-style: italic;
  white-space: pre-wrap;

  p {
    margin: 0 0 0.5em;
  }
  p:last-child {
    margin-bottom: 0;
  }
`;

const CodeBlockWrapper = styled.div`
  position: relative;
  margin: 8px 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  overflow: hidden;

  pre {
    margin: 0 !important;
    border-radius: 0 !important;
  }
`;

const CodeBlockHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 6px 4px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
`;

const CodeBlockLanguage = styled.span`
  font-size: 0.75em;
  font-weight: 500;
  letter-spacing: 0.6px;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.45);
  font-family: inherit;
`;

const CopyButton = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  border-radius: 4px;
  padding: 3px 7px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.55);
  font-size: 0.78em;
  font-weight: 500;
  font-family: inherit;
  transition: background 0.15s ease, color 0.15s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.9);
  }

  &:focus {
    outline: none;
  }
`;

// Electron's renderer exposes the async clipboard API, but it is unavailable in
// non-secure contexts and in the test environment, so fall back rather than
// letting a missing API throw inside a click handler.
async function writeClipboardText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const scratch = document.createElement('textarea');
  scratch.value = text;
  scratch.setAttribute('readonly', '');
  scratch.style.position = 'fixed';
  scratch.style.opacity = '0';
  document.body.appendChild(scratch);
  scratch.select();
  try {
    document.execCommand('copy');
  } finally {
    document.body.removeChild(scratch);
  }
}

function CodeBlock({ language, code }: { language: string; code: string }) {
  const [copied, setCopied] = useState(false);
  const resetTimer = useRef<ReturnType<typeof setTimeout>>();

  useEffect(
    () => () => {
      if (resetTimer.current) clearTimeout(resetTimer.current);
    },
    [],
  );

  const onCopy = useCallback(async () => {
    try {
      await writeClipboardText(code);
      setCopied(true);
      if (resetTimer.current) clearTimeout(resetTimer.current);
      resetTimer.current = setTimeout(() => setCopied(false), 1600);
    } catch {
      // A failed copy should not take the message down with it.
    }
  }, [code]);

  return (
    <CodeBlockWrapper>
      <CodeBlockHeader>
        <CodeBlockLanguage>{language}</CodeBlockLanguage>
        <CopyButton
          type="button"
          onClick={onCopy}
          aria-label={copied ? 'Copied' : 'Copy code'}
        >
          {copied ? (
            <IconCheck size={13} stroke={2} />
          ) : (
            <IconCopy size={13} stroke={1.8} />
          )}
          {copied ? 'Copied' : 'Copy'}
        </CopyButton>
      </CodeBlockHeader>
      <SyntaxHighlighter PreTag="div" language={language} style={coldarkDark}>
        {code}
      </SyntaxHighlighter>
    </CodeBlockWrapper>
  );
}

// GFM tables arrive as bare <table>/<th>/<td>, which inherit nothing from the
// bubble and render as unreadable runs of text. These give them the same
// borders and spacing the rest of the message already uses.
const MarkdownTableWrapper = styled.div`
  overflow-x: auto;
  margin: 8px 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.94em;
  }

  th,
  td {
    padding: 7px 12px;
    text-align: left;
    vertical-align: top;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  th {
    font-weight: 600;
    background: rgba(255, 255, 255, 0.05);
    white-space: nowrap;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }

  tbody tr:hover td {
    background: rgba(255, 255, 255, 0.03);
  }
`;

const markdownComponents = {
  code(props: any) {
    const { children, className, node, ...rest } = props;
    const match = /language-(\w+)/.exec(className || '');
    if (!match) {
      return (
        <code {...rest} className={className}>
          {children}
        </code>
      );
    }
    const raw = match[1].toLowerCase();
    return (
      <CodeBlock
        language={CODE_LANGUAGE_ALIASES[raw] ?? raw}
        code={String(children).replace(/\n$/, '')}
      />
    );
  },
  table(props: any) {
    const { node, ...rest } = props;
    return (
      <MarkdownTableWrapper>
        <table {...rest} />
      </MarkdownTableWrapper>
    );
  },
};

const remarkPlugins = [remarkGfm];

function ReasoningBlock({
  tag,
  content,
  complete,
}: {
  tag: ReasoningKind;
  content: string;
  complete: boolean;
}) {
  // While streaming, show the reasoning as it arrives so the user has feedback.
  // The moment the closing tag arrives, auto-collapse the block once. After
  // that the user can toggle freely without us flipping it back.
  const [open, setOpen] = useState(true);
  const prevComplete = useRef(complete);

  // How long the model spent thinking. Only measured when we actually watched
  // the block stream — a message reloaded from history arrives already closed,
  // and inventing a duration for it would be a lie.
  const startedAt = useRef<number | undefined>(
    complete ? undefined : Date.now(),
  );
  const [elapsedMs, setElapsedMs] = useState(0);
  const [finalMs, setFinalMs] = useState<number | undefined>(undefined);

  useEffect(() => {
    if (complete || startedAt.current === undefined) return undefined;
    const id = setInterval(() => {
      if (startedAt.current !== undefined) {
        setElapsedMs(Date.now() - startedAt.current);
      }
    }, 250);
    return () => clearInterval(id);
  }, [complete]);

  useEffect(() => {
    if (!prevComplete.current && complete) {
      setOpen(false);
      if (startedAt.current !== undefined) {
        setFinalMs(Date.now() - startedAt.current);
      }
    }
    prevComplete.current = complete;
  }, [complete]);

  const trimmed = content.trim();
  const isEmpty = complete && trimmed.length === 0;

  const Caret = open ? IconChevronDown : IconChevronRight;
  const label = complete ? LABELS_COMPLETE[tag] : LABELS_STREAMING[tag];
  // The count keeps running in the collapsed header, so the user can still see
  // what the wait cost them after the block folds itself away.
  const shownMs = complete ? finalMs : elapsedMs;

  return (
    <ThinkingContainer>
      <ThinkingHeader
        type="button"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <Caret size={14} stroke={1.8} />
        {complete ? (
          label
        ) : (
          <>
            {label}
            <ThinkingDots />
          </>
        )}
        {shownMs !== undefined && shownMs >= 1000 && (
          <ThinkingElapsed>{formatElapsed(shownMs)}</ThinkingElapsed>
        )}
        {isEmpty && <EmptyTag>empty</EmptyTag>}
      </ThinkingHeader>
      <ThinkingBody $hidden={!open}>
        {isEmpty ? (
          <em style={{ opacity: 0.6 }}>
            No reasoning content emitted by the provider.
          </em>
        ) : (
          <Markdown components={markdownComponents} remarkPlugins={remarkPlugins}>
            {trimmed}
          </Markdown>
        )}
      </ThinkingBody>
    </ThinkingContainer>
  );
}

export function ThinkingMessageBody({ text }: { text: string }) {
  const segments = parseSegments(text);

  // Fast path: no reasoning tag anywhere → plain markdown.
  if (segments.every((s) => s.kind === 'text')) {
    return (
      <Markdown components={markdownComponents} remarkPlugins={remarkPlugins}>
        {text}
      </Markdown>
    );
  }

  return (
    <>
      {segments.map((seg, i) => {
        if (seg.kind === 'reasoning') {
          return (
            <ReasoningBlock
              key={`reasoning-${i}`}
              tag={seg.tag}
              content={seg.content}
              complete={seg.complete}
            />
          );
        }
        const trimmed = seg.content.replace(/^\s+/, '');
        if (trimmed.length === 0) return null;
        return (
          <Markdown
            key={`text-${i}`}
            components={markdownComponents}
            remarkPlugins={remarkPlugins}
          >
            {trimmed}
          </Markdown>
        );
      })}
    </>
  );
}
