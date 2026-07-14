package repository

import (
	"backend/internal/domain"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatRepository interface {
	ListSchoolRooms(userID string, schoolID string, search string) ([]ChatRoomRow, error)
	ListChatMembers(schoolID string, search string, excludeRoomID *string) ([]ChatMemberRow, error)
	GetSchoolRoom(schoolID string) (*domain.ChatRoom, error)
	CreateSchoolRoom(room *domain.ChatRoom) error
	GetRoomContext(roomID string, schoolID string, userID string) (*ChatRoomRow, error)
	FindDirectMessageRoom(schoolID string, userID string, targetUserID string) (*ChatRoomRow, error)
	CreateDirectMessageRoom(schoolID string, creatorID string, targetUserID string) (*ChatRoomRow, error)
	GetGroupInfo(roomID string, schoolID string) (*ChatGroupInfoRow, []ChatGroupMemberRow, error)
	UserIsActiveSchoolMember(userID string, schoolID string) (bool, error)
	UsersAreActiveSchoolMembers(userIDs []string, schoolID string) (map[string]bool, error)
	UserIsActiveRoomMember(userID string, roomID string) (bool, error)
	UserIsRoomAdmin(userID string, roomID string) (bool, error)
	CreateGroupRoomWithMembers(schoolID string, roomName string, creatorID string, memberUserIDs []string) (*domain.ChatRoom, error)
	UpdateGroupRoomName(roomID string, schoolID string, roomName string) error
	LeaveGroupRoom(roomID string, schoolID string, userID string) error
	AddGroupRoomMembers(roomID string, schoolID string, memberUserIDs []string) error
	RemoveGroupRoomMember(roomID string, schoolID string, targetUserID string) error
	ListMessages(roomID string, limit int, before *time.Time) ([]ChatMessageRow, error)
	CreateMessageWithAttachments(message *domain.ChatMessage, mediaIDs []string) error
	GetMessageByID(messageID string, roomID string) (*ChatMessageRow, error)
	ListMessageAttachments(messageIDs []string) (map[string][]ChatAttachmentRow, error)
	UpsertReadReceipt(roomID string, userID string, messageID *string) error
	GetReadReceipt(roomID string, userID string) (*ChatReadReceiptRow, error)
	ListSchoolReadMembers(roomID string, schoolID string) ([]ChatReadMemberRow, error)
	ListRoomReadMembers(roomID string, schoolID string) ([]ChatReadMemberRow, error)
	ListSchoolRecipientUserIDs(schoolID string) ([]string, error)
	ListRoomRecipientUserIDs(roomID string, schoolID string) ([]string, error)
	UnreadCount(roomID string, userID string) (int64, error)
	UnreadCounts(roomIDs []string, userID string) (map[string]int64, error)
}

type chatRepository struct {
	db *gorm.DB
}

type ChatRoomRow struct {
	RoomID                 string     `gorm:"column:room_id"`
	RoomName               string     `gorm:"column:room_name"`
	RoomType               string     `gorm:"column:room_type"`
	RoomRefType            *string    `gorm:"column:room_ref_type"`
	RoomRefID              *string    `gorm:"column:room_ref_id"`
	SchoolID               string     `gorm:"column:school_id"`
	SchoolName             string     `gorm:"column:school_name"`
	LastMessageID          *string    `gorm:"column:last_message_id"`
	LastSenderID           *string    `gorm:"column:last_sender_id"`
	LastSenderName         *string    `gorm:"column:last_sender_name"`
	LastContent            *string    `gorm:"column:last_content"`
	LastType               *string    `gorm:"column:last_type"`
	LastAttachmentCount    int        `gorm:"column:last_attachment_count"`
	LastAttachmentMimeType *string    `gorm:"column:last_attachment_mime_type"`
	LastAttachmentFileName *string    `gorm:"column:last_attachment_file_name"`
	LastMessageAt          *time.Time `gorm:"column:last_message_at"`
	DMTargetUserID         *string    `gorm:"column:dm_target_user_id"`
	DMTargetName           *string    `gorm:"column:dm_target_name"`
	DMTargetEmail          *string    `gorm:"column:dm_target_email"`
}

type ChatMemberRow struct {
	UserID   string `gorm:"column:user_id"`
	FullName string `gorm:"column:full_name"`
	Email    string `gorm:"column:email"`
	Roles    string `gorm:"column:roles"`
}

type ChatGroupInfoRow struct {
	RoomID            string    `gorm:"column:room_id"`
	RoomName          string    `gorm:"column:room_name"`
	RoomType          string    `gorm:"column:room_type"`
	SchoolID          string    `gorm:"column:school_id"`
	SchoolName        string    `gorm:"column:school_name"`
	CreatorID         *string   `gorm:"column:creator_id"`
	CreatorName       *string   `gorm:"column:creator_name"`
	CreatorEmail      *string   `gorm:"column:creator_email"`
	CreatorRoles      *string   `gorm:"column:creator_roles"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	ActiveMemberCount int       `gorm:"column:active_member_count"`
}

type ChatGroupMemberRow struct {
	UserID   string     `gorm:"column:user_id"`
	FullName string     `gorm:"column:full_name"`
	Email    string     `gorm:"column:email"`
	Role     string     `gorm:"column:role"`
	JoinedAt time.Time  `gorm:"column:joined_at"`
	LeftAt   *time.Time `gorm:"column:left_at"`
}

type ChatMessageRow struct {
	MessageID  string    `gorm:"column:message_id"`
	RoomID     string    `gorm:"column:room_id"`
	SenderID   string    `gorm:"column:sender_id"`
	SenderName string    `gorm:"column:sender_name"`
	SenderRole string    `gorm:"column:sender_role"`
	Content    string    `gorm:"column:content"`
	Type       string    `gorm:"column:message_type"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

type ChatAttachmentRow struct {
	MessageID    string `gorm:"column:message_id"`
	AttachmentID string `gorm:"column:attachment_id"`
	MediaID      string `gorm:"column:media_id"`
	FileName     string `gorm:"column:file_name"`
	MimeType     string `gorm:"column:mime_type"`
	SizeBytes    int64  `gorm:"column:size_bytes"`
	URL          string `gorm:"column:url"`
}

type ChatReadReceiptRow struct {
	RoomID            string     `gorm:"column:room_id"`
	LastReadMessageID *string    `gorm:"column:last_read_message_id"`
	LastReadAt        *time.Time `gorm:"column:last_read_at"`
}

type ChatReadMemberRow struct {
	UserID            string     `gorm:"column:user_id"`
	FullName          string     `gorm:"column:full_name"`
	Email             string     `gorm:"column:email"`
	LastReadMessageID *string    `gorm:"column:last_read_message_id"`
	LastReadAt        *time.Time `gorm:"column:last_read_at"`
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) ListSchoolRooms(userID string, schoolID string, search string) ([]ChatRoomRow, error) {
	var rows []ChatRoomRow
	searchPattern := "%" + search + "%"
	err := r.db.Raw(chatRoomListSelect()+`
		JOIN edv.school_users scu
			ON scu.scu_usr_id = ?
			AND scu.scu_sch_id = cr.room_sch_id
			AND scu.deleted_at IS NULL
		JOIN edv.users active_user
			ON active_user.usr_id = scu.scu_usr_id
			AND active_user.deleted_at IS NULL
		WHERE cr.room_sch_id = ?
			AND cr.room_type IN ('group', 'dm')
			AND cr.deleted_at IS NULL
			AND s.deleted_at IS NULL
			AND (
				? = ''
				OR cr.room_name ILIKE ?
				OR s.sch_name ILIKE ?
				OR COALESCE(dm_target.usr_nama_lengkap, '') ILIKE ?
				OR COALESCE(dm_target.usr_email, '') ILIKE ?
			)
			AND (
				(cr.room_ref_type = 'school' AND cr.room_ref_id = ?)
				OR (
					cr.room_type = 'group'
					AND cr.room_ref_type IS NULL
					AND cr.room_ref_id IS NULL
					AND EXISTS (
						SELECT 1
						FROM edv.chat_room_members crm
						WHERE crm.crm_room_id = cr.room_id
							AND crm.crm_usr_id = ?
							AND crm.left_at IS NULL
					)
				)
				OR (
					cr.room_type = 'dm'
					AND cr.room_ref_type IS NULL
					AND cr.room_ref_id IS NULL
					AND EXISTS (
						SELECT 1
						FROM edv.chat_room_members crm
						WHERE crm.crm_room_id = cr.room_id
							AND crm.crm_usr_id = ?
							AND crm.left_at IS NULL
					)
				)
			)
		ORDER BY COALESCE(lm.created_at, cr.created_at) DESC
	`, userID, userID, schoolID, search, searchPattern, searchPattern, searchPattern, searchPattern, schoolID, userID, userID).Scan(&rows).Error
	return rows, err
}

func (r *chatRepository) ListChatMembers(schoolID string, search string, excludeRoomID *string) ([]ChatMemberRow, error) {
	var rows []ChatMemberRow
	searchPattern := "%" + search + "%"
	err := r.db.Raw(`
		SELECT
			u.usr_id AS user_id,
			COALESCE(u.usr_nama_lengkap, '') AS full_name,
			u.usr_email AS email,
			COALESCE(string_agg(DISTINCT rol.rol_name, ',' ORDER BY rol.rol_name), '') AS roles
		FROM edv.school_users scu
		JOIN edv.users u
			ON u.usr_id = scu.scu_usr_id
			AND u.deleted_at IS NULL
		LEFT JOIN edv.user_roles ur ON ur.urol_scu_id = scu.scu_id
		LEFT JOIN edv.roles rol ON rol.rol_id = ur.urol_rol_id
		WHERE scu.scu_sch_id = ?
			AND scu.deleted_at IS NULL
			AND (
				? = ''
				OR u.usr_nama_lengkap ILIKE ?
				OR u.usr_email ILIKE ?
			)
			AND (
				?::uuid IS NULL
				OR NOT EXISTS (
					SELECT 1
					FROM edv.chat_room_members crm
					WHERE crm.crm_room_id = ?::uuid
						AND crm.crm_usr_id = u.usr_id
						AND crm.left_at IS NULL
				)
			)
		GROUP BY u.usr_id, u.usr_nama_lengkap, u.usr_email
		ORDER BY u.usr_nama_lengkap ASC, u.usr_email ASC
		LIMIT 50
	`, schoolID, search, searchPattern, searchPattern, excludeRoomID, excludeRoomID).Scan(&rows).Error
	return rows, err
}

func (r *chatRepository) GetSchoolRoom(schoolID string) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.db.Where("room_sch_id = ? AND room_type = ? AND room_ref_type = ? AND room_ref_id = ? AND deleted_at IS NULL", schoolID, "group", "school", schoolID).First(&room).Error
	return &room, err
}

func (r *chatRepository) CreateSchoolRoom(room *domain.ChatRoom) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "room_sch_id"}, {Name: "room_ref_type"}, {Name: "room_ref_id"}},
		DoNothing: true,
	}).Create(room).Error
}

