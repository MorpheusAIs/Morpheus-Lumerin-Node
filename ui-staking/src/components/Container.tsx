import type { FC, HTMLAttributes } from "react";

export const Container: FC<HTMLAttributes<HTMLDivElement>> = (props) => {
  return (
    <div {...props} className={["container", props.className].join(" ")}>
      {props.children}
    </div>
  );
};

export const ContainerNarrow: FC<HTMLAttributes<HTMLDivElement>> = (props) => {
  return (
    <div {...props} className={["container-narrow", props.className].join(" ")}>
      {props.children}
    </div>
  );
};
