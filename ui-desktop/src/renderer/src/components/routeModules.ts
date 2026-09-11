import type { ComponentType } from 'react';

type RouteModule = { default: ComponentType<any> };

export type AppRoutePath =
  | '/wallet'
  | '/chat'
  | '/workspace'
  | '/agents'
  | '/models'
  | '/providers'
  | '/settings';

export const routeModules: Record<AppRoutePath, () => Promise<RouteModule>> = {
  '/wallet': () => import('./dashboard/Dashboard'),
  '/chat': () => import('./chat/Chat'),
  '/workspace': () => import('./cowork/Cowork'),
  '/agents': () => import('./agents/Agents'),
  '/models': () => import('./models/Models'),
  '/providers': () => import('./providers/Providers'),
  '/settings': () => import('./settings/Settings'),
};

export const createRoutePreloader = <Path extends string>(
  loaders: Record<Path, () => Promise<unknown>>,
) => {
  const pending = new Map<Path, Promise<unknown>>();
  return (path: Path): Promise<unknown> => {
    const existing = pending.get(path);
    if (existing) return existing;
    const request = loaders[path]().catch((error) => {
      // A transient chunk read must remain retryable on the next intent.
      pending.delete(path);
      throw error;
    });
    pending.set(path, request);
    return request;
  };
};

export const preloadRoute = createRoutePreloader(routeModules);
