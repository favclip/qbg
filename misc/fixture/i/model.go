package i

import (
	"context"
	"go.mercari.io/datastore"
)

var _ datastore.PropertyTranslator = UserID(0)

type contextClient struct{}

const kindUser = "User"

type UserID int64

// +qbg
type User struct {
	ID       UserID `datastore:"-" boom:"id" json:"id"`
	Name     string `json:"name"`
	MentorID UserID `json:"mentorID"`
}

func (id UserID) ToPropertyValue(ctx context.Context) (interface{}, error) {
	client := ctx.Value(contextClient{}).(datastore.Client)
	key := client.IDKey(kindUser, int64(id), nil)
	return key, nil
}

func (id UserID) FromPropertyValue(ctx context.Context, p datastore.Property) (dst interface{}, err error) {
	key, ok := p.Value.(datastore.Key)
	if !ok {
		return nil, datastore.ErrInvalidEntityType
	}
	return UserID(key.ID()), nil
}
