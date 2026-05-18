// Re-export shim. The renderer imports `ApiGateway` from this path historically.
// The actual client used at runtime is constructed in the renderer
// (`renderer/src/client/index.ts`) and exposes IPC-forwarded methods plus extras
// that the main-process handler type doesn't capture. Use that broader type here
// so renderer code sees the right method names (e.g. `getFailoverSetting`).
//
// `any` is intentional: the renderer client is built dynamically and there's no
// single canonical type that captures all of its surface. Keeping this as `any`
// avoids forcing every callsite to wrestle with `Client` vs the renderer wrapper.
export type ApiGateway = any

