export const hashPodColorIndex = (podName: string, paletteSize: number): number => {
  if (paletteSize <= 0) {
    return 0;
  }

  let hash = 0x811c9dc5;
  for (const character of podName) {
    hash ^= character.codePointAt(0) ?? 0;
    hash = Math.imul(hash, 0x01000193);
  }

  return (hash >>> 0) % paletteSize;
};

export const buildStablePodColorMap = (
  podNames: string[],
  palette: string[],
  fallbackColor: string
): Record<string, string> => {
  const colorMap: Record<string, string> = { __fallback__: fallbackColor };
  const usedColorIndexes = new Set<number>();
  const normalizedPodNames = Array.from(
    new Set(podNames.map((podName) => podName.trim()).filter(Boolean))
  ).sort((left, right) => left.localeCompare(right));

  normalizedPodNames.forEach((podName) => {
    let colorIndex = hashPodColorIndex(podName, palette.length);
    if (usedColorIndexes.size < palette.length) {
      while (usedColorIndexes.has(colorIndex)) {
        colorIndex = (colorIndex + 1) % palette.length;
      }
      usedColorIndexes.add(colorIndex);
    }
    colorMap[podName] = palette[colorIndex] ?? fallbackColor;
  });

  return colorMap;
};
