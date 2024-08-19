import type { FC, HTMLAttributes } from "react";

export const Container: FC<HTMLAttributes<HTMLDivElement>> = (props) => {
	return (
		<div {...props} className={["container", props.className].join(" ")}>
			{props.children}
		</div>
	);
};
