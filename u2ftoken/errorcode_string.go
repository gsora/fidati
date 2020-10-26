// Code generated by "stringer -type=ErrorCode"; DO NOT EDIT.

package u2ftoken

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NoError-36864]
	_ = x[ErrConditionNotSatisfied-27013]
	_ = x[ErrWrongData-27264]
	_ = x[ErrWrongLength-26368]
	_ = x[ErrClaNotSupported-28160]
	_ = x[ErrInsNotSupported-27904]
}

const (
	_ErrorCode_name_0 = "ErrWrongLength"
	_ErrorCode_name_1 = "ErrConditionNotSatisfied"
	_ErrorCode_name_2 = "ErrWrongData"
	_ErrorCode_name_3 = "ErrInsNotSupported"
	_ErrorCode_name_4 = "ErrClaNotSupported"
	_ErrorCode_name_5 = "NoError"
)

func (i ErrorCode) String() string {
	switch {
	case i == 26368:
		return _ErrorCode_name_0
	case i == 27013:
		return _ErrorCode_name_1
	case i == 27264:
		return _ErrorCode_name_2
	case i == 27904:
		return _ErrorCode_name_3
	case i == 28160:
		return _ErrorCode_name_4
	case i == 36864:
		return _ErrorCode_name_5
	default:
		return "ErrorCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
