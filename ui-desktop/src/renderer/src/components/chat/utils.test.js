import { describe, expect, it } from 'vitest';
import { isClosed } from './utils';

const nowSeconds = () => Math.floor(Date.now() / 1000);

// `isClosed` decides whether the Chat tab shows "you have an open session" or
// the "Select payment method" screen. Getting it wrong in the permissive
// direction is what let users stake a second time on top of a live session.
describe('isClosed', () => {
  it('treats a session with ClosedAt as closed', () => {
    expect(isClosed({ ClosedAt: 1700000000, EndsAt: nowSeconds() + 3600 })).toBeTruthy();
  });

  it('treats an expired session as closed', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() - 1 })).toBeTruthy();
  });

  it('treats a live session as open', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() + 3600 })).toBeFalsy();
  });

  it('does not mark a session closed merely because it ends soon', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() + 5 })).toBeFalsy();
  });
});
