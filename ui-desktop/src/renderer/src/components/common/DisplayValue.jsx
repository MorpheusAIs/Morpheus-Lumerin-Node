import withDisplayValueState from '../../store/hocs/withDisplayValueState';
import { sanitize } from '../../store/utils';
import React from 'react';

export function DisplayValue(props) {
  let formattedValue;
  try {
    formattedValue = this.round(
      props.toWei ? sanitize(props.value) : props.fromWei(props.value),
      props.shouldFormat
    );
  } catch (e) {
    formattedValue = null;
  }

  function numberWithCommas(x) {
    return x.toString().replace(/\B(?<!\.\d*)(?=(\d{3})+(?!\d))/g, ',');
  }

  return (
    <>
      {numberWithCommas(props.value)} {props.post}
    </>
  );
}

export default withDisplayValueState(DisplayValue);
