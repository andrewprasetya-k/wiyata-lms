import type { AcademicActivityItem } from "../../types/activity"
import { getSubjectColor } from "../../utils/color"

export type ActivityRole = "student" | "teacher"

export interface ActivityFilter {
  label: string
  value: string
  types: string[]
}

export interface ActivityRange {
  label: string
  value: "today" | "7d" | "30d"
  days: number
}

export const activityRanges: ActivityRange[] = [
  { label: "Hari Ini", value: "today", days: 0 },
  { label: "7 Hari", value: "7d", days: 7 },
  { label: "30 Hari", value: "30d", days: 30 },
]

export function activityFilters(role: ActivityRole): ActivityFilter[] {
  if (role === "teacher") {
    return [
      { label: "Semua", value: "all", types: [] },
      {
        label: "Pengumpulan",
        value: "submission_received",
        types: ["submission_received"],
      },
      {
        label: "Perlu Dinilai",
        value: "submission_pending_review",
        types: ["submission_pending_review"],
      },
      { label: "Tugas", value: "assignment_due", types: ["assignment_due"] },
      { label: "Feed", value: "feed_comment", types: ["feed_comment"] },
    ]
  }

  return [
    { label: "Semua", value: "all", types: [] },
    { label: "Tugas", value: "assignment_due", types: ["assignment_due"] },
    { label: "Materi", value: "material_created", types: ["material_created"] },
    { label: "Feed", value: "feed_posted", types: ["feed_posted"] },
    { label: "Nilai", value: "assignment_graded", types: ["assignment_graded"] },
  ]
}

export function activityTypeLabel(type: string, role: ActivityRole) {
  if (role === "teacher") {
    const labels: Record<string, string> = {
      submission_received: "Pengumpulan",
      submission_pending_review: "Perlu dinilai",
      assignment_due: "Tugas",
      feed_comment: "Feed",
    }
    return labels[type] ?? "Aktivitas"
  }

  const labels: Record<string, string> = {
    assignment_due: "Tugas",
    material_created: "Materi",
    feed_posted: "Feed",
    assignment_graded: "Nilai",
  }
  return labels[type] ?? "Aktivitas"
}

export function activitySubjectColor(item: AcademicActivityItem) {
  return (
    item.subject?.color ||
    getSubjectColor(
      item.subject?.id ||
        item.subject?.name ||
        item.subject?.code ||
        item.class?.id ||
        item.class?.name,
    )
  )
}

export function activityTimestamp(item: AcademicActivityItem) {
  const parsed = new Date(`${item.date || ""}T${item.time || "00:00"}`)
  const value = parsed.getTime()
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value
}

export function compareActivities(
  left: AcademicActivityItem,
  right: AcademicActivityItem,
) {
  const priorityDiff =
    priorityWeight(right.priority) - priorityWeight(left.priority)
  if (priorityDiff !== 0) return priorityDiff
  return activityTimestamp(left) - activityTimestamp(right)
}

export function priorityWeight(priority?: string | null) {
  return priority === "high" ? 1 : 0
}

export function isInternalActivityLink(link?: string | null) {
  return Boolean(link && link.startsWith("/") && !link.startsWith("//"))
}

export function formatActivityDate(value?: string | null) {
  const date = parseActivityDate(value)
  if (!date) return "Tanggal belum tersedia"

  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(date)
}

export function activityRelativeLabel(item: AcademicActivityItem) {
  const date = parseActivityDate(item.date)
  if (!date) return "Tanggal belum tersedia"

  const today = startOfDay(new Date())
  const target = startOfDay(date)
  const diffDays = Math.round(
    (target.getTime() - today.getTime()) / 86_400_000,
  )

  let label = formatActivityDate(item.date)
  if (diffDays === 0) label = "Hari ini"
  if (diffDays === 1) label = "Besok"

  return item.time ? `${label}, ${item.time}` : label
}

export function activityGroupLabel(value?: string | null) {
  const date = parseActivityDate(value)
  if (!date) return "Nanti"

  const today = startOfDay(new Date())
  const target = startOfDay(date)
  const diffDays = Math.round(
    (target.getTime() - today.getTime()) / 86_400_000,
  )

  if (diffDays < 0) return "Sebelumnya"
  if (diffDays === 0) return "Hari Ini"
  if (diffDays === 1) return "Besok"
  if (diffDays <= 6) return "Minggu Ini"
  return "Nanti"
}

export function parseActivityDate(value?: string | null) {
  if (!value) return null
  const parts = value.split("-").map(Number)
  if (parts.length !== 3 || parts.some((part) => Number.isNaN(part))) {
    return null
  }
  const [year, month, day] = parts
  return new Date(year, month - 1, day)
}

export function formatApiDate(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}

export function addDays(date: Date, days: number) {
  const next = new Date(date)
  next.setDate(next.getDate() + days)
  return next
}

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}
