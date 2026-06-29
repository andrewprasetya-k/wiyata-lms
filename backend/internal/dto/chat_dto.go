package dto

type ChatLastMessageDTO struct {
	MessageID  string `json:"messageId"`
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
}

type ChatRoomDTO struct {
	RoomID         string              `json:"roomId"`
	RoomName       string              `json:"roomName"`
	RoomType       string              `json:"roomType"`
	RoomRefType    *string             `json:"roomRefType"`
	RoomRefID      *string             `json:"roomRefId"`
	SchoolID       string              `json:"schoolId"`
	SchoolName     string              `json:"schoolName"`
	DMTargetUserID *string             `json:"dmTargetUserId,omitempty"`
	DMTargetName   *string             `json:"dmTargetName,omitempty"`
	DMTargetEmail  *string             `json:"dmTargetEmail,omitempty"`
	LastMessage    *ChatLastMessageDTO `json:"lastMessage"`
	LastMessageAt  *string             `json:"lastMessageAt"`
	UnreadCount    int64               `json:"unreadCount"`
	CanSend        bool                `json:"canSend"`
}

type ChatMessageDTO struct {
	MessageID   string `json:"messageId"`
	RoomID      string `json:"roomId"`
	SenderID    string `json:"senderId"`
	SenderName  string `json:"senderName"`
	SenderRole  string `json:"senderRole"`
	Content     string `json:"content"`
	MessageType string `json:"messageType"`
	CreatedAt   string `json:"createdAt"`
	IsMine      bool   `json:"isMine"`
}

type ChatRoomsResponseDTO struct {
	Rooms []ChatRoomDTO `json:"rooms"`
}

type ChatRoomResponseDTO struct {
	Room ChatRoomDTO `json:"room"`
}

type ChatMemberDTO struct {
	UserID   string   `json:"userId"`
	FullName string   `json:"fullName"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

type ChatGroupMemberDTO struct {
	UserID   string  `json:"userId"`
	FullName string  `json:"fullName"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	JoinedAt string  `json:"joinedAt"`
	LeftAt   *string `json:"leftAt,omitempty"`
}

type ChatGroupInfoDTO struct {
	RoomID      string               `json:"roomId"`
	RoomName    string               `json:"roomName"`
	RoomType    string               `json:"roomType"`
	SchoolID    string               `json:"schoolId"`
	SchoolName  string               `json:"schoolName"`
	Creator     *ChatMemberDTO       `json:"creator"`
	Admins      []ChatGroupMemberDTO `json:"admins"`
	Members     []ChatGroupMemberDTO `json:"members"`
	CreatedAt   string               `json:"createdAt"`
	MemberCount int                  `json:"memberCount"`
}

type ChatMembersResponseDTO struct {
	Members []ChatMemberDTO `json:"members"`
}

type ChatGroupInfoResponseDTO struct {
	Group ChatGroupInfoDTO `json:"group"`
}

type ChatMessagesResponseDTO struct {
	Messages   []ChatMessageDTO `json:"messages"`
	NextBefore *string          `json:"nextBefore"`
	HasMore    bool             `json:"hasMore"`
}

type CreateChatMessageDTO struct {
	Content string `json:"content" binding:"required"`
}

type CreateChatGroupDTO struct {
	RoomName      string   `json:"roomName" binding:"required"`
	MemberUserIDs []string `json:"memberUserIds" binding:"required,dive,uuid"`
}

type OpenDirectMessageDTO struct {
	TargetUserID string `json:"targetUserId" binding:"required,uuid"`
}

type UpdateChatGroupDTO struct {
	RoomName string `json:"roomName" binding:"required"`
}

type AddChatGroupMembersDTO struct {
	MemberUserIDs []string `json:"memberUserIds" binding:"required,dive,uuid"`
}

type MarkChatRoomReadDTO struct {
	LastReadMessageID *string `json:"lastReadMessageId,omitempty" binding:"omitempty,uuid"`
}
