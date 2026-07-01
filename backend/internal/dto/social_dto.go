package dto

// Feed DTOs
type CreateFeedDTO struct {
	SchoolID string   `json:"schoolId" binding:"required,uuid"`
	ClassID  string   `json:"classId" binding:"required,uuid"`
	Content  string   `json:"content" binding:"required"`
	MediaIDs []string `json:"mediaIds"`
}

type UpdateFeedDTO struct {
	Content  *string  `json:"content"`
	MediaIDs []string `json:"mediaIds"`
}

type FeedResponseDTO struct {
	ID           string             `json:"feedId"`
	Content      string             `json:"content"`
	CreatorName  string             `json:"creatorName,omitempty"`
	CreatedAt    string             `json:"createdAt"`
	Attachments  []MediaResponseDTO `json:"attachments,omitempty"`
	CommentCount int                `json:"commentCount"`
}

type CreateFeedResponseDTO struct {
	Message string           `json:"message"`
	Feed    *FeedResponseDTO `json:"feed,omitempty"`
}

type ClassWithFeedsDTO struct {
	Class ClassHeaderDTO    `json:"class"`
	Data  PaginatedResponse `json:"data"`
}

// Comment DTOs
type CreateCommentDTO struct {
	SchoolID   string `json:"schoolId"`
	SourceType string `json:"sourceType" binding:"required"`
	SourceID   string `json:"sourceId" binding:"required,uuid"`
	Content    string `json:"content" binding:"required"`
}

type UpdateCommentDTO struct {
	Content *string `json:"content"`
}

type CommentResponseDTO struct {
	ID          string `json:"commentId"`
	SourceType  string `json:"sourceType"`
	SourceID    string `json:"sourceId"`
	Content     string `json:"content"`
	CreatorName string `json:"creatorName"`
	CreatedAt   string `json:"createdAt"`
	IsMine      bool   `json:"isMine"`
}
