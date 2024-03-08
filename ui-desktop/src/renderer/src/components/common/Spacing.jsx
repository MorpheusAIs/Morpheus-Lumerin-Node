import PropTypes from 'prop-types';
import styled from 'styled-components';

// eslint-disable-next-line complexity
const getMargin = function({ m, mt, mb, my, ml, mr, mx, theme: { spacing } }) {
  const marginBottom = `${0.1 * spacing(m || mb || my || 0)}rem`;
  const marginRight = `${0.1 * spacing(m || mr || mx || 0)}rem`;
  const marginLeft = `${0.1 * spacing(m || ml || mx || 0)}rem`;
  const marginTop = `${0.1 * spacing(m || mt || my || 0)}rem`;

  return `${marginTop} ${marginRight} ${marginBottom} ${marginLeft}`;
};

// eslint-disable-next-line complexity
const getPadding = function({ p, pt, pb, py, pl, pr, px, theme: { spacing } }) {
  const paddingBottom = `${0.1 * spacing(p || pb || py || 0)}rem`;
  const paddingRight = `${0.1 * spacing(p || pr || px || 0)}rem`;
  const paddingLeft = `${0.1 * spacing(p || pl || px || 0)}rem`;
  const paddingTop = `${0.1 * spacing(p || pt || py || 0)}rem`;

  return `${paddingTop} ${paddingRight} ${paddingBottom} ${paddingLeft}`;
};

const Spacing = styled.div`
  display: ${({ display }) => display || 'block'};
  padding: ${getPadding};
  margin: ${getMargin};
`;

Spacing.propTypes = {
  display: PropTypes.oneOf(['block', 'inline', 'inline-block', 'flex']),
  mt: PropTypes.number, // margin-top only
  mb: PropTypes.number, // margin-bottom only
  ml: PropTypes.number, // margin-left only
  mr: PropTypes.number, // margin-right only
  mx: PropTypes.number, // margin left & right (horizontal)
  my: PropTypes.number, // margin top & bottom (vertical)
  m: PropTypes.number, // all 4 sides,
  pt: PropTypes.number, // padding-top only
  pb: PropTypes.number, // padding-bottom only
  pl: PropTypes.number, // padding-left only
  pr: PropTypes.number, // padding-right only
  px: PropTypes.number, // padding left & right (horizontal)
  py: PropTypes.number, // padding top & bottom (vertical)
  p: PropTypes.number // all 4 sides
};

export default Spacing;
