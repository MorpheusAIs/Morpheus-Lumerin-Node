import { useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import Markdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { coldarkDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { IconChevronRight, IconChevronDown } from '@tabler/icons-react';

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

type ReasoningKind = 'think' | 'thinking' | 'thought' | 'reasoning' | 'reflection';

type Segment =
  | { kind: 'text'; content: string }
  | { kind: 'reasoning'; tag: ReasoningKind; content: string; complete: boolean };

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
    0% { width: 0; }
    33% { width: 0.4em; }
    66% { width: 0.8em; }
    100% { width: 1.2em; }
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

const markdownComponents = {
  code(props: any) {
    const { children, className, node, ...rest } = props;
    const match = /language-(\w+)/.exec(className || '');
    return match ? (
      <SyntaxHighlighter
        {...rest}
        PreTag="div"
        language={match[1]}
        style={coldarkDark}
      >
        {String(children).replace(/\n$/, '')}
      </SyntaxHighlighter>
    ) : (
      <code {...rest} className={className}>
        {children}
      </code>
    );
  },
};

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
  useEffect(() => {
    if (!prevComplete.current && complete) {
      setOpen(false);
    }
    prevComplete.current = complete;
  }, [complete]);

  const trimmed = content.trim();
  const isEmpty = complete && trimmed.length === 0;

  const Caret = open ? IconChevronDown : IconChevronRight;
  const label = complete ? LABELS_COMPLETE[tag] : LABELS_STREAMING[tag];

  return (
    <ThinkingContainer>
      <ThinkingHeader
        type="button"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <Caret size={14} stroke={1.8} />
        {complete ? label : (
          <>
            {label}<ThinkingDots />
          </>
        )}
        {isEmpty && <EmptyTag>empty</EmptyTag>}
      </ThinkingHeader>
      <ThinkingBody $hidden={!open}>
        {isEmpty ? (
          <em style={{ opacity: 0.6 }}>
            No reasoning content emitted by the provider.
          </em>
        ) : (
          <Markdown components={markdownComponents}>{trimmed}</Markdown>
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
      <Markdown components={markdownComponents}>{text}</Markdown>
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
          <Markdown key={`text-${i}`} components={markdownComponents}>
            {trimmed}
          </Markdown>
        );
      })}
    </>
  );
}
