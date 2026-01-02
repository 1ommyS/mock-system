package domain

import "time"

type User struct {
	ID           string     `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	LastLoginAt  *time.Time `db:"last_login_at"`
}

type Role struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type Session struct {
	ID          string     `db:"id"`
	UserID      string     `db:"user_id"`
	RefreshHash string     `db:"refresh_hash"`
	CreatedAt   time.Time  `db:"created_at"`
	LastUsedAt  time.Time  `db:"last_used_at"`
	RevokedAt   *time.Time `db:"revoked_at"`
	IP          *string    `db:"ip"`
	UserAgent   *string    `db:"user_agent"`
}

type Resource struct {
	ResourceType string    `db:"resource_type"`
	ResourceID   string    `db:"resource_id"`
	OwnerUserID  string    `db:"owner_user_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type Grant struct {
	ID           string    `db:"id"`
	ResourceType string    `db:"resource_type"`
	ResourceID   string    `db:"resource_id"`
	UserID       string    `db:"user_id"`
	Permission   string    `db:"permission"`
	CreatedAt    time.Time `db:"created_at"`
}

type AuditLog struct {
	ID           string    `db:"id"`
	ActorUserID  *string   `db:"actor_user_id"`
	EventType    string    `db:"event_type"`
	ResourceType *string   `db:"resource_type"`
	ResourceID   *string   `db:"resource_id"`
	IP           *string   `db:"ip"`
	UserAgent    *string   `db:"user_agent"`
	Metadata     []byte    `db:"metadata"`
	CreatedAt    time.Time `db:"created_at"`
}

type ResourceAccess struct {
	ResourceID          string `db:"resource_id"`
	EffectivePermission string `db:"effective_permission"`
	IsOwner             bool   `db:"is_owner"`
}

type AccessCheck struct {
	EffectivePermission string `db:"effective_permission"`
	IsOwner             bool   `db:"is_owner"`
}