func (r *chatRepository) GetRoomContext(roomID string, schoolID string, userID string) (*ChatRoomRow, error) {
	var row ChatRoomRow
	err := r.db.Raw(chatRoomListSelect()+`
		WHERE cr.room_id = ?
			AND cr.room_sch_id = ?
			AND cr.room_type IN ('group', 'dm')
			AND cr.deleted_at IS NULL
			AND s.deleted_at IS NULL
		LIMIT 1
	`, userID, roomID, schoolID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.RoomID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *chatRepository) FindDirectMessageRoom(schoolID string, userID string, targetUserID string) (*ChatRoomRow, error) {
	var row ChatRoomRow
	err := r.db.Raw(chatRoomListSelect()+`
		WHERE cr.room_sch_id = ?
			AND cr.room_type = 'dm'
			AND cr.room_ref_type IS NULL
			AND cr.room_ref_id IS NULL
			AND cr.deleted_at IS NULL
			AND s.deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM edv.chat_room_members me
				WHERE me.crm_room_id = cr.room_id
					AND me.crm_usr_id = ?
					AND me.left_at IS NULL
			)
			AND EXISTS (
				SELECT 1
				FROM edv.chat_room_members target
				WHERE target.crm_room_id = cr.room_id
					AND target.crm_usr_id = ?
					AND target.left_at IS NULL
			)
			AND (
				SELECT COUNT(*)
				FROM edv.chat_room_members active_members
				WHERE active_members.crm_room_id = cr.room_id
					AND active_members.left_at IS NULL
			) = 2
		ORDER BY COALESCE(lm.created_at, cr.created_at) DESC, cr.created_at ASC
		LIMIT 1
	`, userID, schoolID, userID, targetUserID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.RoomID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *chatRepository) CreateDirectMessageRoom(schoolID string, creatorID string, targetUserID string) (*ChatRoomRow, error) {
	var created domain.ChatRoom
	err := r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var existing ChatRoomRow
		if err := tx.Raw(chatRoomListSelect()+`
			WHERE cr.room_sch_id = ?
				AND cr.room_type = 'dm'
				AND cr.room_ref_type IS NULL
				AND cr.room_ref_id IS NULL
				AND cr.deleted_at IS NULL
				AND s.deleted_at IS NULL
				AND EXISTS (
					SELECT 1
					FROM edv.chat_room_members me
					WHERE me.crm_room_id = cr.room_id
						AND me.crm_usr_id = ?
						AND me.left_at IS NULL
				)
				AND EXISTS (
					SELECT 1
					FROM edv.chat_room_members target
					WHERE target.crm_room_id = cr.room_id
						AND target.crm_usr_id = ?
						AND target.left_at IS NULL
				)
				AND (
					SELECT COUNT(*)
					FROM edv.chat_room_members active_members
					WHERE active_members.crm_room_id = cr.room_id
						AND active_members.left_at IS NULL
				) = 2
			ORDER BY COALESCE(lm.created_at, cr.created_at) DESC, cr.created_at ASC
			LIMIT 1
		`, creatorID, schoolID, creatorID, targetUserID).Scan(&existing).Error; err != nil {
			return err
		}
		if existing.RoomID != "" {
			created.ID = existing.RoomID
			return nil
		}

		if err := tx.Raw(`
			INSERT INTO edv.chat_rooms (room_sch_id, room_name, room_type, room_ref_type, room_ref_id, created_by, created_at)
			VALUES (?, '', 'dm', NULL, NULL, ?, ?)
			RETURNING room_id, room_sch_id, room_name, room_type, created_by, created_at
		`, schoolID, creatorID, now).Scan(&created).Error; err != nil {
			return err
		}

		members := []domain.ChatRoomMember{
			{RoomID: created.ID, UserID: creatorID, Role: "member", JoinedAt: now},
			{RoomID: created.ID, UserID: targetUserID, Role: "member", JoinedAt: now},
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "crm_room_id"}, {Name: "crm_usr_id"}},
			DoNothing: true,
		}).Create(&members).Error
	})
	if err != nil {
		return nil, err
	}

	return r.FindDirectMessageRoom(schoolID, creatorID, targetUserID)
}

