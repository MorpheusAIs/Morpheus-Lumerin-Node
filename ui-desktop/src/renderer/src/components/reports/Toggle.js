import React from 'react';
import { GlobalContainer, InputContainer, Label, Slider, Input } from './Style';

const Toggle = ({
  onChange,
  checked,
  disabled,
  width = 60,
  height = 34,
  translate = 26,
  backgroundColorChecked,
  backgroundColorUnchecked,
  backgroundColorButton = '#fff',
  name,
  value,
  labelRight = '',
  labelLeft = '',
  sliderWidth = 26,
  sliderHeight = 26,
  labelColor = '#fff'
}) => {
  return (
    <GlobalContainer>
      {labelLeft && <Label color={labelColor}>{labelLeft}</Label>}
      <InputContainer width={width} height={height}>
        <Input
          type="checkbox"
          name={name}
          onChange={onChange}
          value={value}
          checked={checked}
          disabled={disabled}
          translate={translate}
          backgroundColorUnchecked={backgroundColorUnchecked}
          backgroundColorChecked={backgroundColorChecked}
        />
        <Slider
          sliderWidth={sliderWidth}
          sliderHeight={sliderHeight}
          backgroundColorUnchecked={backgroundColorUnchecked}
          backgroundColorButton={backgroundColorButton}
        />
      </InputContainer>
      {labelRight && <Label>{labelRight}</Label>}
    </GlobalContainer>
  );
};

export default Toggle;
