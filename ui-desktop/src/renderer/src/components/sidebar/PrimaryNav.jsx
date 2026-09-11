import React from 'react';
import {
  IconMessage,
  IconSparkles,
  IconBrandStackshare,
  IconWallet,
  IconPackages,
  IconUsers,
} from '@tabler/icons-react';
import { NavGroup, NavItem } from './Nav.styles';

const pages = [
  { path: '/wallet', label: 'Wallet', icon: IconWallet },
  { path: '/chat', label: 'Chat', icon: IconMessage },
  { path: '/workspace', label: 'Workspace', icon: IconSparkles },
  { path: '/models', label: 'Models', icon: IconPackages },
  { path: '/agents', label: 'Agents', icon: IconUsers },
  { path: '/providers', label: 'Provider Hub', icon: IconBrandStackshare },
];

export default function PrimaryNav({ onRouteIntent }) {
  const warm = (path) => {
    void onRouteIntent?.(path)?.catch(() => undefined);
  };
  return (
    <NavGroup>
      {pages.map(({ path, label, icon: Icon }) => (
        <NavItem
          key={path}
          to={path}
          aria-label={label}
          title={label}
          data-guide={path.slice(1)}
          data-testid={`${path.slice(1)}-nav-btn`}
          onFocus={() => warm(path)}
          onPointerEnter={() => warm(path)}
        >
          <Icon aria-hidden="true" stroke={1.7} />
          <span>{label}</span>
        </NavItem>
      ))}
    </NavGroup>
  );
}
