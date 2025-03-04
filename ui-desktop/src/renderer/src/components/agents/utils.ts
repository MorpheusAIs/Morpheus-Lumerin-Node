import { formatAddress } from '@renderer/lib/address';
import { isAddress } from '@renderer/client/utils';
import {
  fromTokenBaseUnitsToLMR,
  fromTokenBaseUnitsToETH,
} from '@renderer/utils/coinValue';

export function formatTokenNameValue(
  tokenAddress: string,
  value: string,
  props: { symbol: string; symbolEth: string; morTokenAddress: string },
): { name: string; value: string } {
  if (isAddress(tokenAddress)) {
    if (tokenAddress.toLowerCase() === props.morTokenAddress.toLowerCase()) {
      return { name: props.symbol, value: fromTokenBaseUnitsToLMR(value) };
    } else {
      return { name: formatAddress(tokenAddress), value: value + ' units' };
    }
  } else if (tokenAddress.toLowerCase() === 'eth') {
    return { name: props.symbolEth, value: fromTokenBaseUnitsToETH(value) };
  } else {
    return { name: tokenAddress, value: value };
  }
}

export function getAbbreviation(username: string) {
  let [first, second] = username.split(' ');
  if (!second) {
    second = first.slice(1, 2);
  }
  if (!second) {
    second = '';
  }

  return `${first[0]}${second[0]}`.toUpperCase();
}
