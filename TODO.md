# EduVerse MVP TODO

This file tracks visible unfinished work that is intentionally kept in the product navigation for planning and follow-up. These items are not part of the current school MVP unless explicitly promoted.

## Post-MVP Features

- Realtime chat for teacher and student workspaces.
- Global student notes workspace, assignment/subject notes, and autosave.
- Student material progress tracking.
- Feed comments and replies.
- Feed attachments.
- Realtime notifications and chat delivery.
- Superadmin platform management for schools and global users.

## Placeholder Routes

- `/teacher/chat` - planned realtime teacher chat.
- `/student/chat` - planned realtime student chat.
- `/student/notes` - planned student notes workspace.
- `/superadmin/schools` - planned platform school management.
- `/superadmin/users` - planned platform user management.

## Student Features

- Student Chat: keep route and navigation visible as a post-MVP feature marker.
- Student Notes: material detail now supports one private plain-text note with manual save.
- Student global Notes route, subject notes, and assignment notes remain post-MVP.
- Student material progress tracking: currently planned copy only, no completion state is stored.

## Teacher Features

- Teacher Chat: keep route and navigation visible as a post-MVP feature marker.
- Feed comments, reactions, and attachments are deferred.
- Realtime class communication is deferred.

## Superadmin Features

- Superadmin dashboard is a static post-MVP platform overview.
- Schools management is planned after the school MVP flow.
- Platform users management is planned after the school MVP flow.

## Known UI Follow-ups

- Replace planned-copy panels with real implementations when backend contracts are ready.
- Keep placeholder pages visibly intentional and avoid showing fake data.
- Revisit notification surfaces when realtime delivery is introduced.
- Revisit feed detail interactions when comments and attachments are supported.

Admin assessment weight UI

- Fix input parsing for number/string values.
- Ensure total validation blocks invalid submit.
- Revisit UX after backend weights endpoint is stable.

# Set-up super admin e2e process

- user registration
- user school registration
- school authentication through email
- user confirmation on school member invitation
