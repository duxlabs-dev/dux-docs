// see https://github.com/asplunds/use-ripple and LICENSE-ISC-use-ripple in the repo root
import { useCallback, useEffect, useLayoutEffect, useRef, type Ref } from "react";
import { cn } from "@/lib/utils";

export interface RippleOptions<T extends HTMLElement = HTMLElement> {
  duration?: number;
  ref?: Ref<T>;
}

const DEFAULT_DURATION = 600;

const FAILSAFE_FACTOR = 4;

const REMOVE_MARGIN = 50;

const CIRCLE =
  "absolute rounded-full bg-[image:radial-gradient(closest-side,currentColor_92%,transparent)] " +
  "opacity-10 pointer-events-none -translate-x-1/2 -translate-y-1/2 scale-0 " +
  "transition duration-(--ripple-duration) ease-effect";

const applyHostClasses = (host: HTMLElement) => {
  const position =
    !host.classList.contains("relative") && getComputedStyle(host).position === "static"
      ? "relative"
      : "";
  const next = cn(host.className, position, "overflow-hidden");
  if (next !== host.className) host.className = next;
};

const assignRef = <T extends HTMLElement>(external: Ref<T> | null | undefined, node: T | null) => {
  if (typeof external === "function") {
    const cleanup = external(node);
    return typeof cleanup === "function" ? cleanup : () => external(null);
  }
  if (external) {
    external.current = node;
    return () => {
      external.current = null;
    };
  }
  return () => {};
};

const spawnCircle = (
  host: HTMLElement,
  event: PointerEvent,
  duration: number,
  live: Set<() => void>,
) => {
  if (event.button !== 0 || host.matches(":disabled")) return;

  const { left, top, width, height } = host.getBoundingClientRect();
  const offsetX = event.clientX - left;
  const offsetY = event.clientY - top;
  const reach =
    2 * Math.hypot(Math.max(offsetX, width - offsetX), Math.max(offsetY, height - offsetY));

  const circle = document.createElement("span");
  circle.className = CIRCLE;
  circle.setAttribute("aria-hidden", "true");
  circle.style.cssText =
    `left:${offsetX}px; top:${offsetY}px; width:${reach}px; height:${reach}px; ` +
    `--ripple-duration:${duration}ms`;
  host.appendChild(circle);

  let removal: ReturnType<typeof setTimeout> | undefined;
  const stop = () => {
    window.removeEventListener("pointerup", release);
    window.removeEventListener("pointercancel", release);
    clearTimeout(failSafe);
    clearTimeout(removal);
    circle.remove();
    live.delete(stop);
  };
  const release = (releaseEvent?: PointerEvent) => {
    if (releaseEvent && releaseEvent.pointerId !== event.pointerId) return;
    window.removeEventListener("pointerup", release);
    window.removeEventListener("pointercancel", release);
    clearTimeout(failSafe);
    circle.classList.replace("opacity-10", "opacity-0");
    removal = setTimeout(stop, duration + REMOVE_MARGIN);
  };
  const failSafe = setTimeout(() => release(), duration * FAILSAFE_FACTOR);
  window.addEventListener("pointerup", release);
  window.addEventListener("pointercancel", release);
  live.add(stop);

  void circle.offsetWidth;
  circle.classList.replace("scale-0", "scale-100");
};

export function useRipple<T extends HTMLElement = HTMLButtonElement>({
  duration = DEFAULT_DURATION,
  ref,
}: RippleOptions<T> = {}) {
  const hosts = useRef(new Map<T, () => void>());
  const live = useRef(new Set<() => void>());

  useEffect(() => {
    const pending = live.current;
    return () => {
      for (const stop of pending) stop();
      pending.clear();
    };
  }, []);

  useLayoutEffect(() => {
    for (const host of hosts.current.keys()) applyHostClasses(host);
  });

  return useCallback(
    (node: T | null) => {
      if (!node) {
        for (const [host, dispose] of hosts.current)
          if (!host.isConnected) {
            dispose();
            hosts.current.delete(host);
          }
        return;
      }

      hosts.current.get(node)?.();
      applyHostClasses(node);
      const spawn = (event: PointerEvent) => spawnCircle(node, event, duration, live.current);
      node.addEventListener("pointerdown", spawn);
      const dispose = () => {
        node.removeEventListener("pointerdown", spawn);
        hosts.current.delete(node);
      };
      hosts.current.set(node, dispose);
      const detachExternal = assignRef(ref, node);
      return () => {
        dispose();
        detachExternal();
      };
    },
    [duration, ref],
  );
}
