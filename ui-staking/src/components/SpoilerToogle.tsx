import { Arrow } from "../icons/Arrow.tsx";

export const SpoilerToogle = () => {
	return (
		<>
			<label className="toogle">
				<input type="checkbox" className="toogle-checkbox" />
				<Arrow fill="#fff" className="toogle-arrow" />
			</label>
		</>
	);
};
