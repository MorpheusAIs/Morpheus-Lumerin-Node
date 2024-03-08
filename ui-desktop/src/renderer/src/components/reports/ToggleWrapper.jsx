import React from 'react';
import StyledToggle from './Toggle';

export const Toggle = ({ name, active, toggle }) => {
  return (
    <>
      <StyledToggle
        name={name}
        checked={active}
        labelRight={name}
        onChange={toggle}
        backgroundColorUnchecked="#dddddd"
        backgroundColorChecked="#11B4BF"
      />
      {/* <Label for={name}>
        <Input name={name} checked={active} onChange={toggle} type="checkbox" id={name} />
        <Switch />
      </Label> */}
    </>
  );
};
