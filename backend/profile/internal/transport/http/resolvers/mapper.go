package resolvers

import (
	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
	"github.com/pure-golang/monorepo/backend/profile/internal/transport/http/graphql/model"
)

func profileToModel(profile *domain.Profile) *model.Profile {
	if profile == nil {
		return nil
	}
	return &model.Profile{
		UserID:   profile.UserID,
		Nickname: profile.Nickname,
	}
}
