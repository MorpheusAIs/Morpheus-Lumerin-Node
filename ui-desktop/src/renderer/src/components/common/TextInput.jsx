import PropTypes from 'prop-types';
import React from 'react';
import { ErrorMsg, Input, Label, TextArea } from './TextInput.styles';

const TextInput = ({
  id,
  type = 'text',
  rows,
  cols,
  label,
  value,
  onChange,
  error,
  disabled,
  ['data-testid']: dataTestId,
  ...rest
}) => {
  const InputControl = rows || cols ? TextArea : Input;
  const hasErrors = error && error.length > 0;

  return (
    <div>
      <Label hasErrors={hasErrors} htmlFor={id}>
        {label}
      </Label>
      <InputControl
        id={id}
        rows={rows}
        cols={cols}
        value={value || ''}
        type={type}
        onChange={e => onChange({ id: id, value: e.target.value })}
        hasErrors={hasErrors}
        disabled={disabled}
        {...rest}
      />
      {hasErrors && (
        <ErrorMsg data-testid={`${dataTestId}-error`}>
          {typeof error === 'string' ? error : error.join('. ')}
        </ErrorMsg>
      )}
    </div>
  );
};

TextInput.propTypes = {
  'data-testid': PropTypes.string,
  placeholder: PropTypes.string,
  autoFocus: PropTypes.bool,
  onChange: PropTypes.func.isRequired,
  error: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.string),
    PropTypes.string
  ]),
  label: PropTypes.string.isRequired,
  value: PropTypes.string,
  type: PropTypes.oneOf(['text', 'number', 'password', 'url']),
  rows: PropTypes.number,
  cols: PropTypes.number,
  id: PropTypes.string.isRequired,
  disabled: PropTypes.bool
};

export default TextInput;
