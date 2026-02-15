package postgres

type URL struct {
	Code        string `gorm:"primaryKey;column:code"`
	OriginalURL string `gorm:"column:original_url"`
	Hits        int64  `gorm:"column:hits"`
}

func (URL) TableName() string {
	return "urls"
}
