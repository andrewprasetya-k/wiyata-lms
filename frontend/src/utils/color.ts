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
