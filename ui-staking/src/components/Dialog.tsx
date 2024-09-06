import "@reach/dialog/styles.css";
import "./Dialog.css";

import { type DialogProps, Dialog as ReachDialog } from "@reach/dialog";

export const Dialog = (props: DialogProps) => {
	return <ReachDialog {...props} />;
};
