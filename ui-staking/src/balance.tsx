import React from "react";
import { formatLMR, formatMOR } from "./lib/units.ts";

export const BalanceLMR = (props: { value: bigint }) => {
	return (
		<span title={`${Number(props.value) / 1e8} LMR`}>
			≈ {formatLMR(props.value)}
		</span>
	);
};

export const BalanceMOR = (props: { value: bigint }) => {
	return (
		<span title={`${Number(props.value) / 1e18} LMR`}>
			≈ {formatMOR(props.value)}
		</span>
	);
};
