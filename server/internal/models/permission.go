package models

import (
	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:255"`
	Code        string `json:"code" gorm:"size:100;uniqueIndex;not null"`
	Roles       []Role `json:"roles" gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
