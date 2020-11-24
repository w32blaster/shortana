package db

type (
	ShortURL struct {
		ShortUrl    string `storm:"id"` // primary key
		FullUrl     string
		Description string
		IsPublic    bool
	}
)
