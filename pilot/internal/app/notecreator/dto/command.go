package dto

import (
	"unicode"
	"unicode/utf8"

	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator/errors"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

const (
	maxTitleLen   = 48
	maxDescLen    = 2048
	maxTagNameLen = 255
	maxTagCount   = 10
)

type CreateNoteBasic struct {
	Title   string            `json:"title"`
	Desc    string            `json:"desc"`
	Privacy notevo.Visibility `json:"privacy"`
	Type    notevo.NoteType   `json:"type"`
}

func (c *CreateNoteBasic) Validate() error {
	if !c.Privacy.CheckValid() {
		return errors.ErrUnsupportedVisibility
	}
	if utf8.RuneCountInString(c.Title) > maxTitleLen {
		return errors.ErrWrongTitleLength
	}
	if utf8.RuneCountInString(c.Desc) > maxDescLen {
		return errors.ErrWrongDescLength
	}
	if !c.Type.CheckValid() {
		return errors.ErrWrongNoteType
	}

	return nil
}

type ImageForCreateNote struct {
	FileId string `json:"file_id"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

func (i *ImageForCreateNote) Validate() error {
	if i == nil {
		return errors.ErrNilImageParam
	}

	if i.FileId == "" {
		return errors.ErrEmptyImageFileId
	}

	if i.Width == 0 || i.Height == 0 || i.Format == "" {
		return errors.ErrEmptyImageInfo
	}

	if err := commondto.CheckImageFormat(i.Format); err != nil {
		return err
	}

	return nil
}

type ImageListForCreateNote []ImageForCreateNote

type VideoForCreateNote struct {
	FileId      string `json:"file_id"`
	CoverFileId string `json:"cover_file_id,optional"`
}

func (v *VideoForCreateNote) ValidateForCreate() error {
	if v == nil {
		return errors.ErrNilVideoParam
	}

	if v.FileId == "" {
		return errors.ErrEmptyVideoFileId
	}

	if v.CoverFileId == "" {
		return errors.ErrEmptyCoverFileId
	}

	return nil
}

func (v *VideoForCreateNote) ValidateForUpdate() error {
	if v == nil {
		return errors.ErrNilVideoParam
	}

	return nil
}

type TagId struct { // 必须再包一层 直接用数组无法解析
	Id notevo.TagId `json:"id"`
}

// 创作者创建笔记请求参数
type CreateNoteCommand struct {
	Basic   CreateNoteBasic        `json:"basic"`
	Images  ImageListForCreateNote `json:"images"`
	Video   *VideoForCreateNote    `json:"video,optional"`
	TagList []TagId                `json:"tag_list,optional"`
	AtUsers commondto.AtUserList   `json:"at_users,optional"`

	// internal usage only
	strictVideo bool `json:"-"`
}

func (c *CreateNoteCommand) ValidateBasic() error {
	if err := c.Basic.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *CreateNoteCommand) ValidateImages() error {
	for _, img := range c.Images {
		if err := img.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CreateNoteCommand) ValidateVideo() error {
	if c.Basic.Type == notevo.NoteTypeVideo {
		if c.strictVideo {
			return c.Video.ValidateForCreate()
		}
		return c.Video.ValidateForUpdate()
	}

	return nil
}

func (c *CreateNoteCommand) ValidateTags() error {
	if len(c.TagList) > maxTagCount {
		return errors.ErrTagCountExceed
	}
	return nil
}

func (c *CreateNoteCommand) Validate() error {
	if c == nil {
		return errors.ErrNilArg
	}
	c.strictVideo = true

	if err := c.ValidateBasic(); err != nil {
		return err
	}

	if err := c.ValidateImages(); err != nil {
		return err
	}

	if err := c.ValidateVideo(); err != nil {
		return err
	}

	if err := c.ValidateTags(); err != nil {
		return err
	}

	c.AtUsers = c.AtUsers.Filter()

	return nil
}

type CreateNoteResult struct {
	NoteId notevo.NoteId `json:"note_id"`
}

type UpdateNoteCommand struct {
	NoteId notevo.NoteId `json:"note_id"`
	CreateNoteCommand
}

func (c *UpdateNoteCommand) Validate() error {
	if c.NoteId <= 0 {
		return errors.ErrNoteNotFound
	}

	c.strictVideo = false

	if err := c.CreateNoteCommand.Validate(); err != nil {
		return err
	}

	return nil
}

type UpdateNoteResult struct {
	NoteId notevo.NoteId `json:"note_id"`
}

// 新增标签
type AddTagCommand struct {
	Name string `json:"name"`
}

func checkTagName(s string) error {
	if !utf8.ValidString(s) {
		return errors.ErrTagNameInvalid
	}

	for _, r := range s {
		if unicode.IsSpace(r) {
			return errors.ErrTagNameSpace
		}
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return errors.ErrTagNameSpecialChar
		}
	}

	return nil
}

func (r *AddTagCommand) Validate() error {
	if r == nil {
		return errors.ErrNilArg
	}

	if r.Name == "" {
		return errors.ErrTagNameInvalid
	}

	if l := utf8.RuneCountInString(r.Name); l > maxTagNameLen {
		return errors.ErrTagNameTooLong
	}

	if err := checkTagName(r.Name); err != nil {
		return err
	}

	return nil
}

// AddTagResult 添加标签响应
type AddTagResult struct {
	TagId notevo.TagId `json:"tag_id"`
}
