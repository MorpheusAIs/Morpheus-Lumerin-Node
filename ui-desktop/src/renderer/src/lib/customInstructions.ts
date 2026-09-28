// A user-written system prompt applied to every model request, app-wide.
//
// Why it is sent on every turn rather than once per chat: the proxy-router
// stores each request verbatim but replays only the LAST message of each stored
// turn back into the context (see AppendChatHistory in
// proxy-router/internal/chatstorage/genericchatstorage/interface.go). A system
// message sent only on the first turn would therefore be dropped from turn two
// onward. Sending it every turn costs a little storage and puts the instruction
// in front of the model exactly once per request, which is the behaviour the
// user asked for and needs no router change.

export const CUSTOM_INSTRUCTIONS_KEY = 'custom-instructions';

/** Beyond this the instruction stops being an instruction and starts being a
 * document the user pays to resend on every single turn. */
export const CUSTOM_INSTRUCTIONS_MAX_LENGTH = 4000;

export function loadCustomInstructions(): string {
  try {
    return window.localStorage.getItem(CUSTOM_INSTRUCTIONS_KEY) ?? '';
  } catch {
    return '';
  }
}

export function saveCustomInstructions(value: string): void {
  const trimmed = value.trim();
  try {
    if (trimmed) {
      window.localStorage.setItem(CUSTOM_INSTRUCTIONS_KEY, trimmed);
    } else {
      // An empty box means "no instructions", not "an instruction that is
      // blank" — storing whitespace would make the feature impossible to turn
      // off once it had been turned on.
      window.localStorage.removeItem(CUSTOM_INSTRUCTIONS_KEY);
    }
  } catch {
    // Preference storage is best-effort; the chat still works without it.
  }
}

export type SystemMessage = { role: 'system'; content: string };

/**
 * The system message to prepend to a request, or null when the user has not
 * written any instructions. Null rather than an empty message so the feature is
 * completely inert when unused — an empty system message is still a message,
 * and some providers treat one as a meaningful (and confusing) instruction.
 */
export function customInstructionsMessage(
  instructions: string = loadCustomInstructions(),
): SystemMessage | null {
  const content = instructions.trim().slice(0, CUSTOM_INSTRUCTIONS_MAX_LENGTH);
  if (!content) return null;
  return { role: 'system', content };
}

/** Prepends the user's instructions to an outgoing request, if there are any. */
export function withCustomInstructions<T>(
  messages: T[],
  instructions?: string,
): (T | SystemMessage)[] {
  const system = customInstructionsMessage(
    instructions ?? loadCustomInstructions(),
  );
  return system ? [system, ...messages] : messages;
}
