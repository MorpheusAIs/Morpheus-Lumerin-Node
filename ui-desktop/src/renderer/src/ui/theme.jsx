const darkShade = 'rgba(0, 0, 0, 0.2)';

export default {
  colors: {
    transparent: 'transparent',
    primary: '#0e4353',
    primaryLight: '#53b1bd',
    primaryDark: '#252b34',
    translucentPrimary: '#014353',
    inactive: '#384764',
    active: '#5ADCE2',
    cancelled: '#8B8B96',
    // translucentPrimary: 'rgba(126, 97, 248, 0.2)',
    secondary: '#014353',
    tertiary: '#DB2642',
    light: '#fff',
    copy: '#545454',
    dark: '#fff',
    darker: '#1d1d1d',
    translucentDark: '#323232EE',
    lightShade: 'rgba(0, 0, 0, 0.1)',
    darkShade,
    darkSuccess: 'rgba(119, 132, 125, 0.68)',
    success: '#399E5A',
    warning: '#FFC857',
    danger: '#d46045',
    darkDanger: 'rgba(212, 96, 69, 0.12)',
    weak: '#888',
    morMain: '#20dc8e',
    // BACKGROUNDS
    medium: '#f4f4f4',
    lightBlue: '#EAF7FC',
    lightBG: '#ededed',
    xLight: '#f7f7f7',
    darkGradient: 'linear-gradient(to bottom, #353535, #323232)',

    lumerin: {
      gray: '#F2F5F9',
      darkGray: '#DEE3EA',
      inputGray: '#EDEEF2',
      placeholderGray: '#C4C4C4',
      helpertextGray: '#707070',
      aqua: '#11B4BF',
      tableBorderGray: '#E5E7EB',
      lightAqua: '#DBECED',
      green: '#66BE26'
    }
  },

  // SIZES
  sizes: {
    xSmall: 11,
    small: 14,
    medium: 16,
    large: 20,
    xLarge: 24,
    xxLarge: 32
  },

  // FONT WEIGHTS
  weights: {
    xLight: { fileName: 'Muli-ExtraLight', value: '200' },
    light: { fileName: 'Muli-Light', value: '300' },
    regular: { fileName: 'Muli-Regular', value: '400' },
    semibold: { fileName: 'Muli-SemiBold', value: '600' },
    bold: { fileName: 'Muli-Bold', value: '700' },
    xBold: { fileName: 'Muli-ExtraBold', value: '800' },
    black: { fileName: 'Muli-Black', value: '900' }
  },

  textShadow: `0 1px 1px ${darkShade}`,
  spacing: n => n * 8 // used as rem multiplier
};
