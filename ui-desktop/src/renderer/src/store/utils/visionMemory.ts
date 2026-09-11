// Remembers which models have actually rejected an image.
//
// Vision capability is not published on-chain — model Tags carry modality
// (llm/stt/tts/embedding) and `tee`, nothing about image input. So the UI
// guesses from the model name, which is unavoidably incomplete.
//
// A provider rejecting an image is a *definitive* answer to the question the
// heuristic was guessing at. Recording it turns a static guess into something
// that self-corrects: the second time you attach an image to that model, the
// warning is a statement of fact rather than a maybe.
//
// Only negatives are stored. A successful image request means the model does
// support vision, which the name heuristic will usually have predicted anyway —
// and caching positives risks pinning a wrong answer if a provider swaps the
// model behind an ID.

const KEY = 'vision-rejected-models';

const read = (): string[] => {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch (e) {
    return [];
  }
};

const write = (ids: string[]) => {
  try {
    // Bounded so a long-lived install can't grow this forever.
    localStorage.setItem(KEY, JSON.stringify(ids.slice(-200)));
  } catch (e) {
    // Quota / private mode — this is an optimisation, not a requirement.
  }
};

/** True if this model previously refused an image. */
export const hasRejectedImages = (modelId?: string): boolean =>
  !!modelId && read().includes(modelId);

/** Records a definitive rejection reported by the provider. */
export const rememberImageRejection = (modelId?: string) => {
  if (!modelId) return;
  const ids = read();
  if (!ids.includes(modelId)) {
    write([...ids, modelId]);
  }
};

/** Clears a recorded rejection, e.g. if the user wants to retry. */
export const forgetImageRejection = (modelId?: string) => {
  if (!modelId) return;
  write(read().filter((id) => id !== modelId));
};

/** Detects the provider's "this model has no vision" refusal in any wrapping. */
export const isVisionRejection = (raw: unknown): boolean => {
  const msg = String(
    typeof raw === 'string' ? raw : ((raw as any)?.message ?? raw ?? ''),
  ).toLowerCase();

  return (
    (msg.includes('image') &&
      (msg.includes('not supported') || msg.includes('does not support'))) ||
    msg.includes('supports vision') ||
    msg.includes('vision model')
  );
};
