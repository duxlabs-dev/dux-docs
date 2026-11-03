import {
  createContext,
  useContext,
  type ButtonHTMLAttributes,
  type HTMLAttributes,
  type KeyboardEvent,
} from "react";
import { useRipple } from "@/hooks/ui/use-ripple";
import { cn } from "@/lib/utils";

interface RadioContextValue<T extends string = string> {
  value: T;
  onChange: (value: T) => void;
}

const RadioContext = createContext<RadioContextValue | null>(null);

interface RadioGroupProps<T extends string> extends Omit<
  HTMLAttributes<HTMLDivElement>,
  "onChange"
> {
  value: T;
  onChange: (value: T) => void;
  label: string;
}

const keyDelta: Record<string, number> = {
  ArrowRight: 1,
  ArrowDown: 1,
  ArrowLeft: -1,
  ArrowUp: -1,
};

const moveFocus = (event: KeyboardEvent<HTMLDivElement>) => {
  const items = [...event.currentTarget.querySelectorAll<HTMLButtonElement>('[role="radio"]')];
  const focusedIndex = items.indexOf(
    (event.target as HTMLElement).closest('[role="radio"]') as HTMLButtonElement,
  );
  if (focusedIndex < 0) return;
  let nextIndex: number | undefined;
  if (event.key in keyDelta)
    nextIndex = (focusedIndex + keyDelta[event.key] + items.length) % items.length;
  else if (event.key === "Home") nextIndex = 0;
  else if (event.key === "End") nextIndex = items.length - 1;
  if (nextIndex === undefined) return;
  event.preventDefault();
  event.stopPropagation();
  items[nextIndex].focus();
  items[nextIndex].click();
};

export function RadioGroup<T extends string>({
  value,
  onChange,
  label,
  children,
  ...rest
}: RadioGroupProps<T>) {
  return (
    <RadioContext.Provider value={{ value, onChange: onChange as (value: string) => void }}>
      <div role="radiogroup" aria-label={label} onKeyDown={moveFocus} {...rest}>
        {children}
      </div>
    </RadioContext.Provider>
  );
}

interface RadioItemProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  value: string;
}

export function RadioItem({ value, className, onClick, ...rest }: RadioItemProps) {
  const ripple = useRipple();
  const ctx = useContext(RadioContext);
  if (!ctx) throw new Error("RadioItem must be inside RadioGroup");
  const checked = ctx.value === value;
  return (
    <button
      type="button"
      ref={ripple}
      role="radio"
      aria-checked={checked}
      tabIndex={checked ? 0 : -1}
      data-checked={checked || undefined}
      onClick={(event) => {
        ctx.onChange(value);
        onClick?.(event);
      }}
      className={cn("cursor-pointer", className)}
      {...rest}
    />
  );
}
