import React from 'react';
import {shapePath, type ShapeName} from './paths';

export default function ExpressiveShape({shape = 'Flower', className}: {shape?: ShapeName; className?: string}) {
  return (
    <svg viewBox="0 0 1 1" className={className} aria-hidden="true" focusable="false">
      <path d={shapePath(shape)} fill="currentColor" />
    </svg>
  );
}
