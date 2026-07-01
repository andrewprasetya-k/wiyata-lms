import { useChatRoomSummary } from "./useChatRoomSummary";

export function useChatUnreadCount() {
  const {
    totalUnreadCount,
    badgeLabel,
    loading,
    error,
    refreshRooms,
  } = useChatRoomSummary();

  return {
    unreadCount: totalUnreadCount,
    badgeLabel,
    loading,
    error,
    refreshUnreadCount: refreshRooms,
  };
}
