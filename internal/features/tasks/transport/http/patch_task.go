package task_http

import (
	"fmt"
	"net/http"

	"github.com/inraise/todo-app/internal/core/domain"
	"github.com/inraise/todo-app/internal/core/logger"
	core_request "github.com/inraise/todo-app/internal/core/transport/http/request"
	"github.com/inraise/todo-app/internal/core/transport/http/response"
	"github.com/inraise/todo-app/internal/core/transport/http/types"
)

type PatchTaskRequest struct {
	Title       types.Nullable[string] `json:"title"`
	Description types.Nullable[string] `json:"description"`
	Completed   types.Nullable[bool]   `json:"completed"`
}

type PatchTaskResponse TaskDTOResponse

func (r *PatchTaskRequest) Validate() error {
	if r.Title.Set {
		if r.Title.Value == nil {
			return fmt.Errorf("`Title` cannot be null")
		}

		titleLen := len([]rune(*r.Title.Value))
		if titleLen < 1 || titleLen > 100 {
			return fmt.Errorf("`Title` must be between 1 and 100 characters")
		}
	}

	if r.Description.Set {
		if r.Description.Value != nil {
			descLen := len([]rune(*r.Description.Value))
			if descLen < 1 || descLen > 1000 {
				return fmt.Errorf("`Description` must be between 1 and 1000 characters")
			}
		}
	}

	if r.Completed.Set {
		if r.Completed.Value == nil {
			return fmt.Errorf("`Completed` cannot be null")
		}
	}

	return nil
}

// PatchTasks godoc
// @Summary Изменить задачу
// @Description Изменить задачу из системы по его id
// @Description ### Логика обновления полей (Three-state logic):
// @Description 1. **Поле не передано**: `description` игнорируется, значение в бд не меняется
// @Description 2. **Явное значение**: `description` устанавливается новое значение
// @Description 3. **Значение null**: `description` удаляется, заменяясь на null
// @Description `title` и `completed` не может быть null
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "ID изменяемой задачи"
// @Param request body PatchTaskRequest true "PatchTask тело запроса"
// @Success 200 {object} PatchTaskResponse "Успешное изменение"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 404 {object} response.ErrorResponse "Task not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tasks/{id} [patch]
func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(log, rw)

	taskID, err := core_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get task ID from path",
		)

		return
	}

	var request PatchTaskRequest
	if err := core_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	taskPatch := taskPatchFromRequest(request)

	taskDomain, err := h.tasksService.PatchTask(ctx, taskID, taskPatch)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch task",
		)

		return
	}

	response := PatchTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(
		response,
		http.StatusOK,
	)
}

func taskPatchFromRequest(request PatchTaskRequest) domain.TaskPatch {
	return domain.NewTaskPatch(
		request.Title.ToDomain(),
		request.Description.ToDomain(),
		request.Completed.ToDomain(),
	)
}
