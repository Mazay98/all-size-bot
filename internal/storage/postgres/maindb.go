package postgres

import (
	"context"
	"fmt"
	"sizebot/internal/entities"
)

// Commands returns all commands from the storage. Any error returned is internal.
func (s Storage) Commands(ctx context.Context) (entities.Commands, error) {
	rows, err := s.mainDB.Query(ctx, `
		SELECT 
			id,
			description,
			command,
			pattern,
			min_range,
			max_range
		FROM commands
		WHERE 
			is_active = true
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query commands: %w", err)
	}
	defer rows.Close()

	var command entities.Command
	commands := make(entities.Commands)

	for rows.Next() {
		if err := rows.Scan(
			&command.Id,
			&command.Description,
			&command.Command,
			&command.Pattern,
			&command.MinRange,
			&command.MaxRange,
		); err != nil {
			return nil, fmt.Errorf("failed to scan commands: %w", err)
		}

		commands[command.Command] = command
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan commands: %w", err)
	}

	return commands, nil
}
