package data

import (
	"errors"
	"strings"
)

type FailureClass int

const (
	FailureClassNone FailureClass = iota
	FailureClassBusinessTerminal
	FailureClassTransientInfra
	FailureClassPoisonMessage
)

func (f FailureClass) String() string {
	switch f {
	case FailureClassNone:
		return "none"
	case FailureClassBusinessTerminal:
		return "business_terminal"
	case FailureClassTransientInfra:
		return "transient_infra"
	case FailureClassPoisonMessage:
		return "poison_message"
	default:
		return "unknown"
	}
}

func ClassifyError(err error) FailureClass {
	if err == nil {
		return FailureClassNone
	}

	errStr := err.Error()

	if isBusinessTerminalError(errStr) {
		return FailureClassBusinessTerminal
	}

	if isTransientInfraError(errStr) {
		return FailureClassTransientInfra
	}

	return FailureClassTransientInfra
}

func isBusinessTerminalError(errStr string) bool {
	terminalPatterns := []string{
		"stock not enough",
		"quota not enough",
		"duplicate seckill",
		"limit exceeded",
		"already seckilled",
		"库存不足",
		"额度不足",
		"重复秒杀",
		"已秒杀",
		"duplicate entry",
		"unique constraint",
		"sec_num",
	}
	for _, pattern := range terminalPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func isTransientInfraError(errStr string) bool {
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadline exceeded",
		"temporary failure",
		"too many connections",
		"redis",
		"mysql",
		"kafka",
		"dial error",
		"i/o timeout",
	}
	for _, pattern := range transientPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func IsPoisonMessage(msg []byte) bool {
	if len(msg) == 0 {
		return true
	}
	if len(msg) > 1024*1024 {
		return true
	}
	return false
}

func IsPoisonEnvelope(envelope *SeckillEnvelope) bool {
	if envelope == nil {
		return true
	}
	if envelope.Payload == nil {
		return true
	}
	if envelope.Payload.Goods == nil {
		return true
	}
	if envelope.Payload.SecNum == "" {
		return true
	}
	if envelope.Payload.UserID <= 0 {
		return true
	}
	return false
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	class := ClassifyError(err)
	return class == FailureClassTransientInfra
}

func IsTerminalError(err error) bool {
	if err == nil {
		return false
	}
	class := ClassifyError(err)
	return class == FailureClassBusinessTerminal
}

func WrapError(class FailureClass, msg string) error {
	return errors.New(class.String() + ": " + msg)
}
