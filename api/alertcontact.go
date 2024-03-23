package uptimerobot

import "fmt"

func AlertContactTypeToString(contactType int64) (string, error) {
	typeStr, ok := AlertContactTypes[contactType]
	if !ok {
		return "", fmt.Errorf("no alert contact type exists for type designator %d", contactType)
	}

	return typeStr, nil
}

func AlertContactStatusToString(status int64) (string, error) {
	statusStr, ok := AlertContactStatuses[status]
	if !ok {
		return "", fmt.Errorf("no alert contact type exists for type designator %d", status)
	}

	return statusStr, nil
}
