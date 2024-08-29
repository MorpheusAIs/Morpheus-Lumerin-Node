import type { PropsWithChildren } from "react";
import { PieChart } from "react-minimal-pie-chart";

type Props = PropsWithChildren<{
	progress: number;
	lineWidth?: number;
	className?: string;
}>;

export const Chart = (props: Props) => {
	const { progress, children, lineWidth = 27, className } = props;
	if (progress < 0 || progress > 1) {
		throw new Error(`Progress must be between 0 and 1: ${progress}`);
	}

	const data = [
		{ value: progress, color: "#000" },
		{ value: 1 - progress, color: "#fff" },
	];

	return (
		<div className={["chart", className].join(" ")}>
			<PieChart
				segmentsStyle={(index) => {
					return {
						stroke: index === 0 ? "url(#cl1)" : "#404040",
					};
				}}
				data={data}
				totalValue={1}
				lineWidth={lineWidth}
				rounded={false}
				startAngle={-90}
			/>
			<div className="chart-text">{children}</div>
		</div>
	);
};
