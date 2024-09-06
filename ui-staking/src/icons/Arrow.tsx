interface Props extends React.SVGProps<SVGSVGElement> {
  angle?: number;
  fill?: string;
}

export const Arrow: React.FC<Props> = (props) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 20 13"
    fill={props.fill || "#fff"}
    // style={{ transform: `rotate(${props.angle || 0}deg)` }}
    {...props}
  >
    <title>Arrow</title>
    <path d="m10 0 10 10-2.35 2.35L10 4.717 2.35 12.35 0 10 10 0Z" />
  </svg>
);
