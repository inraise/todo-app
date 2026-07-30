package filesystem

import (
	"fmt"
	"os"

	core_errors "github.com/inraise/todo-app/internal/core/errors"
)

func (r *WebRepository) GetFile(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"file: %s: %w",
				filePath,
				core_errors.ErrNotFound,
			)
		}

		return nil, fmt.Errorf(
			"get file: %s: %w",
			filePath,
			core_errors.ErrNotFound,
		)
	}

	return file, nil
}
