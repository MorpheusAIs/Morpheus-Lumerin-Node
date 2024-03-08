export const RANGE = {
  LOCAL: 'local',
  SUBNET_16: '192.168.1.1/16',
  SUBNET_24: '192.168.1.1/24',
  CUSTOM: 'custom'
};

export const rangeSelectOptions = [
  {
    label: 'Local network',
    value: RANGE.LOCAL
  },
  {
    label: '192.168.1.1/24',
    value: RANGE.SUBNET_24
  },
  {
    label: '192.168.1.1/16',
    value: RANGE.SUBNET_16
  },
  {
    label: 'Custom range',
    value: RANGE.CUSTOM
  }
];

export const mapRangeNameToIpRange = range => {
  switch (range) {
    case RANGE.SUBNET_16:
      return ['192.168.1.1', '192.168.255.255'];
    case RANGE.SUBNET_24:
      return ['192.168.1.1', '192.168.1.255'];
    default:
      throw new Error(`Unknown range ${range}`);
  }
};