func (r *chatRepository) GetGroupInfo(roomID string, schoolID string) (*ChatGroupInfoRow, []ChatGroupMemberRow, error) {
	var info ChatGroupInfoRow
	err := r.db.Raw(`
		SELECT
			cr.room_id,
			cr.room_name,
			cr.room_type,
			s.sch_id AS school_id,
			s.sch_name AS school_name,
			creator.usr_id AS creator_id,
			creator.usr_nama_lengkap AS creator_name,
			creator.usr_email AS creator_email,
			creator_roles.roles AS creator_roles,
			cr.created_at,
			COUNT(active_members.crm_id)::int AS active_member_count
		FROM edv.chat_rooms cr
		JOIN edv.schools s
			ON s.sch_id = cr.room_sch_id
			AND s.deleted_at IS NULL
		LEFT JOIN edv.users creator
			ON creator.usr_id = cr.created_by
			AND creator.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT COALESCE(string_agg(DISTINCT rol.rol_name, ',' ORDER BY rol.rol_name), '') AS roles
			FROM edv.school_users scu
			LEFT JOIN edv.user_roles ur ON ur.urol_scu_id = scu.scu_id
			LEFT JOIN edv.roles rol ON rol.rol_id = ur.urol_rol_id
			WHERE scu.scu_usr_id = creator.usr_id
				AND scu.scu_sch_id = cr.room_sch_id
				AND scu.deleted_at IS NULL
		) creator_roles ON true
		LEFT JOIN edv.chat_room_members active_members
			ON active_members.crm_room_id = cr.room_id
			AND active_members.left_at IS NULL
		WHERE cr.room_id = ?
			AND cr.room_sch_id = ?
			AND cr.room_type = 'group'
			AND cr.room_ref_type IS NULL
			AND cr.room_ref_id IS NULL
			AND cr.deleted_at IS NULL
		GROUP BY cr.room_id, cr.room_name, cr.room_type, s.sch_id, s.sch_name, creator.usr_id, creator.usr_nama_lengkap, creator.usr_email, creator_roles.roles, cr.created_at
		LIMIT 1
	`, roomID, schoolID).Scan(&info).Error
	if err != nil {
		return nil, nil, err
	}
	if info.RoomID == "" {
		return nil, nil, gorm.ErrRecordNotFound
	}

	var members []ChatGroupMemberRow
	err = r.db.Raw(`
		SELECT
			u.usr_id AS user_id,
			COALESCE(u.usr_nama_lengkap, '') AS full_name,
			u.usr_email AS email,
			crm.crm_role AS role,
			crm.joined_at,
			crm.left_at
		FROM edv.chat_room_members crm
		JOIN edv.users u
			ON u.usr_id = crm.crm_usr_id
			AND u.deleted_at IS NULL
		WHERE crm.crm_room_id = ?
			AND crm.left_at IS NULL
		ORDER BY CASE crm.crm_role WHEN 'admin' THEN 1 ELSE 2 END, crm.joined_at ASC
	`, roomID).Scan(&members).Error
	if err != nil {
		return nil, nil, err
	}
	return &info, members, nil
}

