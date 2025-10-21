package user

import (
	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/util"
)

// updateUserFields updates the user entity with fields from the update request
func updateUserFields(user *entity.User, req request.UpdateUserReq) error {
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Location != "" {
		user.Location = req.Location
	}
	if req.PhotoURL != "" {
		user.PhotoURL = req.PhotoURL
	}
	if !req.BirthDate.IsZero() {
		user.BirthDate = req.BirthDate
	}

	if req.Password != "" {
		hashedPassword, err := util.HashPassword(req.Password, []byte(config.Env.Token.PasswordSalt))
		if err != nil {
			return err
		}
		user.PasswordHash = hashedPassword
	}

	return nil
}
