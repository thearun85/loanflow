package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Application struct {
	ID             string    `json:"id"`
	ApplicantName  string    `json:"applicant_name"`
	ApplicantEmail string    `json:"applicant_email"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	TermMonths     int       `json:"term_months"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type NewApplication struct {
	ApplicantName  string
	ApplicantEmail string
	AmountMinor    int64
	Currency       string
	TermMonths     int
}

const applicationColumns = `id::text, applicant_name, applicant_email, amount_minor,
	currency, term_months, status, created_at, updated_at`

type Store struct {
	pool *pgxpool.Pool
}

func (s *Store) CreateApplication(ctx context.Context, in NewApplication) (Application, error) {
	q := `INSERT INTO applications (applicant_name, applicant_email, amount_minor, currency, term_months)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + applicationColumns

	row := s.pool.QueryRow(ctx, q, in.ApplicantName, in.ApplicantEmail, in.AmountMinor, in.Currency, in.TermMonths)
	a, err := scanApplication(row)
	if err != nil {
		return Application{}, fmt.Errorf("insert application: %w", err)
	}
	return a, nil
}

func (s *Store) GetApplication(ctx context.Context, id pgtype.UUID) (Application, error) {
	q := `SELECT ` + applicationColumns + ` FROM applications WHERE id = $1`

	a, err := scanApplication(s.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, fmt.Errorf("get application: %w", err)
	}
	return a, nil
}

func scanApplication(row pgx.Row) (Application, error) {
	var a Application
	err := row.Scan(&a.ID, &a.ApplicantName, &a.ApplicantEmail, &a.AmountMinor,
		&a.Currency, &a.TermMonths, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
