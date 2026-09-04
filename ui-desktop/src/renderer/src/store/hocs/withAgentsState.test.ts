import { describe, expect, it, vi } from 'vitest';
import { loadAgentsPageData } from './withAgentsState';

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};

describe('loadAgentsPageData', () => {
  it('starts independent reads together and categorizes the cached result', async () => {
    const users = deferred<any>();
    const allowances = deferred<any>();
    const client = {
      getAgentUsers: vi.fn(() => users.promise),
      getAgentAllowanceRequests: vi.fn(() => allowances.promise),
    };

    const result = loadAgentsPageData(client);

    expect(client.getAgentUsers).toHaveBeenCalledOnce();
    expect(client.getAgentAllowanceRequests).toHaveBeenCalledOnce();
    users.resolve({
      agents: [
        { username: 'pending', isConfirmed: false },
        { username: 'active', isConfirmed: true },
      ],
    });
    allowances.resolve({ requests: [{ username: 'pending', token: 'MOR' }] });

    await expect(result).resolves.toMatchObject({
      pendingAgents: [{ username: 'pending' }],
      activeAgents: [{ username: 'active' }],
      allowanceRequests: [{ username: 'pending', token: 'MOR' }],
    });
  });
});
