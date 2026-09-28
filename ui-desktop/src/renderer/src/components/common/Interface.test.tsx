import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../ui/theme';
import TextInput from './TextInput';
import PrimaryNav from '../sidebar/PrimaryNav';
import { Btn } from './Btn';

function themed(children: React.ReactNode) {
  // styled-components v4's legacy component typing predates React 18.
  const Provider = ThemeProvider as any;
  return render(<Provider theme={theme}>{children}</Provider>);
}

describe('shared interface accessibility', () => {
  it('exposes the actual active wallet route and warms pages on keyboard focus', () => {
    const warm = vi.fn(async () => undefined);
    themed(
      <MemoryRouter initialEntries={['/wallet']}>
        <PrimaryNav onRouteIntent={warm} />
      </MemoryRouter>,
    );
    const wallet = screen.getByRole('link', { name: 'Wallet' });
    expect(wallet).toHaveAttribute('aria-current', 'page');
    expect(wallet).toHaveClass('active');
    const workspace = screen.getByRole('link', { name: 'Workspace' });
    fireEvent.focus(workspace);
    expect(warm).toHaveBeenCalledWith('/workspace');
    expect(workspace).not.toHaveAttribute('aria-current');
    expect(screen.getAllByRole('link')).toHaveLength(6);
  });

  it('associates field errors with the input and keeps test identifiers', () => {
    const change = vi.fn();
    themed(
      <TextInput
        id="password"
        data-testid="pass-field"
        label="Password"
        type="password"
        value=""
        error="Enter your password"
        onChange={change}
      />,
    );
    const input = screen.getByLabelText('Password');
    expect(screen.getByTestId('pass-field')).toBe(input);
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(input).toHaveAccessibleDescription('Enter your password');
    expect(screen.getByRole('alert')).toHaveTextContent('Enter your password');
    fireEvent.change(input, { target: { value: 'example' } });
    expect(change).toHaveBeenCalledWith({ id: 'password', value: 'example' });
  });

  it('keeps native button and submit semantics while aligning icon labels', () => {
    const action = vi.fn();
    themed(
      <>
        <Btn onClick={action}>
          <svg aria-hidden="true" />
          Continue in Workspace
        </Btn>
        <Btn submit>Save</Btn>
      </>,
    );
    const button = screen.getByRole('button', {
      name: 'Continue in Workspace',
    });
    expect(button).toHaveAttribute('type', 'button');
    fireEvent.click(button);
    expect(action).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('button', { name: 'Save' })).toHaveAttribute(
      'type',
      'submit',
    );
  });
});
