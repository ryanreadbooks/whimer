package vo

type CommentImage struct {
	StoreKey string
	Width    uint32
	Height   uint32
	Format   string
}

type AtUser struct {
	Uid      int64
	Nickname string
}
