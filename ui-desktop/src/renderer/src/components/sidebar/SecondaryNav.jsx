import React from 'react';
import { IconSettings, IconHelp, IconRoute } from '@tabler/icons-react';
import { withClient } from '../../store/hocs/clientContext';
import { NavAction, NavGroup, NavItem } from './Nav.styles';
import { openQuickStartGuide } from '../onboarding/QuickStartGuide';

function SecondaryNav({ client: { onHelpLinkClick }, onRouteIntent }) {
  const warm = () => {
    void onRouteIntent?.('/settings')?.catch(() => undefined);
  };
  return (
    <NavGroup>
      <NavItem
        to="/settings"
        data-guide="settings"
        data-testid="tools-nav-btn"
        aria-label="Settings"
        title="Settings"
        onFocus={warm}
        onPointerEnter={warm}
      >
        <IconSettings aria-hidden="true" stroke={1.7} />
        <span>Settings</span>
      </NavItem>
      <NavAction
        onClick={openQuickStartGuide}
        data-guide-launcher
        aria-label="Quick start guide"
        title="Quick start guide"
      >
        <IconRoute aria-hidden="true" stroke={1.7} />
        <span>Quick start</span>
      </NavAction>
      <NavAction
        data-testid="help-nav-btn"
        onClick={onHelpLinkClick}
        aria-label="Help and documentation"
        title="Help and documentation"
      >
        <IconHelp aria-hidden="true" stroke={1.7} />
        <span>Help & docs</span>
      </NavAction>
    </NavGroup>
  );
}

export default withClient(SecondaryNav);
