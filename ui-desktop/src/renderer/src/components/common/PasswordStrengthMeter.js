import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { GetPasswordStrength, MaxScore } from '../../lib/PasswordStrength';
import { BarElem, Container, Message } from './PasswordStrengthMeter.styles';

/**
 * @param {number} score
 * @returns string
 */
const mapScoreToMessage = score => {
  switch (score) {
    case 0:
      return 'Too weak';
    case 1:
      return 'Very weak';
    case 2:
      return 'Almost there';
    case 3:
      return 'Strong';
    case 4:
      return 'Very strong';
    default:
      return '';
  }
};

/**
 * Returns an interpolated CSS hue value between red & green
 * based on passwordEntropy / targetEntropy ratio
 * @param {number} ratio passwordEntropy / targetEntropy ratio
 * @returns {number} interpolated CSS hue value between red & green
 */
function getHue(ratio) {
  // Hues are adapted to match the theme's success and danger colors
  const orangeHue = 50;
  const greenHue = 139;
  const redHue = 11;

  return ratio < 1 ? ratio * orangeHue + redHue : greenHue;
}

/**
 * @component
 * @param {Object} param
 * @param {string} param.password
 * @param {(result: import("../../lib/PasswordStrength").ScoreResult)=>void} param.onChange
 */
const PasswordStrengthMeter = ({ password, onChange }) => {
  const [score, setScore] = useState(0);

  useEffect(() => {
    const res = GetPasswordStrength(password);
    setScore(res.score);
    onChange(res);
  }, [password]);

  const hue = getHue(score / MaxScore);
  // adding 1 so if the score is 0 the bar will not be zero width
  const barWidthFraction = (score + 1) / (MaxScore + 1);

  let colorCSS = `hsl(${hue}, 62%, 55%)`;
  let barWidthCSS = `${barWidthFraction * 100}%`;

  let message = mapScoreToMessage(score);

  if (!password) {
    barWidthCSS = '0%';
    message = '';
  }

  return (
    <Container>
      <BarElem width={barWidthCSS} color={colorCSS} />
      <Message color={colorCSS}>{message}</Message>
    </Container>
  );
};

PasswordStrengthMeter.propTypes = {
  password: PropTypes.string
};

export default PasswordStrengthMeter;
