package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	chatRoomTypeGroup   = "group"
	chatRoomTypeDM      = "dm"
	chatRefTypeSchool   = "school"
	chatMessageTypeText = "text"
	chatMessageTypeFile = "file"
	maxChatMessageLimit = 50
	maxChatContentLen   = 5000
	maxChatAttachments  = 5
	maxChatRoomNameLen = 150
)

type ChatService interface {
	ListMyRooms(userID string, schoolID string, search string) ([]dto.ChatRoomDTO, error)
	ListMembers(userID string, schoolID string, search string, excludeRoomID *string) ([]dto.ChatMemberDTO, error)
	OpenSchoolRoom(userID string, schoolID string) (*dto.ChatRoomDTO, error)
	OpenDirectMessage(userID string, schoolID string, targetUserID string) (*dto.ChatRoomDTO, error)
	CreateGroupRoom(userID string, schoolID string, roomName string, memberUserIDs []string) (*dto.ChatRoomDTO, error)
	GetGroupInfo(userID string, schoolID string, roomID string) (*dto.ChatGroupInfoDTO, error)
	RenameGroupRoom(userID string, schoolID string, roomID string, roomName string) (*dto.ChatRoomDTO, error)
	LeaveGroupRoom(userID string, schoolID string, roomID string) error
	AddGroupMembers(userID string, schoolID string, roomID string, memberUserIDs []string) error
	RemoveGroupMember(userID string, schoolID string, roomID string, targetUserID string) error
	ListMessages(userID string, schoolID string, roomID string, limit int, before *time.Time) (*dto.ChatMessagesResponseDTO, error)
	CreateMessage(userID string, schoolID string, roomID string, content string, mediaIDs []string) (*dto.ChatMessageDTO, error)
	MarkRead(userID string, schoolID string, roomID string, lastReadMessageID *string) (*dto.ChatReadReceiptDTO, error)
	GetReadSummary(userID string, schoolID string, roomID string) (*dto.ChatReadSummaryDTO, error)
	ListRealtimeRecipients(userID string, schoolID string, roomID string) ([]string, error)
	CanAccessSchoolChat(userID string, schoolID string) (bool, error)
	CanAccessRoom(userID string, schoolID string, roomID string) (bool, *repository.ChatRoomRow, error)
}

type chatService struct {
	repo      repository.ChatRepository
	mediaRepo repository.MediaRepository
}

func NewChatService(repo repository.ChatRepository, mediaRepo repository.MediaRepository) ChatService {
	return &chatService{repo: repo, mediaRepo: mediaRepo}
}

func (s *chatService) ListMyRooms(userID string, schoolID string, search string) ([]dto.ChatRoomDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	rows, err := s.repo.ListSchoolRooms(userID, schoolID, strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}

	roomIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		roomIDs = append(roomIDs, row.RoomID)
	}
	unreadByRoom, err := s.repo.UnreadCounts(roomIDs, userID)
	if err != nil {
		return nil, err
	}

	rooms := make([]dto.ChatRoomDTO, 0, len(rows))
	for _, row := range rows {
		rooms = append(rooms, mapChatRoomRow(row, unreadByRoom[row.RoomID]))
	}
	return rooms, nil
}

