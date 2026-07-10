import type { StudentAssignmentInboxItem } from "../types/assignment";
import { parseBackendTimestamp } from "./date";

export function compareAssignments(
  a: StudentAssignmentInboxItem,
  b: StudentAssignmentInboxItem,
) {
  const overdueNotSubmittedDiff =
    Number(b.isOverdue && !b.isSubmitted) - Number(a.isOverdue && !a.isSubmitted);
  if (overdueNotSubmittedDiff !== 0) return overdueNotSubmittedDiff;

  const notSubmittedDiff = Number(!b.isSubmitted) - Number(!a.isSubmitted);
  if (notSubmittedDiff !== 0) return notSubmittedDiff;

  const deadlineDiff = getDeadlineTime(a.deadline) - getDeadlineTime(b.deadline);
  if (deadlineDiff !== 0) return deadlineDiff;

  return (a.assignmentTitle || "").localeCompare(b.assignmentTitle || "");
}

export function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = parseBackendTimestamp(deadline)?.getTime() ?? Number.NaN;
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value;
}

export function assignmentStatusLabel(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "Sudah dinilai";
  if (item.isSubmitted) return "Sudah dikumpulkan";
  if (item.isOverdue) return "Lewat deadline";
  return "Belum dikumpulkan";
}

export function assignmentStatusClasses(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "bg-success-soft text-success";
  if (item.isSubmitted) return "bg-brand-soft text-brand";
  if (item.isOverdue) return "bg-danger-soft text-danger";
  return "bg-warning-soft text-warning";
}
