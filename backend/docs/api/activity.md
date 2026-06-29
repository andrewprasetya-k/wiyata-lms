# Academic Activity API

Academic Activity is a normalized, actionable academic stream. It is not the notification feed, class feed, or a calendar table. The API merges academic sources into one role-aware response for dashboard My Day, future activity pages, and future calendar markers.

## List Academic Activity

`GET /api/academic-activity?from=YYYY-MM-DD&to=YYYY-MM-DD`

Authentication:
- Requires `Authorization: Bearer <token>`
- Requires active school context through `SchoolId`
- Requires active school membership
- Requires `student` or `teacher` role in the active school

Query:
- `from` optional, format `YYYY-MM-DD`
- `to` optional, format `YYYY-MM-DD`
- If omitted, the range defaults to today through today + 7 days
- Maximum range is 60 days

Response:

```json
{
  "items": [
    {
      "id": "assignment_due:uuid",
      "type": "assignment_due",
      "title": "Tugas Bab 3",
      "description": "Tenggat tugas Matematika · XII IPA 1",
      "date": "2026-07-01",
      "time": "15:30",
      "priority": "high",
      "subject": {
        "id": "uuid",
        "name": "Matematika",
        "code": "MTK",
        "color": "#4f46e5"
      },
      "class": {
        "id": "uuid",
        "name": "XII IPA 1",
        "code": "XII-IPA-1"
      },
      "link": "/student/subjects/uuid/assignments/uuid",
      "metadata": {
        "assignmentId": "uuid",
        "subjectClassId": "uuid"
      }
    }
  ]
}
```

Items are sorted ascending by event datetime.

## Student Activity Types

- `assignment_due` - upcoming assignment deadlines for actively enrolled classes; high priority
- `material_created` - new materials in actively enrolled classes
- `feed_posted` - class feed announcements visible to the student
- `assignment_graded` - assessed student submissions

Student access is derived from active class enrollment:
- `school_users.deleted_at IS NULL`
- `enrollments.left_at IS NULL`
- active school scope must match

## Teacher Activity Types

- `submission_received` - received submissions that are already assessed, to avoid duplicating pending review items in the MVP
- `submission_pending_review` - unassessed submissions waiting for review; high priority
- `assignment_due` - deadlines for assignments in subject classes taught by the teacher; high priority
- `feed_comment` - comments on feed posts in classes taught by the teacher

Teacher access is derived from active subject-class teaching assignment and active teacher enrollment:
- `school_users.deleted_at IS NULL`
- `enrollments.left_at IS NULL`
- active school scope must match

## Notes

- This API does not create or duplicate notification rows.
- Chat, notification, admin, and recurring calendar events are intentionally outside this MVP.
- Subject color is returned when available and can be used by frontend timeline/calendar UI with fallback colors.
