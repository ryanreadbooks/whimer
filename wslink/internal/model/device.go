package model

import pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"

type Device string

const (
	DeviceUnspec Device = "unspec"
	DeviceWeb    Device = "web"
)

func (d Device) Empty() bool {
	return d == ""
}

func (d Device) Unspec() bool {
	return d == DeviceUnspec
}

// implements encoding.BinaryMarshaler
func (d Device) MarshalBinary() ([]byte, error) {
	return []byte(d), nil
}

func GetDeviceFromPb(pbdv pushv1.Device) Device {
	switch pbdv {
	case pushv1.Device_DEVICE_UNSPECIFIED:
		return DeviceUnspec
	case pushv1.Device_DEVICE_WEB:
		return DeviceWeb
	}

	return ""
}