func (s *chatService) ListMembers(userID string, schoolID string, search string, excludeRoomID *string) ([]dto.ChatMemberDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	if excludeRoomID != nil && *excludeRoomID != "" {
		allowed, room, err := s.CanAccessRoom(userID, schoolID, *excludeRoomID)
		if err != nil {
			return nil, err
		}
		if !allowed || !isCustomGroupRoom(room) {
			return nil, fmt.Errorf("forbidden: chat room access denied")
		}
		isAdmin, err := s.repo.UserIsRoomAdmin(userID, *excludeRoomID)
		if err != nil {
			return nil, err
		}
		if !isAdmin {
			return nil, fmt.Errorf("forbidden: chat group admin required")
		}
	}

	rows, err := s.repo.ListChatMembers(schoolID, strings.TrimSpace(search), excludeRoomID)
	if err != nil {
		return nil, err
	}

	members := make([]dto.ChatMemberDTO, 0, len(rows))
	for _, row := range rows {
		roles := make([]string, 0)
		if row.Roles != "" {
			for _, role := range strings.Split(row.Roles, ",") {
				role = strings.TrimSpace(role)
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		members = append(members, dto.ChatMemberDTO{
			UserID:   row.UserID,
			FullName: row.FullName,
			Email:    row.Email,
			Roles:    roles,
		})
	}
	return members, nil
}

func (s *chatService) OpenSchoolRoom(userID string, schoolID string) (*dto.ChatRoomDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	room, err := s.repo.GetSchoolRoom(schoolID)
	if err != nil {
		return nil, err
	}

	context, err := s.repo.GetRoomContext(room.ID, schoolID, userID)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.UnreadCount(room.ID, userID)
	if err != nil {
		return nil, err
	}
	roomDTO := mapChatRoomRow(*context, unread)
	return &roomDTO, nil
}

func (s *chatService) OpenDirectMessage(userID string, schoolID string, targetUserID string) (*dto.ChatRoomDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		return nil, fmt.Errorf("chat dm target is required")
	}
	if targetUserID == userID {
		return nil, fmt.Errorf("chat dm cannot target self")
	}

	activeMembers, err := s.repo.UsersAreActiveSchoolMembers([]string{targetUserID}, schoolID)
	if err != nil {
		return nil, err
	}
	if !activeMembers[targetUserID] {
		return nil, fmt.Errorf("invalid chat dm target")
	}

	room, err := s.repo.FindDirectMessageRoom(schoolID, userID, targetUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || room == nil {
		room, err = s.repo.CreateDirectMessageRoom(schoolID, userID, targetUserID)
		if err != nil {
			return nil, err
		}
	}

	unread, err := s.repo.UnreadCount(room.RoomID, userID)
	if err != nil {
		return nil, err
	}
	mapped := mapChatRoomRow(*room, unread)
	return &mapped, nil
}

func (s *chatService) CreateGroupRoom(userID string, schoolID string, roomName string, memberUserIDs []string) (*dto.ChatRoomDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	roomName = strings.TrimSpace(roomName)
	if roomName == "" {
		return nil, fmt.Errorf("chat group room name is required")
	}
	if len([]rune(roomName)) > maxChatRoomNameLen {
		return nil, fmt.Errorf("chat group room name exceeds %d characters", maxChatRoomNameLen)
	}
	if len(memberUserIDs) == 0 {
		return nil, fmt.Errorf("chat group members are required")
	}

	seenInput := make(map[string]bool, len(memberUserIDs))
	for _, memberID := range memberUserIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			return nil, fmt.Errorf("chat group member is required")
		}
		if seenInput[memberID] {
			return nil, fmt.Errorf("duplicate chat group member")
		}
		seenInput[memberID] = true
	}

	memberSet := make(map[string]bool, len(memberUserIDs)+1)
	memberSet[userID] = true
	for _, memberID := range memberUserIDs {
		memberSet[memberID] = true
	}

	allMemberIDs := make([]string, 0, len(memberSet))
	for memberID := range memberSet {
		allMemberIDs = append(allMemberIDs, memberID)
	}

	activeMembers, err := s.repo.UsersAreActiveSchoolMembers(allMemberIDs, schoolID)
	if err != nil {
		return nil, err
	}
	for _, memberID := range allMemberIDs {
		if !activeMembers[memberID] {
			return nil, fmt.Errorf("invalid chat group member")
		}
	}

	room, err := s.repo.CreateGroupRoomWithMembers(schoolID, roomName, userID, allMemberIDs)
	if err != nil {
		return nil, err
	}
	context, err := s.repo.GetRoomContext(room.ID, schoolID, userID)
	if err != nil {
		return nil, err
	}
	mapped := mapChatRoomRow(*context, 0)
	return &mapped, nil
}

