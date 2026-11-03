import type { ReactNode } from "react";
import { RadioGroup, RadioItem } from "./radio-group";
import { cn } from "@/lib/utils";

interface Option<T extends string> {
  value: T;
  label: ReactNode | ((selected: boolean, value: T) => ReactNode);
  icon?: ReactNode | ((selected: boolean, value: T) => ReactNode);
}

export interface SegmentedProps<T extends string> {
  value: T;
  onChange: (value: T) => void;
  options: readonly Option<T>[];
  label: string;
  className?: string;
}

export function Segmented<T extends string>({
  value,
  onChange,
  options,
  label,
  className,
}: SegmentedProps<T>) {
  return (
    <RadioGroup
      value={value}
      onChange={onChange}
      label={label}
      className={cn("flex w-full gap-0.5", className)}
    >
      {options.map((option) => {
        const selected = option.value === value;
        return (
          <RadioItem
            key={option.value}
            value={option.value}
            className="inline-flex min-w-0 flex-1 items-center justify-center gap-2 rounded-[20px] border-0 bg-primary-container px-4 py-3 text-sm font-medium text-on-primary-container outline-none morph first:rounded-l-[100px] last:rounded-r-[100px] hover:state-layer-8 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-on-surface data-checked:rounded-[100px] data-checked:bg-primary data-checked:text-on-primary [&_svg]:size-5"
          >
            {typeof option.icon === "function" ? option.icon(selected, value) : option.icon}
            {typeof option.label === "function" ? option.label(selected, value) : option.label}
          </RadioItem>
        );
      })}
    </RadioGroup>
  );
}
