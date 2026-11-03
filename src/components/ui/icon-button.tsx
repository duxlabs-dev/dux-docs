import type { ComponentProps } from "react";
import { useRipple } from "@/hooks/ui/use-ripple";
import { cn } from "@/lib/utils";
import { type ButtonVariant, buttonVariants } from "./button";

interface IconButtonProps extends ComponentProps<"button"> {
  label: string;
  variant?: ButtonVariant;
  size?: "sm" | "md";
}

export function IconButton({
  label,
  variant = "tonal",
  size = "md",
  className,
  onClick,
  ref,
  ...rest
}: IconButtonProps) {
  const ripple = useRipple({ ref });
  return (
    <button
      type="button"
      ref={ripple}
      title={rest.title ?? label}
      aria-label={label}
      onClick={(event) => {
        event.currentTarget.blur();
        onClick?.(event);
      }}
      className={cn(
        "inline-flex shrink-0 cursor-pointer items-center justify-center border-0 p-0 focus-ring morph active:rounded-lg disabled:pointer-events-none disabled:opacity-40 [&_svg]:size-6",
        size === "sm" ? "size-8 rounded-[50%] [&_svg]:size-5" : "size-10 rounded-[50%]",
        buttonVariants[variant],
        className,
      )}
      {...rest}
    />
  );
}