func (r *chatRepository) UserIsActiveSchoolMember(userID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.school_users AS scu").
		Joins("JOIN edv.users u ON u.usr_id = scu.scu_usr_id AND u.deleted_at IS NULL").
		Where("scu.scu_usr_id = ? AND scu.scu_sch_id = ? AND scu.deleted_at IS NULL", userID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) UsersAreActiveSchoolMembers(userIDs []string, schoolID string) (map[string]bool, error) {
	result := make(map[string]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		UserID string `gorm:"column:user_id"`
	}
	err := r.db.Raw(`
		SELECT DISTINCT scu.scu_usr_id AS user_id
		FROM edv.school_users scu
		JOIN edv.users u
			ON u.usr_id = scu.scu_usr_id
			AND u.deleted_at IS NULL
		WHERE scu.scu_sch_id = ?
			AND scu.deleted_at IS NULL
			AND scu.scu_usr_id IN ?
	`, schoolID, userIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserID] = true
	}
	return result, nil
}

func (r *chatRepository) UserIsActiveRoomMember(userID string, roomID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.chat_room_members AS crm").
		Joins("JOIN edv.users u ON u.usr_id = crm.crm_usr_id AND u.deleted_at IS NULL").
		Where("crm.crm_room_id = ? AND crm.crm_usr_id = ? AND crm.left_at IS NULL", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) UserIsRoomAdmin(userID string, roomID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.chat_room_members AS crm").
		Joins("JOIN edv.users u ON u.usr_id = crm.crm_usr_id AND u.deleted_at IS NULL").
		Where("crm.crm_room_id = ? AND crm.crm_usr_id = ? AND crm.crm_role = 'admin' AND crm.left_at IS NULL", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) CreateGroupRoomWithMembers(schoolID string, roomName string, creatorID string, memberUserIDs []string) (*domain.ChatRoom, error) {
	var created domain.ChatRoom
	err := r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Raw(`
			INSERT INTO edv.chat_rooms (room_sch_id, room_name, room_type, room_ref_type, room_ref_id, created_by, created_at)
			VALUES (?, ?, 'group', NULL, NULL, ?, ?)
			RETURNING room_id, room_sch_id, room_name, room_type, created_by, created_at
		`, schoolID, roomName, creatorID, now).Scan(&created).Error; err != nil {
			return err
		}

		members := make([]domain.ChatRoomMember, 0, len(memberUserIDs))
		for _, memberID := range memberUserIDs {
			role := "member"
			if memberID == creatorID {
				role = "admin"
			}
			members = append(members, domain.ChatRoomMember{
				RoomID:   created.ID,
				UserID:   memberID,
				Role:     role,
				JoinedAt: now,
			})
		}
		if len(members) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "crm_room_id"}, {Name: "crm_usr_id"}},
			DoNothing: true,
		}).Create(&members).Error
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *chatRepository) UpdateGroupRoomName(roomID string, schoolID string, roomName string) error {
	result := r.db.Exec(`
		UPDATE edv.chat_rooms
		SET room_name = ?
		WHERE room_id = ?
			AND room_sch_id = ?
			AND room_type = 'group'
			AND room_ref_type IS NULL
			AND room_ref_id IS NULL
			AND deleted_at IS NULL
	`, roomName, roomID, schoolID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *chatRepository) LeaveGroupRoom(roomID string, schoolID string, userID string) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		var room struct {
			RoomID    string `gorm:"column:room_id"`
			CreatedBy string `gorm:"column:created_by"`
		}
		if err := tx.Raw(`
			SELECT room_id, created_by
			FROM edv.chat_rooms
			WHERE room_id = ?
				AND room_sch_id = ?
				AND room_type = 'group'
				AND room_ref_type IS NULL
				AND room_ref_id IS NULL
				AND deleted_at IS NULL
			FOR UPDATE
		`, roomID, schoolID).Scan(&room).Error; err != nil {
			return err
		}
		if room.RoomID == "" {
			return gorm.ErrRecordNotFound
		}

		result := tx.Exec(`
			UPDATE edv.chat_room_members
			SET left_at = ?
			WHERE crm_room_id = ?
				AND crm_usr_id = ?
				AND left_at IS NULL
		`, now, roomID, userID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var activeCount int64
		if err := tx.Table("edv.chat_room_members").
			Where("crm_room_id = ? AND left_at IS NULL", roomID).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount == 0 {
			return tx.Exec("UPDATE edv.chat_rooms SET deleted_at = ? WHERE room_id = ?", now, roomID).Error
		}

		if room.CreatedBy == userID {
			return transferGroupOwnership(tx, roomID)
		}
		return nil
	})
}

