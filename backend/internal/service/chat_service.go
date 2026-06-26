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
	chatRoomTypeSchool  = "group"
	chatRefTypeSchool   = "school"
	chatMessageTypeText = "text"
	maxChatMessageLimit = 50
	maxChatContentLen   = 5000
	defaultSchoolRoom   = "Ruang sekolah"
)

type ChatService interface {
	ListMyRooms(userID string, schoolID string) ([]dto.ChatRoomDTO, error)
	OpenSchoolRoom(userID string, schoolID string) (*dto.ChatRoomDTO, error)
	ListMessages(userID string, schoolID string, roomID string, limit int, before *time.Time) (*dto.ChatMessagesResponseDTO, error)
	CreateMessage(userID string, schoolID string, roomID string, content string) (*dto.ChatMessageDTO, error)
	MarkRead(userID string, schoolID string, roomID string, lastReadMessageID *string) error
	CanAccessSchoolChat(userID string, schoolID string) (bool, error)
	CanAccessRoom(userID string, schoolID string, roomID string) (bool, *repository.ChatRoomRow, error)
}

type chatService struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &chatService{repo: repo}
}

func (s *chatService) ListMyRooms(userID string, schoolID string) ([]dto.ChatRoomDTO, error) {
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat school access denied")
	}

	rows, err := s.repo.ListSchoolRooms(userID, schoolID)
	if err != nil {
		return nil, err
	}

	rooms := make([]dto.ChatRoomDTO, 0, len(rows))
	for _, row := range rows {
		unread, err := s.repo.UnreadCount(row.RoomID, userID)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, mapChatRoomRow(row, unread))
	}
	return rooms, nil
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
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		room = &domain.ChatRoom{
			SchoolID:  schoolID,
			Name:      defaultSchoolRoom,
			Type:      chatRoomTypeSchool,
			RefType:   chatRefTypeSchool,
			RefID:     schoolID,
			CreatedBy: userID,
		}
		if err := s.repo.CreateSchoolRoom(room); err != nil {
			return nil, err
		}
		room, err = s.repo.GetSchoolRoom(schoolID)
		if err != nil {
			return nil, err
		}
	}

	context, err := s.repo.GetRoomContext(room.ID, schoolID)
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
	for _, row := range rows {
		messages = append(messages, mapChatMessageRow(row, userID))
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

func (s *chatService) CreateMessage(userID string, schoolID string, roomID string, content string) (*dto.ChatMessageDTO, error) {
	allowed, _, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: chat room access denied")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("chat message content is required")
	}
	if len([]rune(content)) > maxChatContentLen {
		return nil, fmt.Errorf("chat message content exceeds %d characters", maxChatContentLen)
	}

	message := domain.ChatMessage{
		RoomID:  roomID,
		UserID:  userID,
		Content: content,
		Type:    chatMessageTypeText,
	}
	if err := s.repo.CreateMessage(&message); err != nil {
		return nil, err
	}

	row, err := s.repo.GetMessageByID(message.ID, roomID)
	if err != nil {
		return nil, err
	}
	mapped := mapChatMessageRow(*row, userID)
	return &mapped, nil
}

func (s *chatService) MarkRead(userID string, schoolID string, roomID string, lastReadMessageID *string) error {
	allowed, _, err := s.CanAccessRoom(userID, schoolID, roomID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden: chat room access denied")
	}

	if lastReadMessageID != nil && *lastReadMessageID != "" {
		if _, err := s.repo.GetMessageByID(*lastReadMessageID, roomID); err != nil {
			return err
		}
	}
	return s.repo.UpsertReadReceipt(roomID, userID, lastReadMessageID)
}

func (s *chatService) CanAccessSchoolChat(userID string, schoolID string) (bool, error) {
	if userID == "" || schoolID == "" {
		return false, nil
	}
	return s.repo.UserIsActiveSchoolMember(userID, schoolID)
}

func (s *chatService) CanAccessRoom(userID string, schoolID string, roomID string) (bool, *repository.ChatRoomRow, error) {
	room, err := s.repo.GetRoomContext(roomID, schoolID)
	if err != nil {
		return false, nil, err
	}
	if room.RoomType != chatRoomTypeSchool || room.RoomRefType != chatRefTypeSchool || room.RoomRefID != schoolID || room.SchoolID != schoolID {
		return false, room, nil
	}
	allowed, err := s.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		return false, nil, err
	}
	return allowed, room, nil
}

func mapChatRoomRow(row repository.ChatRoomRow, unread int64) dto.ChatRoomDTO {
	var lastMessage *dto.ChatLastMessageDTO
	var lastMessageAt *string
	if row.LastMessageID != nil && row.LastContent != nil && row.LastMessageAt != nil {
		createdAt := formatChatTime(*row.LastMessageAt)
		lastMessageAt = &createdAt
		lastMessage = &dto.ChatLastMessageDTO{
			MessageID: *row.LastMessageID,
			Content:   *row.LastContent,
			CreatedAt: createdAt,
		}
		if row.LastSenderID != nil {
			lastMessage.SenderID = *row.LastSenderID
		}
		if row.LastSenderName != nil {
			lastMessage.SenderName = *row.LastSenderName
		}
	}

	return dto.ChatRoomDTO{
		RoomID:        row.RoomID,
		RoomName:      row.RoomName,
		RoomType:      row.RoomType,
		RoomRefType:   row.RoomRefType,
		RoomRefID:     row.RoomRefID,
		SchoolID:      row.SchoolID,
		SchoolName:    row.SchoolName,
		LastMessage:   lastMessage,
		LastMessageAt: lastMessageAt,
		UnreadCount:   unread,
		CanSend:       true,
	}
}

func mapChatMessageRow(row repository.ChatMessageRow, currentUserID string) dto.ChatMessageDTO {
	return dto.ChatMessageDTO{
		MessageID:   row.MessageID,
		RoomID:      row.RoomID,
		SenderID:    row.SenderID,
		SenderName:  row.SenderName,
		SenderRole:  row.SenderRole,
		Content:     row.Content,
		MessageType: row.Type,
		CreatedAt:   formatChatTime(row.CreatedAt),
		IsMine:      row.SenderID == currentUserID,
	}
}

func formatChatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
