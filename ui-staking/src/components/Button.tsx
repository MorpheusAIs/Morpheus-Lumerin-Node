import type { ButtonHTMLAttributes } from "react";

export const Button: React.FC<ButtonHTMLAttributes<HTMLButtonElement>> = (
	props,
) => {
	return (
		<button type="button" onClick={props.onClick} className="button">
			{props.children}
		</button>
	);
};
