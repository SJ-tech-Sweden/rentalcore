package repository

import (
	"errors"
	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/services/warehousecore"
	"net"
)

type CustomerRepository struct {
	db                    *Database
	warehouseClient       *warehousecore.Client
	useWarehouseCustomers bool
}

func NewCustomerRepository(db *Database) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// WithWarehouseCoreClient attaches a WarehouseCore client and toggles API-backed
// reads when enabled is true. Writes remain local to RentalCore.
func (r *CustomerRepository) WithWarehouseCoreClient(client *warehousecore.Client, enabled bool) *CustomerRepository {
	r.warehouseClient = client
	r.useWarehouseCustomers = enabled
	return r
}

func (r *CustomerRepository) Create(customer *models.Customer) error {
	result := r.db.Create(customer)
	return result.Error
}

func (r *CustomerRepository) GetByID(id uint) (*models.Customer, error) {
	// If API mode enabled, attempt to read from WarehouseCore and map to local model
	if r.useWarehouseCustomers && r.warehouseClient != nil {
		if c, err := r.warehouseClient.GetCustomer(id); err == nil {
			cust := &models.Customer{CustomerID: id}
			cust.CompanyName = c.CompanyName
			cust.FirstName = c.FirstName
			cust.LastName = c.LastName
			cust.Email = c.Email
			return cust, nil
		} else if !shouldFallbackToDBFromWarehouseCustomerError(err) {
			return nil, err
		}
		// Fall back to DB only on transient upstream errors.
	}

	var customer models.Customer
	err := r.db.First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func shouldFallbackToDBFromWarehouseCustomerError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, warehousecore.ErrCustomerNotFound) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return warehousecore.IsServerStatusError(err)
}

func (r *CustomerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

func (r *CustomerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Customer{}, id).Error
}

func (r *CustomerRepository) List(params *models.FilterParams) ([]models.Customer, error) {
	if params == nil {
		params = &models.FilterParams{}
	}

	// If API mode enabled, use WarehouseCore listing and map results
	if r.useWarehouseCustomers && r.warehouseClient != nil {
		items, err := r.warehouseClient.ListCustomers(params.SearchTerm)
		if err == nil {
			var out []models.Customer
			for _, it := range items {
				cust := models.Customer{CustomerID: it.ID}
				cust.CompanyName = it.CompanyName
				cust.FirstName = it.FirstName
				cust.LastName = it.LastName
				cust.Email = it.Email
				out = append(out, cust)
			}
			return out, nil
		}
		if !shouldFallbackToDBFromWarehouseCustomerError(err) {
			return nil, err
		}
		// Fall back to DB only on transient upstream errors.
	}

	var customers []models.Customer

	query := r.db.Model(&models.Customer{})

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("companyname LIKE ? OR firstname LIKE ? OR lastname LIKE ? OR email LIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("companyname ASC")

	err := query.Find(&customers).Error
	return customers, err
}