func (r *chatRepository) AddGroupRoomMembers(roomID string, schoolID string, memberUserIDs []string) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		var roomIDFound string
		if err := tx.Raw(`
			SELECT room_id
			FROM edv.chat_rooms
			WHERE room_id = ?
				AND room_sch_id = ?
				AND room_type = 'group'
				AND room_ref_type IS NULL
				AND room_ref_id IS NULL
				AND deleted_at IS NULL
			FOR UPDATE
		`, roomID, schoolID).Scan(&roomIDFound).Error; err != nil {
			return err
		}
		if roomIDFound == "" {
			return gorm.ErrRecordNotFound
		}

		for _, memberID := range memberUserIDs {
			var existing struct {
				UserID string     `gorm:"column:crm_usr_id"`
				LeftAt *time.Time `gorm:"column:left_at"`
			}
			if err := tx.Raw(`
				SELECT crm_usr_id, left_at
				FROM edv.chat_room_members
				WHERE crm_room_id = ?
					AND crm_usr_id = ?
				LIMIT 1
			`, roomID, memberID).Scan(&existing).Error; err != nil {
				return err
			}
			if existing.UserID != "" && existing.LeftAt == nil {
				return fmt.Errorf("chat group member already active")
			}
			if existing.UserID != "" {
				if err := tx.Exec(`
					UPDATE edv.chat_room_members
					SET left_at = NULL, crm_role = 'member'
					WHERE crm_room_id = ?
						AND crm_usr_id = ?
				`, roomID, memberID).Error; err != nil {
					return err
				}
				continue
			}
			member := domain.ChatRoomMember{
				RoomID:   roomID,
				UserID:   memberID,
				Role:     "member",
				JoinedAt: now,
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *chatRepository) RemoveGroupRoomMember(roomID string, schoolID string, targetUserID string) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		var room struct {
			RoomID    string `gorm:"column:room_id"`
			CreatedBy string `gorm:"column:created_by"`
		}
		if err := tx.Raw(`
			SELECT room_id, created_by
			FROM edv.chat_rooms
			WHERE room_id = ?
				AND room_sch_id = ?
				AND room_type = 'group'
				AND room_ref_type IS NULL
				AND room_ref_id IS NULL
				AND deleted_at IS NULL
			FOR UPDATE
		`, roomID, schoolID).Scan(&room).Error; err != nil {
			return err
		}
		if room.RoomID == "" {
			return gorm.ErrRecordNotFound
		}

		var target struct {
			UserID string `gorm:"column:crm_usr_id"`
			Role   string `gorm:"column:crm_role"`
		}
		if err := tx.Raw(`
			SELECT crm_usr_id, crm_role
			FROM edv.chat_room_members
			WHERE crm_room_id = ?
				AND crm_usr_id = ?
				AND left_at IS NULL
			LIMIT 1
		`, roomID, targetUserID).Scan(&target).Error; err != nil {
			return err
		}
		if target.UserID == "" {
			return gorm.ErrRecordNotFound
		}

		if target.Role == "admin" {
			var otherAdmins int64
			if err := tx.Table("edv.chat_room_members").
				Where("crm_room_id = ? AND crm_usr_id <> ? AND crm_role = 'admin' AND left_at IS NULL", roomID, targetUserID).
				Count(&otherAdmins).Error; err != nil {
				return err
			}
			if otherAdmins == 0 {
				if _, err := promoteOldestMember(tx, roomID, targetUserID); err != nil {
					return err
				}
			}
		}

		result := tx.Exec(`
			UPDATE edv.chat_room_members
			SET left_at = ?
			WHERE crm_room_id = ?
				AND crm_usr_id = ?
				AND left_at IS NULL
		`, now, roomID, targetUserID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if room.CreatedBy == targetUserID {
			return transferGroupOwnership(tx, roomID)
		}
		return nil
	})
}

func transferGroupOwnership(tx *gorm.DB, roomID string) error {
	var newOwnerID string
	if err := tx.Raw(`
		SELECT crm_usr_id
		FROM edv.chat_room_members
		WHERE crm_room_id = ?
			AND crm_role = 'admin'
			AND left_at IS NULL
		ORDER BY joined_at ASC
		LIMIT 1
	`, roomID).Scan(&newOwnerID).Error; err != nil {
		return err
	}
	if newOwnerID == "" {
		var err error
		newOwnerID, err = promoteOldestMember(tx, roomID, "")
		if err != nil {
			return err
		}
	}
	if newOwnerID == "" {
		return fmt.Errorf("chat group has no active member")
	}
	return tx.Exec("UPDATE edv.chat_rooms SET created_by = ? WHERE room_id = ?", newOwnerID, roomID).Error
}

func promoteOldestMember(tx *gorm.DB, roomID string, excludeUserID string) (string, error) {
	var newAdminID string
	if excludeUserID == "" {
		if err := tx.Raw(`
			SELECT crm_usr_id
			FROM edv.chat_room_members
			WHERE crm_room_id = ?
				AND left_at IS NULL
			ORDER BY joined_at ASC
			LIMIT 1
		`, roomID).Scan(&newAdminID).Error; err != nil {
			return "", err
		}
	} else {
		if err := tx.Raw(`
			SELECT crm_usr_id
			FROM edv.chat_room_members
			WHERE crm_room_id = ?
				AND left_at IS NULL
				AND crm_usr_id <> ?
			ORDER BY joined_at ASC
			LIMIT 1
		`, roomID, excludeUserID).Scan(&newAdminID).Error; err != nil {
			return "", err
		}
	}
	if newAdminID == "" {
		return "", nil
	}
	if err := tx.Exec(`
		UPDATE edv.chat_room_members
		SET crm_role = 'admin'
		WHERE crm_room_id = ?
			AND crm_usr_id = ?
			AND left_at IS NULL
	`, roomID, newAdminID).Error; err != nil {
		return "", err
	}
	return newAdminID, nil
}

