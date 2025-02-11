const darkShade = 'rgba(0, 0, 0, 0.2)';

export default {
  colors: {
    transparent: 'transparent',
    primary: 'rgba(23, 54, 41, 1)',
    primaryLight: 'rgba(32, 220, 142, 1)',
    primaryDark: 'rgba(12,31,23,1',
    translucentPrimary: 'rgb(1, 67, 83)',
    inactive: 'rgba(56, 71, 100, 1)',
    active: 'rgba(90, 220, 226, 1)',
    cancelled: 'rgba(139, 139, 150, 1)',
    secondary: 'rgba(1, 67, 83, 1)',
    tertiary: 'rgba(219, 38, 66, 1)',
    light: 'rgba(255, 255, 255, 1)',
    copy: 'rgba(84, 84, 84, 1)',
    dark: 'rgba(255, 255, 255, 1)',
    darker: 'rgba(29, 29, 29, 1)',
    translucentDark: 'rgba(50, 50, 50, 0.93)',
    lightShade: 'rgba(0, 0, 0, 0.1)',
    darkShade,
    darkSuccess: 'rgba(119, 132, 125, 0.68)',
    success: 'rgba(57, 158, 90, 1)',
    warning: 'rgba(255, 200, 87, 1)',
    danger: 'rgba(212, 96, 69, 1)',
    darkDanger: 'rgba(212, 96, 69, 0.12)',
    weak: 'rgba(136, 136, 136, 1)',
    morMain: 'rgba(32, 220, 142, 1)',
    morLight: 'rgba(23, 54, 41, 1)',
    // BACKGROUNDS
    medium: 'rgba(244, 244, 244, 1)',
    lightBlue: 'rgba(234, 247, 252, 1)',
    lightBG: 'rgba(237, 237, 237, 1)',
    xLight: 'rgba(247, 247, 247, 1)',
    darkGradient: 'linear-gradient(to bottom, #353535, #323232)',
    helpertextGray: 'rgba(112, 112, 112, 1)',
    placeholderGray: 'rgba(196, 196, 196, 1)',
  },

  // SIZES
  sizes: {
    xSmall: 11,
    small: 14,
    medium: 16,
    large: 20,
    xLarge: 24,
    xxLarge: 32,
  },

  // FONT WEIGHTS
  weights: {
    xLight: { fileName: 'Muli-ExtraLight', value: '200' },
    light: { fileName: 'Muli-Light', value: '300' },
    regular: { fileName: 'Muli-Regular', value: '400' },
    semibold: { fileName: 'Muli-SemiBold', value: '600' },
    bold: { fileName: 'Muli-Bold', value: '700' },
    xBold: { fileName: 'Muli-ExtraBold', value: '800' },
    black: { fileName: 'Muli-Black', value: '900' },
  },

  textShadow: `0 1px 1px ${darkShade}`,
  spacing: (n) => n * 8, // used as rem multiplier
};
