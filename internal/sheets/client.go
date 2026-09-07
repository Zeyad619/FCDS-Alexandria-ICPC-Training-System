package sheets

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the Google Sheets API used by the training system.
type Client struct {
	Service *sheets.Service
}

// NewClient creates a Sheets client from a service-account credentials file.
// The credentials file must never be committed to the repository.
func NewClient(ctx context.Context, credentialsFile string) (*Client, error) {
	if credentialsFile == "" {
		return nil, fmt.Errorf("Google credentials file is required")
	}
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("create Google Sheets service: %w", err)
	}
	return &Client{Service: service}, nil
}

// Read reads a rectangular range and returns rows as strings.
func (c *Client) Read(ctx context.Context, spreadsheetID, readRange string) ([][]string, error) {
	if c == nil || c.Service == nil {
		return nil, fmt.Errorf("Google Sheets client is not initialized")
	}
	if spreadsheetID == "" || readRange == "" {
		return nil, fmt.Errorf("spreadsheet ID and range are required")
	}

	resp, err := c.Service.Spreadsheets.Values.Get(spreadsheetID, readRange).
		Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("read Google Sheet: %w", err)
	}

	rows := make([][]string, len(resp.Values))
	for i, row := range resp.Values {
		rows[i] = make([]string, len(row))
		for j, value := range row {
			rows[i][j] = fmt.Sprint(value)
		}
	}
	return rows, nil
}

// Write replaces the supplied range with the provided rows.
func (c *Client) Write(ctx context.Context, spreadsheetID, writeRange string, rows [][]string) error {
	if c == nil || c.Service == nil {
		return fmt.Errorf("Google Sheets client is not initialized")
	}
	if spreadsheetID == "" || writeRange == "" {
		return fmt.Errorf("spreadsheet ID and range are required")
	}

	values := make([][]interface{}, len(rows))
	for i, row := range rows {
		values[i] = make([]interface{}, len(row))
		for j, value := range row {
			values[i][j] = value
		}
	}

	body := &sheets.ValueRange{Values: values}
	_, err := c.Service.Spreadsheets.Values.Update(spreadsheetID, writeRange, body).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("write Google Sheet: %w", err)
	}
	return nil
}

// AssignmentRows converts trainer/trainee pairs into a sheet-friendly table.
func AssignmentRows(assignments map[string]string) [][]string {
	rows := [][]string{{"Trainee ID", "Trainer ID"}}
	for traineeID, trainerID := range assignments {
		rows = append(rows, []string{traineeID, trainerID})
	}
	return rows
}

// QuotaCell is a tiny helper for generating numeric quota values when mapping
// the assignment engine's result into a Google Sheet.
func QuotaCell(quota int) string { return strconv.Itoa(quota) }