func (s *chatService) GetGroupInfo(userID string, schoolID string, roomID string) (*dto.ChatGroupInfoDTO, error) {
	allowed, room, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed || !isCustomGroupRoom(room) {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	info, members, err := s.repo.GetGroupInfo(roomID, schoolID)
	if err != nil {
		return nil, err
	}
	mapped := mapChatGroupInfo(*info, members)
	return &mapped, nil
}

func (s *chatService) RenameGroupRoom(userID string, schoolID string, roomID string, roomName string) (*dto.ChatRoomDTO, error) {
	roomName = strings.TrimSpace(roomName)
	if len([]rune(roomName)) < 3 {
		return nil, fmt.Errorf("chat group room name is too short")
	}
	if len([]rune(roomName)) > maxChatRoomNameLen {
		return nil, fmt.Errorf("chat group room name exceeds %d characters", maxChatRoomNameLen)
	}
	if err := s.requireGroupAdmin(userID, schoolID, roomID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateGroupRoomName(roomID, schoolID, roomName); err != nil {
		return nil, err
	}
	context, err := s.repo.GetRoomContext(roomID, schoolID, userID)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.UnreadCount(roomID, userID)
	if err != nil {
		return nil, err
	}
	mapped := mapChatRoomRow(*context, unread)
	return &mapped, nil
}

func (s *chatService) LeaveGroupRoom(userID string, schoolID string, roomID string) error {
	allowed, room, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return err
	}
	if !allowed || !isCustomGroupRoom(room) {
		return fmt.Errorf("forbidden: chat room access denied")
	}
	return s.repo.LeaveGroupRoom(roomID, schoolID, userID)
}

func (s *chatService) AddGroupMembers(userID string, schoolID string, roomID string, memberUserIDs []string) error {
	if err := s.requireGroupAdmin(userID, schoolID, roomID); err != nil {
		return err
	}
	memberIDs, err := s.validateGroupMemberIDs(memberUserIDs, schoolID)
	if err != nil {
		return err
	}
	return s.repo.AddGroupRoomMembers(roomID, schoolID, memberIDs)
}

func (s *chatService) RemoveGroupMember(userID string, schoolID string, roomID string, targetUserID string) error {
	if targetUserID == userID {
		return fmt.Errorf("chat group cannot remove self")
	}
	if strings.TrimSpace(targetUserID) == "" {
		return fmt.Errorf("chat group member is required")
	}
	if err := s.requireGroupAdmin(userID, schoolID, roomID); err != nil {
		return err
	}
	activeMembers, err := s.repo.UsersAreActiveSchoolMembers([]string{targetUserID}, schoolID)
	if err != nil {
		return err
	}
	if !activeMembers[targetUserID] {
		return fmt.Errorf("invalid chat group member")
	}
	return s.repo.RemoveGroupRoomMember(roomID, schoolID, targetUserID)
}

func (s *chatService) ListMessages(userID string, schoolID string, roomID string, limit int, before *time.Time) (*dto.ChatMessagesResponseDTO, error) {
	allowed, _, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	if limit <= 0 || limit > maxChatMessageLimit {
		limit = maxChatMessageLimit
	}
	rows, err := s.repo.ListMessages(roomID, limit+1, before)
	if err != nil {
		return nil, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[1:]
	}

	messages := make([]dto.ChatMessageDTO, 0, len(rows))
	messageIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		messageIDs = append(messageIDs, row.MessageID)
	}
	attachmentsByMessage, err := s.repo.ListMessageAttachments(messageIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		messages = append(messages, mapChatMessageRow(row, userID, attachmentsByMessage[row.MessageID]))
	}

	var nextBefore *string
	if hasMore && len(rows) > 0 {
		value := formatChatTime(rows[0].CreatedAt)
		nextBefore = &value
	}

	return &dto.ChatMessagesResponseDTO{
		Messages:   messages,
		NextBefore: nextBefore,
		HasMore:    hasMore,
	}, nil
}