func (r *chatRepository) ListMessages(roomID string, limit int, before *time.Time) ([]ChatMessageRow, error) {
	var rows []ChatMessageRow
	query := `
		SELECT *
		FROM (
			SELECT
				msg.msg_id AS message_id,
				msg.msg_room_id AS room_id,
				msg.msg_usr_id AS sender_id,
				COALESCE(u.usr_nama_lengkap, 'Pengguna') AS sender_name,
				COALESCE(sender_role.role_name, 'member') AS sender_role,
				msg.msg_content AS content,
				msg.msg_type AS message_type,
				msg.created_at AS created_at
			FROM edv.chat_messages msg
			JOIN edv.users u ON u.usr_id = msg.msg_usr_id AND u.deleted_at IS NULL
			LEFT JOIN LATERAL (
				SELECT r.rol_name AS role_name
				FROM edv.school_users scu
				JOIN edv.user_roles ur ON ur.urol_scu_id = scu.scu_id
				JOIN edv.roles r ON r.rol_id = ur.urol_rol_id
				JOIN edv.chat_rooms cr_role ON cr_role.room_id = msg.msg_room_id
				WHERE scu.scu_usr_id = msg.msg_usr_id
					AND scu.scu_sch_id = cr_role.room_sch_id
					AND scu.deleted_at IS NULL
					AND r.rol_name IN ('admin', 'teacher', 'student')
				ORDER BY CASE r.rol_name WHEN 'admin' THEN 1 WHEN 'teacher' THEN 2 WHEN 'student' THEN 3 ELSE 4 END
				LIMIT 1
			) sender_role ON true
			WHERE msg.msg_room_id = ?
				AND msg.msg_type IN ('text', 'file')
				AND msg.deleted_at IS NULL
				AND (?::timestamptz IS NULL OR msg.created_at < ?)
			ORDER BY msg.created_at DESC
			LIMIT ?
		) page
		ORDER BY created_at ASC
	`
	var beforeValue any
	if before != nil {
		beforeValue = *before
	}
	err := r.db.Raw(query, roomID, beforeValue, beforeValue, limit).Scan(&rows).Error
	return rows, err
}

func (r *chatRepository) CreateMessageWithAttachments(message *domain.ChatMessage, mediaIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, mediaID := range mediaIDs {
			attachment := domain.ChatAttachment{
				MessageID: message.ID,
				MediaID:   mediaID,
				CreatedAt: now,
			}
			if err := tx.Create(&attachment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *chatRepository) GetMessageByID(messageID string, roomID string) (*ChatMessageRow, error) {
	var row ChatMessageRow
	err := r.db.Raw(`
		SELECT
			msg.msg_id AS message_id,
			msg.msg_room_id AS room_id,
			msg.msg_usr_id AS sender_id,
			COALESCE(u.usr_nama_lengkap, 'Pengguna') AS sender_name,
			COALESCE(sender_role.role_name, 'member') AS sender_role,
			msg.msg_content AS content,
			msg.msg_type AS message_type,
			msg.created_at AS created_at
		FROM edv.chat_messages msg
		JOIN edv.users u ON u.usr_id = msg.msg_usr_id AND u.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT r.rol_name AS role_name
			FROM edv.school_users scu
			JOIN edv.user_roles ur ON ur.urol_scu_id = scu.scu_id
			JOIN edv.roles r ON r.rol_id = ur.urol_rol_id
			JOIN edv.chat_rooms cr_role ON cr_role.room_id = msg.msg_room_id
			WHERE scu.scu_usr_id = msg.msg_usr_id
				AND scu.scu_sch_id = cr_role.room_sch_id
				AND scu.deleted_at IS NULL
				AND r.rol_name IN ('admin', 'teacher', 'student')
			ORDER BY CASE r.rol_name WHEN 'admin' THEN 1 WHEN 'teacher' THEN 2 WHEN 'student' THEN 3 ELSE 4 END
			LIMIT 1
		) sender_role ON true
	WHERE msg.msg_id = ?
			AND msg.msg_room_id = ?
			AND msg.msg_type IN ('text', 'file')
			AND msg.deleted_at IS NULL
		LIMIT 1
	`, messageID, roomID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.MessageID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *chatRepository) ListMessageAttachments(messageIDs []string) (map[string][]ChatAttachmentRow, error) {
	result := make(map[string][]ChatAttachmentRow, len(messageIDs))
	if len(messageIDs) == 0 {
		return result, nil
	}

	var rows []ChatAttachmentRow
	err := r.db.Raw(`
		SELECT
			ca.cat_msg_id AS message_id,
			ca.cat_id AS attachment_id,
			m.med_id AS media_id,
			m.med_name AS file_name,
			m.med_mime_type AS mime_type,
			m.med_file_size AS size_bytes,
			m.med_file_url AS url
		FROM edv.chat_attachments ca
		JOIN edv.medias m
			ON m.med_id = ca.cat_med_id
			AND m.deleted_at IS NULL
		WHERE ca.cat_msg_id IN ?
		ORDER BY ca.created_at ASC
	`, messageIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.MessageID] = append(result[row.MessageID], row)
	}
	return result, nil
}

func (r *chatRepository) UpsertReadReceipt(roomID string, userID string, messageID *string) error {
	now := time.Now()
	receipt := domain.ChatReadReceipt{
		RoomID:            roomID,
		UserID:            userID,
		LastReadMessageID: messageID,
		LastReadAt:        now,
	}
	assignments := map[string]any{
		"last_read_at": now,
	}
	if messageID != nil && *messageID != "" {
		assignments["last_read_msg_id"] = messageID
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "rct_room_id"}, {Name: "rct_usr_id"}},
		DoUpdates: clause.Assignments(assignments),
	}).Create(&receipt).Error
}

