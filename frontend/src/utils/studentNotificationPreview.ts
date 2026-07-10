import type { NotificationItem } from "../types/dashboard";

export function initials(value: string) {
  return value
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

export function notificationTitle(item: NotificationItem) {
  if (item.type === "assignment_created") return "Tugas baru";
  if (item.type === "feed_posted") return "Pengumuman kelas baru";
  if (item.type === "assignment_graded") return "Tugas sudah dinilai";
  if (item.type === "material_added") return "Materi baru";
  return item.title || "Notifikasi";
}

export function notificationBadge(item: NotificationItem) {
  if (item.type === "assignment_created") return "TB";
  if (item.type === "feed_posted") return "PG";
  if (item.type === "assignment_graded") return "AG";
  if (item.type === "material_added") return "MT";
  return initials(notificationTitle(item));
}

export function notificationMessage(item: NotificationItem) {
  return item.message || "Buka notifikasi untuk melihat informasi terbaru.";
}

export function notificationAriaLabel(item: NotificationItem) {
  const action = item.link ? "Buka notifikasi" : "Tandai notifikasi dibaca";
  return `${action}: ${notificationTitle(item)}`;
}

export function isInternalNotificationLink(link?: string) {
  return Boolean(link && link.startsWith("/") && !link.startsWith("//"));
}
