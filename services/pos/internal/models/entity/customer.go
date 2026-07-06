package entity

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Customer merepresentasikan skema data pelanggan untuk segmentasi RFM di masa depan.
type Customer struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Phone      string             `bson:"phone" json:"phone"`
	TotalSpend int64              `bson:"total_spend" json:"total_spend"`
	VisitCount int64              `bson:"visit_count" json:"visit_count"`
	LastVisit  time.Time          `bson:"last_visit" json:"last_visit"`
	IsArchived bool               `bson:"is_archived" json:"is_archived"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
