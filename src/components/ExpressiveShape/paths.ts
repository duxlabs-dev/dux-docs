import {MaterialShapes, roundedPolygonToPath, type MaterialShapeName} from 'material-shapes-ts';

export type ShapeName = MaterialShapeName;

const cache = new Map<ShapeName, string>();

export function shapePath(name: ShapeName) {
  const cached = cache.get(name);
  if (cached) return cached;
  const path = roundedPolygonToPath(MaterialShapes.byName(name)).toSvgPathData();
  cache.set(name, path);
  return path;
}
