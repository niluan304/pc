package bemfa

const (
	MsgOn  = "on"
	MsgOff = "off"
)

// Topic
//
// 001	插座设备
//
// 002	灯泡设备
//
// 003	风扇设备
//
// 004	传感器
//
// 005	空调设备
//
// 006	开关设备
//
// 009	窗帘设备
type Topic interface {
	Handle(msg string) error
}
