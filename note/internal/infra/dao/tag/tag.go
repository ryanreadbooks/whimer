package tag

type Tag struct {
	Id    int64  `db:"id"`    // primary key
	Name  string `db:"name"`  // tag name
	Ctime int64  `db:"ctime"` // create time
}