func (s *chatService) CreateMessage(userID string, schoolID string, roomID string, content string, mediaIDs []string) (*dto.ChatMessageDTO, error) {
	allowed, _, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	content = strings.TrimSpace(content)
	attachmentMediaIDs, err := validateChatMediaIDs(mediaIDs)
	if err != nil {
		return nil, err
	}
	if content == "" && len(attachmentMediaIDs) == 0 {
		return nil, fmt.Errorf("chat message content is required")
	}
	if len([]rune(content)) > maxChatContentLen {
		return nil, fmt.Errorf("chat message content exceeds %d characters", maxChatContentLen)
	}
	if len(attachmentMediaIDs) > maxChatAttachments {
		return nil, fmt.Errorf("chat message attachments exceed %d files", maxChatAttachments)
	}
	if len(attachmentMediaIDs) > 0 {
		attachmentMediaIDs, err = prepareAttachableMediaIDs(s.mediaRepo, attachmentMediaIDs, schoolID, userID, false)
		if err != nil {
			return nil, err
		}
	}

	messageType := chatMessageTypeText
	if len(attachmentMediaIDs) > 0 {
		messageType = chatMessageTypeFile
	}
	message := domain.ChatMessage{
		RoomID:    roomID,
		UserID:    userID,
		Content:   content,
		Type:      messageType,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateMessageWithAttachments(&message, attachmentMediaIDs); err != nil {
		return nil, err
	}

	row, err := s.repo.GetMessageByID(message.ID, roomID)
	if err != nil {
		return nil, err
	}
	attachmentsByMessage, err := s.repo.ListMessageAttachments([]string{message.ID})
	if err != nil {
		return nil, err
	}
	mapped := mapChatMessageRow(*row, userID, attachmentsByMessage[row.MessageID])
	return &mapped, nil
}

func validateChatMediaIDs(mediaIDs []string) ([]string, error) {
	result := make([]string, 0, len(mediaIDs))
	seen := make(map[string]bool, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		mediaID = strings.TrimSpace(mediaID)
		if mediaID == "" {
			return nil, fmt.Errorf("invalid chat attachment")
		}
		if seen[mediaID] {
			return nil, fmt.Errorf("duplicate chat attachment")
		}
		seen[mediaID] = true
		result = append(result, mediaID)
	}
	return result, nil
}

func (s *chatService) MarkRead(userID string, schoolID string, roomID string, lastReadMessageID *string) (*dto.ChatReadReceiptDTO, error) {
	allowed, _, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	if lastReadMessageID != nil && *lastReadMessageID != "" {
		if _, err := s.repo.GetMessageByID(*lastReadMessageID, roomID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpsertReadReceipt(roomID, userID, lastReadMessageID); err != nil {
		return nil, err
	}
	receipt, err := s.repo.GetReadReceipt(roomID, userID)
	if err != nil {
		return nil, err
	}
	return mapChatReadReceipt(*receipt, userID), nil
}

func (s *chatService) GetReadSummary(userID string, schoolID string, roomID string) (*dto.ChatReadSummaryDTO, error) {
	allowed, room, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	var currentReceipt *repository.ChatReadReceiptRow
	currentReceipt, err = s.repo.GetReadReceipt(roomID, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var memberRows []repository.ChatReadMemberRow
	if isSchoolChatRoom(room, schoolID) {
		memberRows, err = s.repo.ListSchoolReadMembers(roomID, schoolID)
	} else if isCustomGroupRoom(room) || isDirectMessageRoom(room) {
		memberRows, err = s.repo.ListRoomReadMembers(roomID, schoolID)
	} else {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}
	if err != nil {
		return nil, err
	}

	summary := dto.ChatReadSummaryDTO{
		RoomID:  roomID,
		Members: mapChatReadMembers(memberRows),
	}
	if currentReceipt != nil {
		summary.LastReadMessageID = currentReceipt.LastReadMessageID
		if currentReceipt.LastReadAt != nil {
			value := formatChatTime(*currentReceipt.LastReadAt)
			summary.LastReadAt = &value
		}
	}
	return &summary, nil
}

func (s *chatService) ListRealtimeRecipients(userID string, schoolID string, roomID string) ([]string, error) {
	allowed, room, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}
	if isSchoolChatRoom(room, schoolID) {
		return s.repo.ListSchoolRecipientUserIDs(schoolID)
	}
	if isCustomGroupRoom(room) || isDirectMessageRoom(room) {
		return s.repo.ListRoomRecipientUserIDs(roomID, schoolID)
	}
	return nil, fmt.Errorf("forbidden: chat room access denied")
}

func (s *chatService) CanAccessSchoolChat(userID string, schoolID string) (bool, error) {
	if userID == "" || schoolID == "" {
		return false, nil
	}
	return s.repo.UserIsActiveSchoolMember(userID, schoolID)
}

func (s *chatService) CanAccessRoom(userID string, schoolID string, roomID string) (bool, *repository.ChatRoomRow, error) {
	room, err := s.repo.GetRoomContext(roomID, schoolID, userID)
	if err != nil {
		return false, nil, err
	}
	if room.SchoolID != schoolID {
		return false, room, nil
	}

	activeSchoolMember, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return false, nil, err
	}
	if !activeSchoolMember {
		return false, room, nil
	}

	if isSchoolChatRoom(room, schoolID) {
		return true, room, nil
	}

	if isCustomGroupRoom(room) || isDirectMessageRoom(room) {
		activeRoomMember, err := s.repo.UserIsActiveRoomMember(userID, roomID)
		if err != nil {
			return false, nil, err
		}
		return activeRoomMember, room, nil
	}

	return false, room, nil
}

func (s *chatService) requireGroupAdmin(userID string, schoolID string, roomID string) error {
	allowed, room, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return err
	}
	if !allowed || !isCustomGroupRoom(room) {
		return fmt.Errorf("forbidden: chat room access denied")
	}
	isAdmin, err := s.repo.UserIsRoomAdmin(userID, roomID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return fmt.Errorf("forbidden: chat group admin required")
	}
	return nil
}

func (s *chatService) validateGroupMemberIDs(memberUserIDs []string, schoolID string) ([]string, error) {
	if len(memberUserIDs) == 0 {
		return nil, fmt.Errorf("chat group members are required")
	}
	seenInput := make(map[string]bool, len(memberUserIDs))
	result := make([]string, 0, len(memberUserIDs))
	for _, memberID := range memberUserIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			return nil, fmt.Errorf("chat group member is required")
		}
		if seenInput[memberID] {
			return nil, fmt.Errorf("duplicate chat group member")
		}
		seenInput[memberID] = true
		result = append(result, memberID)
	}

	activeMembers, err := s.repo.UsersAreActiveSchoolMembers(result, schoolID)
	if err != nil {
		return nil, err
	}
	for _, memberID := range result {
		if !activeMembers[memberID] {
			return nil, fmt.Errorf("invalid chat group member")
		}
	}
	return result, nil
}

