package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/inraise/todo-app/internal/core/domain"
	"github.com/inraise/todo-app/internal/core/logger"
	core_request "github.com/inraise/todo-app/internal/core/transport/http/request"
	"github.com/inraise/todo-app/internal/core/transport/http/response"
	"github.com/inraise/todo-app/internal/core/transport/http/types"
)

type PatchUserRequest struct {
	FullName    types.Nullable[string] `json:"full_name" swaggertype:"string" example:"Ivan Ivanov"`
	PhoneNumber types.Nullable[string] `json:"phone_number" swaggertype:"string" example:"+79998882211"`
}

func (r *PatchUserRequest) Validate() error {
	if r.FullName.Set {
		if r.FullName.Value == nil {
			return fmt.Errorf("`FullName` can't be NULL")
		}

		fullNameLen := len([]rune(*r.FullName.Value))

		if fullNameLen < 3 || fullNameLen > 100 {
			return fmt.Errorf("`FullName` must be between 3 and 100 symbols")
		}
	}

	if r.PhoneNumber.Set {
		if r.PhoneNumber.Value != nil {
			phoneNumberLen := len([]rune(*r.PhoneNumber.Value))

			if phoneNumberLen < 10 || phoneNumberLen > 15 {
				return fmt.Errorf("`PhoneNumber` must be between 10 and 15 symbols")
			}

			if !strings.HasPrefix(*r.PhoneNumber.Value, "+") {
				return fmt.Errorf("`PhoneNumber` must have prefix '+'")
			}
		}
	}

	return nil
}

type PatchUserResponse UserDTOResponse

// PatchUser godoc
// @Summary Изменить пользователя
// @Description Изменить пользователя из системы по его id
// @Description ### Логика обновления полей (Three-state logic):
// @Description 1. **Поле не передано**: `phone_number` игнорируется, значение в бд не меняется
// @Description 2. **Явное значение**: `phone_number` устанавливается новое значение
// @Description 3. **Значение null**: `phone_number` удаляется, заменяясь на null
// @Description `full_name` не может быть null
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID изменяемого пользователя"
// @Param request body PatchUserRequest true "PatchUser тело запроса"
// @Success 200 {object} PatchUserResponse "Успешное изменение"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 409 {object} response.ErrorResponse "Conflict"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /users/{id} [patch]
func (h *UsersHTTPHandler) PatchUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(log, rw)

	userID, err := core_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID path value",
		)

		return
	}

	var request PatchUserRequest
	if err := core_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	userPatch := userPatchFromRequest(request)
	userDomain, err := h.usersService.PatchUser(ctx, userID, userPatch)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch user",
		)

		return
	}

	response := PatchUserResponse(userDTOFromDomain(userDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}

func userPatchFromRequest(request PatchUserRequest) domain.UserPatch {
	return domain.NewUserPatch(request.FullName.ToDomain(), request.PhoneNumber.ToDomain())
}
