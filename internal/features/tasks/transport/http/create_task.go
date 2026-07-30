package task_http

import (
	"net/http"

	"github.com/inraise/todo-app/internal/core/domain"
	"github.com/inraise/todo-app/internal/core/logger"
	core_request "github.com/inraise/todo-app/internal/core/transport/http/request"
	"github.com/inraise/todo-app/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	Title        string  `json:"title" validate:"required,min=1,max=100" example:"Домашнее задание"`
	Description  *string `json:"description" validate:"omitempty,min=1,max=1000" example:"Описание"`
	AuthorUserID int     `json:"author_user_id" validate:"required" example:"2"`
}

type CreateTaskResponse TaskDTOResponse

// CreateTask godoc
// @Summary Создать задачу
// @Description Создать новую задачу в системе
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CreateTaskRequest true "CreateTask тело запроса"
// @Success 201 {object} CreateTaskResponse "Успешно созданная задача"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 404 {object} response.ErrorResponse "Author not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tasks [post]
func (h *TasksHTTPHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(log, rw)

	var request CreateTaskRequest
	if err := core_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	taskDomain := domain.NewTaskUninitialized(
		request.Title,
		request.Description,
		request.AuthorUserID,
	)

	taskDomain, err := h.tasksService.CreateTask(ctx, taskDomain)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to create task",
		)

		return
	}

	response := CreateTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(
		response,
		http.StatusCreated,
	)
}
