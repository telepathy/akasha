package domain

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusActive    = "active"
	StatusArchived  = "archived"
	StatusDeleted   = "deleted"
	StatusCreating  = "creating"
)

type Dependency struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"index;size:128" json:"name"`
	GroupID  string         `gorm:"size:255" json:"groupId"`
	Artifact string         `gorm:"size:255" json:"artifact"`
	Version  string         `gorm:"size:64" json:"version"`
	Branch   string         `gorm:"index:idx_dep_branch_name,priority:1;size:64" json:"branch"`
	SourceIP string         `gorm:"size:64" json:"sourceIp"`
	Remark   string         `gorm:"type:text" json:"remark"`
	CreatedAt time.Time     `gorm:"index" json:"createdAt"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (Dependency) TableName() string {
	return "dependencies"
}

func (d *Dependency) MavenCoord() string {
	return d.GroupID + ":" + d.Artifact + ":" + d.Version
}

type Branch struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64" json:"name"`
	BaseBranch string   `gorm:"size:64" json:"baseBranch"`
	Status    string   `gorm:"size:32;default:active" json:"status"` // active, creating, deleting, deleted
	CreatedAt time.Time `json:"createdAt"`
}

func (Branch) TableName() string {
	return "branches"
}