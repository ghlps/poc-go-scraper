package models

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RunType int

const (
	RunTypeUnknown RunType = iota
	RunTypePrimary
	RunTypeBackup
	RunTypeCheckup
)

func (rt RunType) String() string {
	switch rt {
	case RunTypePrimary:
		return "PRIMARY"
	case RunTypeBackup:
		return "BACKUP"
	case RunTypeCheckup:
		return "CHECKUP"
	default:
		return "UNKNOWN"
	}
}

func ParseRunType(s string) (RunType, error) {
	switch strings.ToUpper(s) {
	case "PRIMARY":
		return RunTypePrimary, nil
	case "BACKUP":
		return RunTypeBackup, nil
	case "CHECKUP":
		return RunTypeCheckup, nil
	default:
		return RunTypeUnknown, fmt.Errorf("invalid run type: %s", s)
	}
}

func (rt RunType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: rt.String()}, nil
}

func (rt *RunType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	sv, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("RunType must be a string in DynamoDB")
	}

	parsed, err := ParseRunType(sv.Value)
	if err != nil {
		return err
	}

	*rt = parsed
	return nil
}
