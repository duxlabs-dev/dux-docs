import React, {useId} from 'react';
import {shapePath, type ShapeName} from './paths';

type Layer = {
  shape: ShapeName;
  x: number;
  y: number;
  size: number;
  rotation: number;
  fill: string;
  overlap?: string;
};

const compositions: Record<'bloom' | 'garden' | 'header', Layer[]> = {
  bloom: [
    {shape: 'Flower', x: -25, y: 10, size: 280, rotation: -12, fill: 'primary-container'},
    {shape: 'SoftBurst', x: 130, y: 150, size: 195, rotation: 15, fill: 'secondary-container', overlap: 'tertiary-container'},
    {shape: 'Cookie6Sided', x: 160, y: 20, size: 90, rotation: 10, fill: 'primary', overlap: 'tertiary-container'},
  ],
  garden: [
    {shape: 'Puffy', x: 70, y: -25, size: 240, rotation: 20, fill: 'secondary-container'},
    {shape: 'Clover4Leaf', x: 10, y: 125, size: 210, rotation: -15, fill: 'primary-container', overlap: 'tertiary-container'},
    {shape: 'Cookie6Sided', x: 220, y: 210, size: 70, rotation: 10, fill: 'primary'},
  ],
  header: [
    {shape: 'Puffy', x: 95, y: 110, size: 230, rotation: 20, fill: 'primary-container'},
    {shape: 'Flower', x: 35, y: 30, size: 195, rotation: -15, fill: 'primary', overlap: 'tertiary-container'},
    {shape: 'Cookie6Sided', x: 210, y: 25, size: 85, rotation: 10, fill: 'secondary-container', overlap: 'tertiary-container'},
  ],
};

function transform(layer: Layer) {
  return `translate(${layer.x} ${layer.y}) rotate(${layer.rotation} ${layer.size / 2} ${layer.size / 2}) scale(${layer.size})`;
}

export default function Composition({variant, className}: {variant: keyof typeof compositions; className?: string}) {
  const id = useId();
  const layers = compositions[variant];

  return (
    <svg viewBox="0 0 320 320" className={className} overflow="visible" aria-hidden="true" focusable="false">
      <defs>
        {layers.map((layer, index) => (
          <mask key={index} id={`${id}-${index}`} maskUnits="userSpaceOnUse" x={layer.x - layer.size} y={layer.y - layer.size} width={layer.size * 3} height={layer.size * 3} style={{maskType: 'alpha'}}>
            <path d={shapePath(layer.shape)} transform={transform(layer)} fill="white" />
          </mask>
        ))}
      </defs>
      {layers.map((layer, index) => (
        <g key={index} mask={`url(#${id}-${index})`}>
          <rect x="-1" y="-1" width="3" height="3" transform={transform(layer)} fill={`var(--md-sys-color-${layer.fill})`} />
          {layer.overlap && layers.slice(0, index).map((previous, previousIndex) => (
            <path key={previousIndex} d={shapePath(previous.shape)} transform={transform(previous)} fill={`var(--md-sys-color-${layer.overlap})`} />
          ))}
        </g>
      ))}
    </svg>
  );
}