func (r *chatRepository) GetReadReceipt(roomID string, userID string) (*ChatReadReceiptRow, error) {
	var row ChatReadReceiptRow
	err := r.db.Raw(`
		SELECT
			rct.rct_room_id AS room_id,
			rct.last_read_msg_id AS last_read_message_id,
			rct.last_read_at
		FROM edv.chat_read_receipts rct
		WHERE rct.rct_room_id = ?
			AND rct.rct_usr_id = ?
		LIMIT 1
	`, roomID, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.RoomID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *chatRepository) ListSchoolReadMembers(roomID string, schoolID string) ([]ChatReadMemberRow, error) {
	var rows []ChatReadMemberRow
	err := r.db.Raw(`
		SELECT
			u.usr_id AS user_id,
			COALESCE(u.usr_nama_lengkap, '') AS full_name,
			u.usr_email AS email,
			rct.last_read_msg_id AS last_read_message_id,
			rct.last_read_at
		FROM edv.school_users scu
		JOIN edv.users u
			ON u.usr_id = scu.scu_usr_id
			AND u.deleted_at IS NULL
		LEFT JOIN edv.chat_read_receipts rct
			ON rct.rct_room_id = ?
			AND rct.rct_usr_id = u.usr_id
		WHERE scu.scu_sch_id = ?
			AND scu.deleted_at IS NULL
		ORDER BY u.usr_nama_lengkap ASC, u.usr_email ASC
	`, roomID, schoolID).Scan(&rows).Error
	return rows, err
}

func (r *chatRepository) ListRoomReadMembers(roomID string, schoolID string) ([]ChatReadMemberRow, error) {
	var rows []ChatReadMemberRow
	err := r.db.Raw(`
		SELECT
			u.usr_id AS user_id,
			COALESCE(u.usr_nama_lengkap, '') AS full_name,
			u.usr_email AS email,
			rct.last_read_msg_id AS last_read_message_id,
			rct.last_read_at
		FROM edv.chat_room_members crm
		JOIN edv.chat_rooms cr
			ON cr.room_id = crm.crm_room_id
			AND cr.room_sch_id = ?
			AND cr.deleted_at IS NULL
		JOIN edv.school_users scu
			ON scu.scu_usr_id = crm.crm_usr_id
			AND scu.scu_sch_id = cr.room_sch_id
			AND scu.deleted_at IS NULL
		JOIN edv.users u
			ON u.usr_id = crm.crm_usr_id
			AND u.deleted_at IS NULL
		LEFT JOIN edv.chat_read_receipts rct
			ON rct.rct_room_id = crm.crm_room_id
			AND rct.rct_usr_id = u.usr_id
		WHERE crm.crm_room_id = ?
			AND crm.left_at IS NULL
		ORDER BY crm.joined_at ASC, u.usr_nama_lengkap ASC, u.usr_email ASC
	`, schoolID, roomID).Scan(&rows).Error
	return rows, err
}

func (r *chatRepository) ListSchoolRecipientUserIDs(schoolID string) ([]string, error) {
	var rows []struct {
		UserID string `gorm:"column:user_id"`
	}
	err := r.db.Raw(`
		SELECT DISTINCT u.usr_id AS user_id
		FROM edv.school_users scu
		JOIN edv.users u
			ON u.usr_id = scu.scu_usr_id
			AND u.deleted_at IS NULL
		WHERE scu.scu_sch_id = ?
			AND scu.deleted_at IS NULL
	`, schoolID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	return userIDs, nil
}

func (r *chatRepository) ListRoomRecipientUserIDs(roomID string, schoolID string) ([]string, error) {
	var rows []struct {
		UserID string `gorm:"column:user_id"`
	}
	err := r.db.Raw(`
		SELECT DISTINCT u.usr_id AS user_id
		FROM edv.chat_room_members crm
		JOIN edv.chat_rooms cr
			ON cr.room_id = crm.crm_room_id
			AND cr.room_sch_id = ?
			AND cr.deleted_at IS NULL
		JOIN edv.school_users scu
			ON scu.scu_usr_id = crm.crm_usr_id
			AND scu.scu_sch_id = cr.room_sch_id
			AND scu.deleted_at IS NULL
		JOIN edv.users u
			ON u.usr_id = crm.crm_usr_id
			AND u.deleted_at IS NULL
		WHERE crm.crm_room_id = ?
			AND crm.left_at IS NULL
	`, schoolID, roomID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	return userIDs, nil
}

func (r *chatRepository) UnreadCount(roomID string, userID string) (int64, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM edv.chat_messages msg
		LEFT JOIN edv.chat_read_receipts rct
			ON rct.rct_room_id = msg.msg_room_id
			AND rct.rct_usr_id = ?
		LEFT JOIN edv.chat_messages last_read_msg
			ON last_read_msg.msg_id = rct.last_read_msg_id
			AND last_read_msg.msg_room_id = msg.msg_room_id
			AND last_read_msg.deleted_at IS NULL
		WHERE msg.msg_room_id = ?
			AND msg.msg_usr_id <> ?
			AND msg.msg_type IN ('text', 'file')
			AND msg.deleted_at IS NULL
			AND (
				CASE
					WHEN last_read_msg.created_at IS NOT NULL AND rct.last_read_at IS NOT NULL
						THEN GREATEST(last_read_msg.created_at, rct.last_read_at)
					WHEN last_read_msg.created_at IS NOT NULL
						THEN last_read_msg.created_at
					ELSE rct.last_read_at
				END IS NULL
				OR msg.created_at > CASE
					WHEN last_read_msg.created_at IS NOT NULL AND rct.last_read_at IS NOT NULL
						THEN GREATEST(last_read_msg.created_at, rct.last_read_at)
					WHEN last_read_msg.created_at IS NOT NULL
						THEN last_read_msg.created_at
					ELSE rct.last_read_at
				END
			)
	`, userID, roomID, userID).Scan(&count).Error
	return count, err
}

func (r *chatRepository) UnreadCounts(roomIDs []string, userID string) (map[string]int64, error) {
	result := make(map[string]int64, len(roomIDs))
	if len(roomIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		RoomID string `gorm:"column:room_id"`
		Count  int64  `gorm:"column:count"`
	}
	err := r.db.Raw(`
		SELECT msg.msg_room_id AS room_id, COUNT(*) AS count
		FROM edv.chat_messages msg
		LEFT JOIN edv.chat_read_receipts rct
			ON rct.rct_room_id = msg.msg_room_id
			AND rct.rct_usr_id = ?
		LEFT JOIN edv.chat_messages last_read_msg
			ON last_read_msg.msg_id = rct.last_read_msg_id
			AND last_read_msg.msg_room_id = msg.msg_room_id
			AND last_read_msg.deleted_at IS NULL
		WHERE msg.msg_room_id IN ?
			AND msg.msg_usr_id <> ?
			AND msg.msg_type IN ('text', 'file')
			AND msg.deleted_at IS NULL
			AND (
				CASE
					WHEN last_read_msg.created_at IS NOT NULL AND rct.last_read_at IS NOT NULL
						THEN GREATEST(last_read_msg.created_at, rct.last_read_at)
					WHEN last_read_msg.created_at IS NOT NULL
						THEN last_read_msg.created_at
					ELSE rct.last_read_at
				END IS NULL
				OR msg.created_at > CASE
					WHEN last_read_msg.created_at IS NOT NULL AND rct.last_read_at IS NOT NULL
						THEN GREATEST(last_read_msg.created_at, rct.last_read_at)
					WHEN last_read_msg.created_at IS NOT NULL
						THEN last_read_msg.created_at
					ELSE rct.last_read_at
				END
			)
		GROUP BY msg.msg_room_id
	`, userID, roomIDs, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RoomID] = row.Count
	}
	return result, nil
}

func chatRoomListSelect() string {
	return chatRoomContextSelect() + `
		LEFT JOIN LATERAL (
			SELECT
				msg.msg_id,
				msg.msg_usr_id,
				msg.msg_content,
				msg.msg_type,
				msg.created_at,
				COUNT(ca.cat_id)::int AS attachment_count,
				MIN(m.med_mime_type) AS attachment_mime_type,
				MIN(m.med_name) AS attachment_file_name
			FROM edv.chat_messages msg
			LEFT JOIN edv.chat_attachments ca ON ca.cat_msg_id = msg.msg_id
			LEFT JOIN edv.medias m
				ON m.med_id = ca.cat_med_id
				AND m.deleted_at IS NULL
			WHERE msg.msg_room_id = cr.room_id
				AND msg.msg_type IN ('text', 'file')
				AND msg.deleted_at IS NULL
			GROUP BY msg.msg_id, msg.msg_usr_id, msg.msg_content, msg.msg_type, msg.created_at
			ORDER BY msg.created_at DESC
			LIMIT 1
		) lm ON true
		LEFT JOIN edv.users last_sender ON last_sender.usr_id = lm.msg_usr_id
	`
}

func chatRoomContextSelect() string {
	return `
		SELECT
			cr.room_id AS room_id,
			cr.room_name AS room_name,
			cr.room_type AS room_type,
			cr.room_ref_type AS room_ref_type,
			cr.room_ref_id AS room_ref_id,
			s.sch_id AS school_id,
			s.sch_name AS school_name,
			lm.msg_id AS last_message_id,
			lm.msg_usr_id AS last_sender_id,
			last_sender.usr_nama_lengkap AS last_sender_name,
			lm.msg_content AS last_content,
			lm.msg_type AS last_type,
			COALESCE(lm.attachment_count, 0) AS last_attachment_count,
			lm.attachment_mime_type AS last_attachment_mime_type,
			lm.attachment_file_name AS last_attachment_file_name,
			lm.created_at AS last_message_at,
			dm_target.usr_id AS dm_target_user_id,
			dm_target.usr_nama_lengkap AS dm_target_name,
			dm_target.usr_email AS dm_target_email
		FROM edv.chat_rooms cr
		JOIN edv.schools s ON s.sch_id = cr.room_sch_id
		LEFT JOIN LATERAL (
			SELECT u.usr_id, u.usr_nama_lengkap, u.usr_email
			FROM edv.chat_room_members crm
			JOIN edv.users u
				ON u.usr_id = crm.crm_usr_id
				AND u.deleted_at IS NULL
			WHERE crm.crm_room_id = cr.room_id
				AND crm.left_at IS NULL
				AND crm.crm_usr_id <> ?
			ORDER BY crm.joined_at ASC
			LIMIT 1
		) dm_target ON cr.room_type = 'dm'
	`
}
