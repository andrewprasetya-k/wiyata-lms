export const subjectPalette = [
  "#7aa7d9",
  "#e58f86",
  "#b889c9",
  "#e5ad72",
  "#74bfa5",
  "#8f95d3",
  "#d99ab5",
  "#9dbb73",
];

export function getSubjectColor(seed?: string | null) {
  const normalized = seed?.trim() || "wiyata";
  const hash = hashString(normalized);
  return subjectPalette[Math.abs(hash) % subjectPalette.length];
}

export function resolveSubjectColor(subject: {
  color?: string | null;
  subjectColor?: string | null;
  subjectId?: string | null;
  subjectClassId?: string | null;
  subjectName?: string | null;
  subjectCode?: string | null;
}) {
  return (
    subject.color ||
    subject.subjectColor ||
    getSubjectColor(
      subject.subjectId ||
        subject.subjectClassId ||
        subject.subjectName ||
        subject.subjectCode,
    )
  );
}

function hashString(value: string) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(index);
    hash |= 0;
  }
  return hash;
}

export function normalizeSubjectColor(color?: string | null) {
  return color?.trim() ?? "";
}

export function isValidSubjectColor(color: string) {
  return /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(color);
}

export function toColorPickerValue(color: string) {
  const normalized = normalizeSubjectColor(color);
  if (/^#[0-9a-fA-F]{6}$/.test(normalized)) return normalized;
  if (/^#[0-9a-fA-F]{3}$/.test(normalized)) {
    const [, r, g, b] = normalized;
    return `#${r}${r}${g}${g}${b}${b}`;
  }
  if (/^#[0-9a-fA-F]{8}$/.test(normalized)) {
    return normalized.slice(0, 7);
  }
  return "#4f46e5";
}

export function subjectDisplayColor(subject: {
  color?: string | null;
  subjectId?: string | null;
  subjectName?: string | null;
}) {
  return (
    subject.color || getSubjectColor(subject.subjectId || subject.subjectName)
  );
}
