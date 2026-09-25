import { describe, expect, it } from 'vitest';
import { groupProviderSessions } from './ProvidersList';

describe('groupProviderSessions', () => {
  it('renders only models with sessions and uses cached names case-insensitively', () => {
    const groups = groupProviderSessions(
      [
        { Id: 'session-1', ModelAgentId: '0xABC' },
        { Id: 'session-2', ModelAgentId: '0xabc' },
      ],
      { '0xabc': 'Model ABC', unused: 'No sessions' },
    );

    expect(groups).toHaveLength(1);
    expect(groups[0].name).toBe('Model ABC');
    expect(groups[0].sessions).toHaveLength(2);
  });
});
