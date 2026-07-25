package models

// Avatars constants
const (
	avatarCacheTTL    = "avatar:ttl:"
	avatarCachePrefix = "avatar:"
	avatarPath        = "avatars/%s.jpg"
	defaultAvatar     = "/static/images/default-avatar.jpg"
)

// Whitelisted sorting options
const (
	Likes       = "likes"
	AvgRating   = "avg_rating"
	RatingCount = "rating_count"
)

// Form fieled types
const (
	FieldTypeInput FieldType = iota
	FieldTypeTextarea
)
