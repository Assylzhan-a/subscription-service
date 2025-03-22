package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainErrors "github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/google/uuid"
)

type VoucherRepository struct {
	db *sql.DB
}

func NewVoucherRepository(db *sql.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

func (r *VoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	if voucher.ID == uuid.Nil {
		voucher.ID = uuid.New()
	}

	now := time.Now()
	voucher.CreatedAt = now
	voucher.UpdatedAt = now

	query := `
		INSERT INTO vouchers (
			id, code, discount_type, discount_value, product_id,
			is_active, expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Handle null product_id
	var productID interface{} = nil
	if voucher.ProductID != nil {
		productID = *voucher.ProductID
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		voucher.ID,
		voucher.Code,
		voucher.DiscountType,
		voucher.DiscountValue,
		productID,
		voucher.IsActive,
		voucher.ExpiresAt,
		voucher.CreatedAt,
		voucher.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation (code already exists)
		if isPgUniqueViolation(err) {
			return errors.New("voucher code already exists")
		}
		return err
	}

	return nil
}

func (r *VoucherRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	query := `
		SELECT 
			id, code, discount_type, discount_value, product_id,
			is_active, expires_at, created_at, updated_at
		FROM vouchers
		WHERE id = $1
	`

	return r.scanVoucher(ctx, query, id)
}

func (r *VoucherRepository) GetByCode(ctx context.Context, code string) (*models.Voucher, error) {
	query := `
		SELECT 
			id, code, discount_type, discount_value, product_id,
			is_active, expires_at, created_at, updated_at
		FROM vouchers
		WHERE code = $1
	`

	return r.scanVoucher(ctx, query, code)
}

func (r *VoucherRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Voucher, error) {
	query := `
		SELECT 
			id, code, discount_type, discount_value, product_id,
			is_active, expires_at, created_at, updated_at
		FROM vouchers
		WHERE product_id = $1 OR product_id IS NULL
		ORDER BY created_at DESC
	`

	return r.scanMultipleVouchers(ctx, query, productID)
}

func (r *VoucherRepository) GetAllActive(ctx context.Context) ([]*models.Voucher, error) {
	query := `
		SELECT 
			id, code, discount_type, discount_value, product_id,
			is_active, expires_at, created_at, updated_at
		FROM vouchers
		WHERE is_active = true AND expires_at > $1
		ORDER BY created_at DESC
	`

	return r.scanMultipleVouchers(ctx, query, time.Now())
}

func (r *VoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	voucher.UpdatedAt = time.Now()

	query := `
		UPDATE vouchers
		SET 
			code = $1, 
			discount_type = $2, 
			discount_value = $3, 
			product_id = $4, 
			is_active = $5, 
			expires_at = $6,
			updated_at = $7
		WHERE id = $8
	`

	var productID interface{} = nil
	if voucher.ProductID != nil {
		productID = *voucher.ProductID
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		voucher.Code,
		voucher.DiscountType,
		voucher.DiscountValue,
		productID,
		voucher.IsActive,
		voucher.ExpiresAt,
		voucher.UpdatedAt,
		voucher.ID,
	)

	if err != nil {
		// Check for unique constraint violation (code already exists)
		if isPgUniqueViolation(err) {
			return errors.New("voucher code already exists")
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrVoucherNotFound
	}

	return nil
}

func (r *VoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vouchers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domainErrors.ErrVoucherNotFound
	}

	return nil
}

func (r *VoucherRepository) scanVoucher(ctx context.Context, query string, args ...interface{}) (*models.Voucher, error) {
	voucher := &models.Voucher{}
	var productID sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&voucher.ID,
		&voucher.Code,
		&voucher.DiscountType,
		&voucher.DiscountValue,
		&productID,
		&voucher.IsActive,
		&voucher.ExpiresAt,
		&voucher.CreatedAt,
		&voucher.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrVoucherNotFound
		}
		return nil, err
	}

	if productID.Valid {
		uid, err := uuid.Parse(productID.String)
		if err != nil {
			return nil, err
		}
		voucher.ProductID = &uid
	}

	return voucher, nil
}

func (r *VoucherRepository) scanMultipleVouchers(ctx context.Context, query string, args ...interface{}) ([]*models.Voucher, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouchers []*models.Voucher

	for rows.Next() {
		voucher := &models.Voucher{}
		var productID sql.NullString

		err := rows.Scan(
			&voucher.ID,
			&voucher.Code,
			&voucher.DiscountType,
			&voucher.DiscountValue,
			&productID,
			&voucher.IsActive,
			&voucher.ExpiresAt,
			&voucher.CreatedAt,
			&voucher.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if productID.Valid {
			uid, err := uuid.Parse(productID.String)
			if err != nil {
				return nil, err
			}
			voucher.ProductID = &uid
		}

		vouchers = append(vouchers, voucher)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vouchers, nil
}
