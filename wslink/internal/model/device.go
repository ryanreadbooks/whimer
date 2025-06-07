package model

type Device string

const (
	DeviceWeb Device = "web"
)

// implements encoding.BinaryMarshaler
func (d Device) MarshalBinary() ([]byte, error) {
	return []byte(d), nil
}