func mapChatRoomRow(row repository.ChatRoomRow, unread int64) dto.ChatRoomDTO {
	var lastMessage *dto.ChatLastMessageDTO
	var lastMessageAt *string
	if row.LastMessageID != nil && row.LastContent != nil && row.LastMessageAt != nil {
		createdAt := formatChatTime(*row.LastMessageAt)
		lastMessageAt = &createdAt
		lastMessage = &dto.ChatLastMessageDTO{
			MessageID:       *row.LastMessageID,
			Content:         *row.LastContent,
			AttachmentCount: row.LastAttachmentCount,
			CreatedAt:       createdAt,
		}
		if row.LastAttachmentMimeType != nil {
			lastMessage.AttachmentMimeType = *row.LastAttachmentMimeType
		}
		if row.LastAttachmentFileName != nil {
			lastMessage.AttachmentFileName = *row.LastAttachmentFileName
		}
		if row.LastType != nil {
			lastMessage.MessageType = *row.LastType
		}
		if row.LastSenderID != nil {
			lastMessage.SenderID = *row.LastSenderID
		}
		if row.LastSenderName != nil {
			lastMessage.SenderName = *row.LastSenderName
		}
	}

	return dto.ChatRoomDTO{
		RoomID:         row.RoomID,
		RoomName:       resolveRoomName(row),
		RoomType:       row.RoomType,
		RoomRefType:    row.RoomRefType,
		RoomRefID:      row.RoomRefID,
		SchoolID:       row.SchoolID,
		SchoolName:     row.SchoolName,
		DMTargetUserID: row.DMTargetUserID,
		DMTargetName:   row.DMTargetName,
		DMTargetEmail:  row.DMTargetEmail,
		LastMessage:    lastMessage,
		LastMessageAt:  lastMessageAt,
		UnreadCount:    unread,
		CanSend:        true,
	}
}

