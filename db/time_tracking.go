package db

import "fmt"

// Tag mirrors the tags table.
type Tag struct {
	ID             string  `db:"id"             json:"id"`
	OrganizationID string  `db:"organizationId" json:"organizationId"`
	Name           string  `db:"name"           json:"name"`
	Color          string  `db:"color"          json:"color"`
	CreatedAt      *string `db:"createdAt"      json:"createdAt"`
}

// CreateTagRequest is the payload for creating a tag.
type CreateTagRequest struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Color          string `json:"color"`
}

// UpdateTagRequest is the payload for updating a tag.
type UpdateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// TimeEntry mirrors the timeEntries table (with an optional joined clientName).
type TimeEntry struct {
	ID             string   `db:"id"             json:"id"`
	OrganizationID string   `db:"organizationId" json:"organizationId"`
	ClientID       *string  `db:"clientId"       json:"clientId"`
	Description    *string  `db:"description"    json:"description"`
	StartTime      int64    `db:"startTime"      json:"startTime"`
	EndTime        *int64   `db:"endTime"        json:"endTime"`
	Duration       int64    `db:"duration"       json:"duration"`
	Tags           *string  `db:"tags"           json:"tags"`
	IsBillable     int64    `db:"isBillable"     json:"isBillable"`
	HourlyRate     *float64 `db:"hourlyRate"     json:"hourlyRate"`
	CreatedAt      *string  `db:"createdAt"      json:"createdAt"`
	ClientName     *string  `db:"clientName"     json:"clientName"`
}

// CreateTimeEntryRequest is the payload for creating a time entry.
type CreateTimeEntryRequest struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	ClientID       *string  `json:"clientId"`
	Description    *string  `json:"description"`
	StartTime      int64    `json:"startTime"`
	EndTime        *int64   `json:"endTime"`
	Duration       int64    `json:"duration"`
	Tags           *string  `json:"tags"`
	IsBillable     int64    `json:"isBillable"`
	HourlyRate     *float64 `json:"hourlyRate"`
}

// UpdateTimeEntryRequest is the payload for updating a time entry.
type UpdateTimeEntryRequest struct {
	ClientID    *string  `json:"clientId"`
	Description *string  `json:"description"`
	StartTime   *int64   `json:"startTime"`
	EndTime     *int64   `json:"endTime"`
	Duration    *int64   `json:"duration"`
	Tags        *string  `json:"tags"`
	IsBillable  *int64   `json:"isBillable"`
	HourlyRate  *float64 `json:"hourlyRate"`
}

// ---- Tags ----

func (d *Database) GetTags(organizationID string) ([]Tag, error) {
	tags := []Tag{}
	err := d.DB.Select(&tags,
		`SELECT * FROM tags WHERE organizationId = ? ORDER BY name`,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("get_tags: %w", err)
	}
	return tags, nil
}

func (d *Database) GetTag(tagID string) (*Tag, error) {
	var tag Tag
	err := d.DB.Get(&tag, `SELECT * FROM tags WHERE id = ?`, tagID)
	if err != nil {
		return nil, fmt.Errorf("get_tag: %w", err)
	}
	return &tag, nil
}

func (d *Database) CreateTag(req CreateTagRequest) (*Tag, error) {
	_, err := d.DB.Exec(
		`INSERT INTO tags (id, organizationId, name, color) VALUES (?, ?, ?, ?)`,
		req.ID, req.OrganizationID, req.Name, req.Color,
	)
	if err != nil {
		return nil, fmt.Errorf("create_tag: %w", err)
	}
	return d.GetTag(req.ID)
}

func (d *Database) UpdateTag(tagID string, updates UpdateTagRequest) (*Tag, error) {
	_, err := d.DB.Exec(
		`UPDATE tags SET name = COALESCE(?, name), color = COALESCE(?, color) WHERE id = ?`,
		updates.Name, updates.Color, tagID,
	)
	if err != nil {
		return nil, fmt.Errorf("update_tag: %w", err)
	}
	return d.GetTag(tagID)
}

func (d *Database) DeleteTag(tagID string) (bool, error) {
	res, err := d.DB.Exec(`DELETE FROM tags WHERE id = ?`, tagID)
	if err != nil {
		return false, fmt.Errorf("delete_tag: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ---- Time Entries ----

func (d *Database) GetTimeEntries(organizationID string) ([]TimeEntry, error) {
	entries := []TimeEntry{}
	err := d.DB.Select(&entries, `
		SELECT t.*, c.name AS clientName
		FROM timeEntries t
		LEFT JOIN clients c ON t.clientId = c.id
		WHERE t.organizationId = ?
		ORDER BY t.startTime DESC`,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("get_time_entries: %w", err)
	}
	return entries, nil
}

func (d *Database) GetTimeEntry(timeEntryID string) (*TimeEntry, error) {
	var entry TimeEntry
	err := d.DB.Get(&entry, `
		SELECT t.*, c.name AS clientName
		FROM timeEntries t
		LEFT JOIN clients c ON t.clientId = c.id
		WHERE t.id = ?`,
		timeEntryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get_time_entry: %w", err)
	}
	return &entry, nil
}

func (d *Database) CreateTimeEntry(req CreateTimeEntryRequest) (*TimeEntry, error) {
	_, err := d.DB.Exec(`
		INSERT INTO timeEntries (id, organizationId, clientId, description, startTime, endTime, duration, tags, isBillable, hourlyRate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ID, req.OrganizationID, req.ClientID, req.Description,
		req.StartTime, req.EndTime, req.Duration, req.Tags,
		req.IsBillable, req.HourlyRate,
	)
	if err != nil {
		return nil, fmt.Errorf("create_time_entry: %w", err)
	}
	return d.GetTimeEntry(req.ID)
}

func (d *Database) UpdateTimeEntry(timeEntryID string, updates UpdateTimeEntryRequest) (*TimeEntry, error) {
	_, err := d.DB.Exec(`
		UPDATE timeEntries
		SET clientId    = COALESCE(?, clientId),
		    description = COALESCE(?, description),
		    startTime   = COALESCE(?, startTime),
		    endTime     = COALESCE(?, endTime),
		    duration    = COALESCE(?, duration),
		    tags        = COALESCE(?, tags),
		    isBillable  = COALESCE(?, isBillable),
		    hourlyRate  = COALESCE(?, hourlyRate)
		WHERE id = ?`,
		updates.ClientID, updates.Description, updates.StartTime, updates.EndTime,
		updates.Duration, updates.Tags, updates.IsBillable, updates.HourlyRate,
		timeEntryID,
	)
	if err != nil {
		return nil, fmt.Errorf("update_time_entry: %w", err)
	}
	return d.GetTimeEntry(timeEntryID)
}

func (d *Database) DeleteTimeEntry(timeEntryID string) (bool, error) {
	res, err := d.DB.Exec(`DELETE FROM timeEntries WHERE id = ?`, timeEntryID)
	if err != nil {
		return false, fmt.Errorf("delete_time_entry: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
