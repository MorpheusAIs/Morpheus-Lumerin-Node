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
