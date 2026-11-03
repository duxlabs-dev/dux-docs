import type { ComponentProps } from "react";
import { useRipple } from "@/hooks/ui/use-ripple";
import { cn } from "@/lib/utils";

export type ButtonVariant = "filled" | "tonal" | "text" | "danger";

interface ButtonProps extends ComponentProps<"button"> {
  variant?: ButtonVariant;
}

export const buttonVariants: Record<ButtonVariant, string> = {
  filled: "bg-primary text-on-primary hover:state-layer-8 active:state-layer-12",
  tonal:
    "bg-secondary-container text-on-secondary-container hover:state-layer-8 active:state-layer-12",
  text: "bg-transparent text-on-surface hover:state-layer-8 active:state-layer-12",
  danger:
    "bg-[color-mix(in_srgb,var(--color-error)_22%,var(--color-surface-container))] text-[color-mix(in_srgb,var(--color-error)_55%,var(--color-on-surface))] hover:state-layer-8 active:state-layer-12",
};

export function Button({ variant = "tonal", className, onClick, ref, ...rest }: ButtonProps) {
  const ripple = useRipple({ ref });
  return (
    <button
      type="button"
      ref={ripple}
      onClick={(event) => {
        event.currentTarget.blur();
        onClick?.(event);
      }}
      className={cn(
        "inline-flex cursor-pointer items-center justify-center gap-1.5 rounded-[20px] border-0 px-2.5 py-3 text-[13px] leading-none font-medium focus-ring morph active:rounded-xl [&_svg]:size-5 [&_svg]:shrink-0",
        buttonVariants[variant],
        className,
      )}
      {...rest}
    />
  );
}
