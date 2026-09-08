# Isolated UI preview

Run `node scripts/ui-preview/server.mjs` from `ui-desktop`, then open
<http://127.0.0.1:5188/>. Only the local loopback interface is exposed.

This renders the actual Workspace, Sidebar, QuickStartGuide, receive form,
common controls, and model picker with synthetic fixtures. It does **not** load
the app bootstrap, `.env`, wallet, proxy-router, Electron preload, saved chats,
or project files. All action handlers are in-memory mocks, including copying.
Application fetch/XHR and external navigation are disabled. The zero address,
sample model names, token amounts, task history, and files are test fixtures,
not live network information. Never enter secrets in this preview.

Use the toolbar for compact/wide layout, a saved/empty project, and replaying
the guide. Useful direct views:

- `/?compact=1#/workspace` — compact route container and drawer navigation
- `/?empty=1#/workspace` — first-project setup
- `/?modal=receive#/wallet` — long address and receive dialog
- `/?modal=models#/chat` — model filters and rows

Browser window resizing additionally exercises the real sidebar media queries.
The preview server uses its own Vite configuration and never loads the normal
Electron build configuration. Stop it with Ctrl+C.

## Session confirmation (Electron)

After `npm run build`, keep the preview server above running and launch
`node scripts/ui-preview/session-confirmation-smoke.mjs` in another terminal.
This opens the real, main-owned confirmation with its dedicated isolated preload,
not the normal wallet app. It uses a fresh temporary user-data directory.

With the parent preview focused, press **R** for a stake confirmation, **D** for
direct payment, **N** for a smaller window or **W** for a wider window. Confirm
and cancel using the actual buttons or keyboard; results and listener cleanup
are printed to the terminal. There is no transaction submission path in this
harness. Escape cancels, Cancel receives initial focus, and Tab stays inside
the confirmation. Stop both processes with Ctrl+C when finished.
