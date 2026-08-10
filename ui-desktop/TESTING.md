# Testing — ui-desktop

## Running

```bash
yarn install          # yarn.lock is authoritative; do not use npm here
yarn test             # single run
yarn test:watch       # re-run on change
yarn test:coverage    # with a coverage report
```

Runner is [Vitest](https://vitest.dev) with a jsdom environment. Config lives in
`vitest.config.ts`, global setup in `src/test/setup.ts`.

## Use yarn, not npm

`yarn.lock` is authoritative here (the release jobs in `build.yml` install with
`yarn --frozen-lockfile`).

This matters more than it looks: **yarn 1 does not auto-install peer
dependencies, but npm 7+ does.** `@testing-library/react@16` declares
`@testing-library/dom` as a peer, so a suite that passes under npm can fail
under yarn with `Cannot find package '@testing-library/dom'`. It is listed
explicitly in `devDependencies` for exactly this reason — don't remove it just
because nothing imports it directly.

## Why the config is standalone

`vitest.config.ts` does not extend `electron.vite.config.ts` on purpose. That
config validates the full runtime env schema when it loads, which would force
every contributor and every CI run to supply a complete `.env` just to run unit
tests. Tests should work on a fresh clone with nothing but an install.

## What is covered

The suite deliberately targets logic that is pure, load-bearing, and has already
broken in production at least once:

| Area | File | Guards against |
|---|---|---|
| Wei conversion | `store/utils/amount.test.ts` | Float rounding silently changing a transfer amount |
| Address validation | `store/utils/amount.test.ts` | Sending to a malformed or truncated address |
| Bounded concurrency | `store/utils/concurrency.test.ts` | Unbounded fan-out stalling the Providers/Models tabs |
| IPC correlation | `client/utils.test.ts` | One response cancelling another request's timeout, hanging the UI |
| Session/stake helpers | `store/queries.test.ts` | Miscounting staked funds; cache keys diverging between tabs |
| Session open/closed | `components/chat/utils.test.js` | Showing "stake now" while a session is already live |

## Conventions

- Co-locate tests with the code: `foo.ts` → `foo.test.ts`.
- Prefer extracting pure logic into its own module over mounting a component to
  reach it. `toBaseUnits` lives in `store/utils/amount.ts` rather than inside the
  transaction-modal HOC precisely so it can be tested without React, redux and
  the IPC client in the way.
- A regression test should fail if the fix is reverted. If you can't state which
  reintroduced bug your test catches, it probably isn't earning its place.

## A note on the deleted suite

`src/renderer/src/components/__tests__/` was removed. Those ten files were
inherited from the Lumerin fork and had been dead for years: they imported a
`testUtils` harness and a `config` module that no longer exist, targeted deleted
components (`ReceiveDrawer`, `SendDrawer`, `Tools`), and depended on
`react-testing-library` — deprecated in 2019 and not in `package.json`. Nothing
ever ran them, and their presence implied coverage that did not exist.