func mapChatGroupInfo(row repository.ChatGroupInfoRow, memberRows []repository.ChatGroupMemberRow) dto.ChatGroupInfoDTO {
	var creator *dto.ChatMemberDTO
	if row.CreatorID != nil {
		roles := make([]string, 0)
		if row.CreatorRoles != nil && *row.CreatorRoles != "" {
			for _, role := range strings.Split(*row.CreatorRoles, ",") {
				role = strings.TrimSpace(role)
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		creator = &dto.ChatMemberDTO{
			UserID:   *row.CreatorID,
			FullName: valueOrEmpty(row.CreatorName),
			Email:    valueOrEmpty(row.CreatorEmail),
			Roles:    roles,
		}
	}

	admins := make([]dto.ChatGroupMemberDTO, 0)
	members := make([]dto.ChatGroupMemberDTO, 0, len(memberRows))
	for _, member := range memberRows {
		mapped := mapChatGroupMember(member)
		members = append(members, mapped)
		if member.Role == "admin" {
			admins = append(admins, mapped)
		}
	}

	return dto.ChatGroupInfoDTO{
		RoomID:      row.RoomID,
		RoomName:    row.RoomName,
		RoomType:    row.RoomType,
		SchoolID:    row.SchoolID,
		SchoolName:  row.SchoolName,
		Creator:     creator,
		Admins:      admins,
		Members:     members,
		CreatedAt:   formatChatTime(row.CreatedAt),
		MemberCount: row.ActiveMemberCount,
	}
}

func mapChatGroupMember(row repository.ChatGroupMemberRow) dto.ChatGroupMemberDTO {
	var leftAt *string
	if row.LeftAt != nil {
		value := formatChatTime(*row.LeftAt)
		leftAt = &value
	}
	return dto.ChatGroupMemberDTO{
		UserID:   row.UserID,
		FullName: row.FullName,
		Email:    row.Email,
		Role:     row.Role,
		JoinedAt: formatChatTime(row.JoinedAt),
		LeftAt:   leftAt,
	}
}

func mapChatMessageRow(row repository.ChatMessageRow, currentUserID string, attachments []repository.ChatAttachmentRow) dto.ChatMessageDTO {
	return dto.ChatMessageDTO{
		MessageID:   row.MessageID,
		RoomID:      row.RoomID,
		SenderID:    row.SenderID,
		SenderName:  row.SenderName,
		SenderRole:  row.SenderRole,
		Content:     row.Content,
		MessageType: row.Type,
		Attachments: mapChatAttachments(attachments),
		CreatedAt:   formatChatTime(row.CreatedAt),
		IsMine:      row.SenderID == currentUserID,
	}
}

func mapChatAttachments(rows []repository.ChatAttachmentRow) []dto.ChatAttachmentDTO {
	attachments := make([]dto.ChatAttachmentDTO, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, dto.ChatAttachmentDTO{
			AttachmentID: row.AttachmentID,
			MediaID:      row.MediaID,
			FileName:     row.FileName,
			MimeType:     row.MimeType,
			SizeBytes:    row.SizeBytes,
			URL:          row.URL,
		})
	}
	return attachments
}

func mapChatReadMembers(rows []repository.ChatReadMemberRow) []dto.ChatReadMemberDTO {
	members := make([]dto.ChatReadMemberDTO, 0, len(rows))
	for _, row := range rows {
		var lastReadAt *string
		if row.LastReadAt != nil {
			value := formatChatTime(*row.LastReadAt)
			lastReadAt = &value
		}
		members = append(members, dto.ChatReadMemberDTO{
			UserID:            row.UserID,
			FullName:          row.FullName,
			Email:             row.Email,
			LastReadMessageID: row.LastReadMessageID,
			LastReadAt:        lastReadAt,
		})
	}
	return members
}

func mapChatReadReceipt(row repository.ChatReadReceiptRow, userID string) *dto.ChatReadReceiptDTO {
	lastReadAt := ""
	if row.LastReadAt != nil {
		lastReadAt = formatChatTime(*row.LastReadAt)
	}
	return &dto.ChatReadReceiptDTO{
		RoomID:            row.RoomID,
		UserID:            userID,
		LastReadMessageID: row.LastReadMessageID,
		LastReadAt:        lastReadAt,
	}
}

func isSchoolChatRoom(room *repository.ChatRoomRow, schoolID string) bool {
	return room != nil &&
		room.RoomRefType != nil &&
		*room.RoomRefType == chatRefTypeSchool &&
		room.RoomRefID != nil &&
		*room.RoomRefID == schoolID
}

func isCustomGroupRoom(room *repository.ChatRoomRow) bool {
	return room != nil &&
		room.RoomType == chatRoomTypeGroup &&
		room.RoomRefType == nil &&
		room.RoomRefID == nil
}

func isDirectMessageRoom(room *repository.ChatRoomRow) bool {
	return room != nil &&
		room.RoomType == chatRoomTypeDM &&
		room.RoomRefType == nil &&
		room.RoomRefID == nil
}

func resolveRoomName(row repository.ChatRoomRow) string {
	if row.RoomType == chatRoomTypeDM {
		if row.DMTargetName != nil && strings.TrimSpace(*row.DMTargetName) != "" {
			return strings.TrimSpace(*row.DMTargetName)
		}
		if row.DMTargetEmail != nil && strings.TrimSpace(*row.DMTargetEmail) != "" {
			return strings.TrimSpace(*row.DMTargetEmail)
		}
		return "Pesan Langsung"
	}
	return row.RoomName
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatChatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
