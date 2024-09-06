import type { ButtonHTMLAttributes } from "react";

export const Button: React.FC<ButtonHTMLAttributes<HTMLButtonElement>> = (props) => {
  const { className, ...rest } = props;
  return (
    <button type="button" className={["button", className].join(" ")} {...rest}>
      {props.children}
    </button>
  );
};

export const ButtonPrimary: React.FC<ButtonHTMLAttributes<HTMLButtonElement>> = (props) => {
  const { className, ...rest } = props;
  return (
    <button type="button" className={["button button-primary", props.className].join(" ")} {...rest}>
      {props.children}
    </button>
  );
};

export const ButtonSecondary: React.FC<ButtonHTMLAttributes<HTMLButtonElement>> = (props) => {
  const { className, ...rest } = props;
  return (
    <button type="button" className={["button button-secondary", props.className].join(" ")} {...rest}>
      {props.children}
    </button>
  );
};
