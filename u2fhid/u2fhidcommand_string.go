// Code generated by "stringer -type=u2fHIDCommand"; DO NOT EDIT.

package u2fhid

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[cmdPing-129]
	_ = x[cmdMsg-131]
	_ = x[cmdInit-134]
	_ = x[cmdError-191]
	_ = x[cmdLock-132]
	_ = x[cmdWink-136]
	_ = x[cmdSync-188]
}

const (
	_u2fHIDCommand_name_0 = "cmdPing"
	_u2fHIDCommand_name_1 = "cmdMsgcmdLock"
	_u2fHIDCommand_name_2 = "cmdInit"
	_u2fHIDCommand_name_3 = "cmdWink"
	_u2fHIDCommand_name_4 = "cmdSync"
	_u2fHIDCommand_name_5 = "cmdError"
)

var (
	_u2fHIDCommand_index_1 = [...]uint8{0, 6, 13}
)

func (i u2fHIDCommand) String() string {
	switch {
	case i == 129:
		return _u2fHIDCommand_name_0
	case 131 <= i && i <= 132:
		i -= 131
		return _u2fHIDCommand_name_1[_u2fHIDCommand_index_1[i]:_u2fHIDCommand_index_1[i+1]]
	case i == 134:
		return _u2fHIDCommand_name_2
	case i == 136:
		return _u2fHIDCommand_name_3
	case i == 188:
		return _u2fHIDCommand_name_4
	case i == 191:
		return _u2fHIDCommand_name_5
	default:
		return "u2fHIDCommand(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
