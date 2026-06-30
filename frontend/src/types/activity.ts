export type ActivityType =
  | "assignment_due"
  | "material_created"
  | "feed_posted"
  | "assignment_graded"
  | "submission_received"
  | "submission_pending_review"
  | "feed_comment"
  | string

export type ActivityPriority = "normal" | "high" | string

export interface ActivitySubject {
  id?: string | null
  name?: string | null
  code?: string | null
  color?: string | null
}

export interface ActivityClass {
  id?: string | null
  name?: string | null
  code?: string | null
}

export interface AcademicActivityItem {
  id: string
  type: ActivityType
  title: string
  description: string
  date: string
  time?: string | null
  priority: ActivityPriority
  subject?: ActivitySubject | null
  class?: ActivityClass | null
  link?: string | null
  metadata?: Record<string, unknown>
}

export interface AcademicActivityResponse {
  items: AcademicActivityItem[]
}

export interface AcademicActivityParams {
  from?: string
  to?: string
}
