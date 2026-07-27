export interface BracketedLogPrefix {
  label: string;
  prefix: string;
  remainder: string;
}

export const parseBracketedLogPrefix = (line: string): BracketedLogPrefix | null => {
  if (line[0] !== '[') {
    return null;
  }

  const closingBracket = line.indexOf(']', 1);
  if (closingBracket <= 1) {
    return null;
  }

  let remainderStart = closingBracket + 1;
  while (remainderStart < line.length && line[remainderStart]?.trim() === '') {
    remainderStart += 1;
  }

  return {
    label: line.slice(1, closingBracket),
    prefix: line.slice(0, remainderStart),
    remainder: line.slice(remainderStart),
  };
};
