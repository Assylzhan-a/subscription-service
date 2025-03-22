package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainErrors "github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (
			id, name, description, price, duration_months, 
			tax_rate, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.DurationMonths,
		product.TaxRate,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*models.Product, error) {
	query := `
		SELECT 
			id, name, description, price, duration_months, 
			tax_rate, is_active, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product

	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.DurationMonths,
			&product.TaxRate,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, price, duration_months, 
			tax_rate, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &models.Product{}

	// Using decimal.Null to handle nullable decimals
	var price, taxRate decimal.Decimal

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&price,
		&product.DurationMonths,
		&taxRate,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrProductNotFound
		}
		return nil, err
	}

	product.Price = price
	product.TaxRate = taxRate

	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET 
			name = $1, 
			description = $2, 
			price = $3, 
			duration_months = $4, 
			tax_rate = $5, 
			is_active = $6, 
			updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.DurationMonths,
		product.TaxRate,
		product.IsActive,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrProductNotFound
	}

	return nil
}
