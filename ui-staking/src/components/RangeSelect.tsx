import { Range } from "react-range";

interface RangeSelect {
  value: number;
  label?: string;
  titles: string[];
  onChange: (value: number) => void;
}

export const RangeSelect = (props: RangeSelect) => {
  return (
    <Range
      label={props.label}
      values={[props.value]}
      min={0}
      max={props.titles.length - 1}
      onChange={(v) => props.onChange(v[0])}
      renderTrack={(p) => (
        <div {...p.props} className="range-track">
          {p.children}
        </div>
      )}
      renderMark={(p) => (
        <div {...p.props} key={p.props.key} className="range-mark">
          <div className="range-mark-label">{props.titles[p.index]}</div>
        </div>
      )}
      renderThumb={(p) => <div {...p.props} className="range-thumb" />}
    />
  );
};
