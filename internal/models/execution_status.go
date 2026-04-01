package models

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ExecutionStatus int

const (
	ExecutionStatusUnknown ExecutionStatus = iota
	ExecutionStatusSuccess
	ExecutionStatusFailed
)

func (es ExecutionStatus) String() string {
	switch es {
	case ExecutionStatusSuccess:
		return "SUCCESS"
	case ExecutionStatusFailed:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

func ParseExecutionStatus(s string) (ExecutionStatus, error) {
	switch strings.ToUpper(s) {
	case "SUCCESS":
		return ExecutionStatusSuccess, nil
	case "FAIL":
		return ExecutionStatusFailed, nil
	default:
		return ExecutionStatusUnknown, fmt.Errorf("invalid run type: %s", s)
	}
}

func (rt ExecutionStatus) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: rt.String()}, nil
}

func (rt *ExecutionStatus) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	sv, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("ExecutionStatus must be a string in DynamoDB")
	}

	parsed, err := ParseExecutionStatus(sv.Value)
	if err != nil {
		return err
	}

	*rt = parsed
	return nil
}
