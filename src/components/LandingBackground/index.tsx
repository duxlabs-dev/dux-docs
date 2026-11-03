import React, { useEffect, useRef } from 'react';
import ExpressiveShape from '../ExpressiveShape';
import styles from './styles.module.css';

/*
 * Material 3 Expressive landing backdrop.
 * Depth comes from tonal blobs and shapes that breathe and scale with
 * page scroll. A faint coordinate grid surfaces only around the cursor.
 */
export default function LandingBackground(): React.JSX.Element {
  const canvasRef = useRef<HTMLDivElement>(null);

  // Cursor-revealed grid: eased pointer position written to CSS vars, one RAF loop while moving.
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    if (window.matchMedia('(hover: none), (prefers-reduced-motion: reduce)').matches) return;

    let frame = 0;
    let targetX = window.innerWidth / 2;
    let targetY = window.innerHeight / 2;
    let x = targetX;
    let y = targetY;

    const update = () => {
      x += (targetX - x) * 0.12;
      y += (targetY - y) * 0.12;
      canvas.style.setProperty('--grid-x', `${x.toFixed(1)}px`);
      canvas.style.setProperty('--grid-y', `${y.toFixed(1)}px`);
      frame = Math.abs(targetX - x) > 0.4 || Math.abs(targetY - y) > 0.4 ? requestAnimationFrame(update) : 0;
    };

    const onMove = (e: MouseEvent) => {
      targetX = e.clientX;
      targetY = e.clientY;
      canvas.dataset.pointer = 'in';
      if (!frame) frame = requestAnimationFrame(update);
    };
    const onLeave = () => {
      delete canvas.dataset.pointer;
    };

    window.addEventListener('mousemove', onMove, { passive: true });
    document.addEventListener('mouseleave', onLeave);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseleave', onLeave);
    };
  }, []);

  return (
    <>
      {/* 1. Viewport-fixed tonal blobs. */}
      <div ref={canvasRef} className={styles.fixedCanvas} aria-hidden="true">
        <div className={styles.blobPrimary} />
        <div className={styles.blobSecondary} />
        <div className={styles.blobTertiary} />
        <div className={styles.gridReveal} />
      </div>

      {/* 2. Hero-anchored expressive shapes. They scale and rotate with scroll. */}
      <div className={styles.heroShapesLayer} aria-hidden="true">
        <div className={styles.shapeHeroLeft}>
          <ExpressiveShape shape="Flower" className={styles.shapeSvg} />
        </div>
        <div className={styles.shapeHeroRight}>
          <ExpressiveShape shape="SoftBurst" className={styles.shapeSvg} />
        </div>
        <div className={styles.shapeDrifterLeft}>
          <ExpressiveShape shape="Cookie6Sided" className={styles.shapeSvg} />
        </div>
        <div className={styles.shapeDrifterRight}>
          <ExpressiveShape shape="Clover4Leaf" className={styles.shapeSvg} />
        </div>
      </div>
    </>
  );
}
