import type { PropsWithChildren } from "react";
import { PieChart } from "react-minimal-pie-chart";

type Props = PropsWithChildren<{
	progress: number;
}>;

export const Chart = ({ progress, children }: Props) => {
	if (progress < 0 || progress > 1) {
		throw new Error("Progress must be between 0 and 1");
	}
	const data = [
		{ value: progress, color: "#36C6D9" },
		{ value: 1 - progress, color: "#fff" },
	];

	return (
		<div className="chart">
			<PieChart
				data={data}
				totalValue={1}
				lineWidth={27}
				rounded={true}
				startAngle={-125}
			/>
			<div className="chart-text">{children}</div>
		</div>
	);
};
