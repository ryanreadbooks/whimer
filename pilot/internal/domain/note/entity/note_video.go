package entity

type NoteVideo struct {
	FileId      string
	CoverFileId string

	targetFileId string
	metadata     *NoteVideoMetadata
}

func (v *NoteVideo) SetTargetFileId(fileId string) {
	v.targetFileId = fileId
}

func (v *NoteVideo) GetTargetFileId() string {
	return v.targetFileId
}

func (v *NoteVideo) SetMetadata(m *NoteVideoMetadata) {
	if v != nil {
		v.metadata = m
	}
}

func (v *NoteVideo) GetMetadata() *NoteVideoMetadata {
	if v != nil && v.metadata != nil {
		return v.metadata
	}
	return &NoteVideoMetadata{}
}

type NoteVideoMetadata struct {
	Width      uint32
	Height     uint32
	Format     string
	Duration   float64
	Bitrate    int64
	Codec      string
	Framerate  float64
	AudioCodec string
}
