import moment from 'moment';
import { lmrDecimals } from '../../utils/coinValue';
import { CONTRACT_STATE } from '../../enums';

const MICRO = 10 ** 6;

export const toMicro = value => {
  return value * MICRO;
};

export const fromMicro = value => {
  return value / MICRO;
};

const getReadableDate = (days, hours, minutes, seconds) => {
  const readableDays = days
    ? days === 1
      ? `${days} day`
      : `${days} days`
    : '';
  const readableHours = hours
    ? hours === 1
      ? `${hours} hour`
      : `${hours} hours`
    : '';
  const readableMinutes = minutes
    ? minutes === 1
      ? `${minutes} minute`
      : `${minutes} minutes`
    : '';
  const readableSeconds =
    !days && !hours && !minutes && seconds
      ? seconds === 1
        ? `1 second`
        : `${seconds} seconds`
      : '';
  const readableDate = `${readableDays} ${readableHours} ${readableMinutes} ${readableSeconds}`.trim();
  return readableDate;
};

export const formatSpeed = speed => {
  return `${Number(speed) / 10 ** 12} TH/s`;
};

export const formatTimestamp = (timestamp, timer, state) => {
  if (+timestamp === 0) {
    return '';
  }
  if (state !== CONTRACT_STATE.Running) {
    return '';
  }
  const startDate = moment.unix(timestamp).format('L');
  const { days, hours, minutes, seconds } = timer;
  if (days || hours || minutes || seconds) {
    const durationLeft = getReadableDate(days, hours, minutes, seconds);
    return `${startDate} (${durationLeft} left)`;
  } else {
    return `${startDate} (completed)`;
  }
};

export const formatPrice = (price, symbol) => {
  const value = Number(price) / lmrDecimals;
  if (Number.isNaN(value)) {
    return `0 ${symbol}`;
  }
  return `${Math.round(value * 100) / 100} ${symbol}`;
};

export const formatBtcPerTh = networkDifficulty => {
  const networkHashrate = (networkDifficulty * Math.pow(2, 32)) / 600;
  const profitPerTh = (1000000000000 / networkHashrate) * (8 * 144);

  const fixedValue = Number(profitPerTh).toFixed(8);
  return fixedValue;
};

export const formatDuration = duration => {
  const numLength = parseFloat(duration / 3600);
  const days = Math.floor(numLength / 24);
  const remainder = numLength % 24;
  const hours = days >= 1 ? Math.floor(remainder) : Math.floor(numLength);
  const minutes =
    days >= 1
      ? Math.floor(60 * (remainder - hours))
      : Math.floor((numLength - Math.floor(numLength)) * 60);
  const seconds = Math.floor(duration % 60);
  const readableDate = getReadableDate(days, hours, minutes, seconds);
  return readableDate;
};

export const isContractClosed = contract => {
  return contract.seller === contract.buyer;
};

export const getContractState = contract => {
  const state = Object.entries(CONTRACT_STATE).find(
    s => contract.state === s[1]
  )[1];
  if (contract.state === CONTRACT_STATE.Avaliable) {
    return 'Available';
  }

  const startDate = moment.unix(contract.timestamp).format('lll');
  return `Running. Started at ${startDate}`;
};

export const getContractEndTimestamp = contract => {
  if (+contract.timestamp === 0) {
    return 0;
  }
  return (+contract.timestamp + +contract.length) * 1000; // in ms
};

export const truncateAddress = (address, desiredLength) => {
  let index;
  switch (desiredLength) {
    case 'SHORT':
      return `${address.substring(0, 5)}...`;
    case 'MEDIUM':
      index = 5;
      break;
    case 'LONG':
      index = 10;
      break;
    default:
      index = 10;
  }
  return `${address.substring(0, index)}...${address.substring(
    address.length - index,
    address.length
  )}`;
};

export const getContractRewardBtcPerTh = (contract, btcRate, lmrRate) => {
  const lengthDays = contract.length / 60 / 60 / 24;
  const speed = Number(contract.speed) / 10 ** 12;

  const contractBtcPrice = convertLmrToBtc(contract.price, btcRate, lmrRate);
  const result = contractBtcPrice / speed / lengthDays;
  return result;
};

export const convertLmrToBtc = (value, btcRate, lmrRate) => {
  const contractUsdPrice = (value / lmrDecimals) * lmrRate;
  return contractUsdPrice / btcRate;
};

export const formatExpNumber = value =>
  value.toFixed(10).replace(/(?<=\.\d*[1-9])0+$|\.0*$/, '');

export const calculateSuggestedPrice = (
  time,
  speed,
  btcRate,
  lmrRate,
  profit,
  multiplier
) => {
  const lengthDays = time / 24;
  return (
    (multiplier * profit * lengthDays * speed * btcRate) /
    lmrRate
  ).toFixed(0);
};
