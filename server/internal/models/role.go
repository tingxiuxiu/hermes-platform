package models

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name        string       `json:"name" gorm:"size:50;uniqueIndex;not null"`
	Description string       `json:"description" gorm:"size:255"`
	Users       []User       `json:"users" gorm:"many2many:user_roles;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
