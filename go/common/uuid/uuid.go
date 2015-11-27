package uuid

import (
	"github.com/satori/uuid"
)

func GetUid() string {
	return uuid.NewV4().String()
}
