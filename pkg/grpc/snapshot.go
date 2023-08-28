package grpc

import "strings"

type ProtoMessage interface {
	String() string
}

// StringifySnapshot converts a protobuf response to a string and removes all double spaces.
// This is needed because of a pseudo-random effect https://github.com/golang/protobuf/issues/1121
func StringifySnapshot(resp ProtoMessage) string {
	respStr := resp.String()
	respStr = strings.ReplaceAll(respStr, "  ", " ")

	return respStr
}
