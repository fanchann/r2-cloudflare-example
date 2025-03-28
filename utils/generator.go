package utils

import "github.com/google/uuid"

func NewIDGen() string {
	u := uuid.New()
	return u.String()
}
