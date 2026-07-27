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

  podNames.forEach((podName) => {
    const trimmed = podName.trim();
    if (!trimmed) {
      return;
    }
    colorMap[trimmed] = palette[hashPodColorIndex(trimmed, palette.length)] ?? fallbackColor;
  });

  return colorMap;
};
