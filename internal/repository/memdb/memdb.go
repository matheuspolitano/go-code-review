package memdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const COUPONS_DEFAULT_DATA_PATH = "coupons.data.json"

type Coupon struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Discount       int    `json:"discount"`
	MinBasketValue int    `json:"min_basket_value"`
}

// Repository defines the in-memory storage for Coupons.
// It implements the repository interface.
type Repository struct {
	entries  map[string]*Coupon
	filePath string
	mu       sync.RWMutex
}

// NewRepository creates and returns a new Repository instance.
// It loads existing coupons from filepath  if the file exists,
// It couponsDataPath is empty the default value will be 'coupons.data.json'
// otherwise, it initializes an empty repository and creates the file.
func NewRepository(couponsDataPath string) (*Repository, error) {
	if couponsDataPath == "" {
		couponsDataPath = COUPONS_DEFAULT_DATA_PATH
	}
	repo := &Repository{
		entries:  make(map[string]*Coupon),
		filePath: couponsDataPath, // Path to the data.json file in the root directory
	}

	// Load existing coupons from the file
	err := repo.loadFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load data from file: %w", err)
	}

	return repo, nil
}

// RepositoryInterface defines the methods that the Repository implements.
// Exported for external usage if needed.
type RepositoryInterface interface {
	FindByCode(string) (*Coupon, error)
	Save(*Coupon) error
}

// Custom errors for better error handling.
var (
	ErrCouponNotFound = errors.New("coupon not found")
	ErrInvalidCoupon  = errors.New("invalid coupon")
)

// FindByCode retrieves a Coupon by its code.
// It returns a copy of the Coupon to prevent external modifications.
func (r *Repository) FindByCode(code string) (Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coupon, exists := r.entries[code]
	if !exists {
		return Coupon{}, ErrCouponNotFound
	}

	// Return a copy to maintain immutability.
	return *coupon, nil
}

// loadFromFile reads the coupons from the data.json file into the repository.
// If the file does not exist, it initializes an empty repository and creates the file.
func (r *Repository) loadFromFile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	absPath, err := filepath.Abs(r.filePath)
	if err != nil {
		return fmt.Errorf("unable to determine absolute path: %w", err)
	}
	log.Printf("Loading data from '%s'", absPath)
	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist; create an empty new one
			file, createErr := os.Create(absPath)
			if createErr != nil {
				return fmt.Errorf("unable to create data file: %w", createErr)
			}
			file.Close()
			return nil
		}
		return fmt.Errorf("unable to open data file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var coupons []*Coupon
	if err := decoder.Decode(&coupons); err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		return fmt.Errorf("error decoding JSON from file: %w", err)
	}

	for _, coupon := range coupons {
		if coupon != nil && coupon.Code != "" {
			couponCopy := *coupon
			r.entries[coupon.Code] = &couponCopy
		}
	}

	return nil
}

// Save stores a Coupon in the repository.
// It returns an error if the coupon is nil or has an empty code.
func (r *Repository) Save(coupon *Coupon) error {
	if coupon == nil {
		return ErrInvalidCoupon
	}
	if coupon.Code == "" {
		return fmt.Errorf("%w: coupon code is empty", ErrInvalidCoupon)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to prevent external modifications affecting the repository.
	couponCopy := *coupon
	r.entries[coupon.Code] = &couponCopy

	// Persist the updated entries to the file
	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to save coupon to file: %w", err)
	}

	return nil
}

// saveToFile writes the current state of coupons to the data.json file.
// It directly writes to data.json without using a temporary file.
func (r *Repository) saveToFile() error {
	// Prepare a slice to hold coupons for JSON encoding
	coupons := make([]*Coupon, 0, len(r.entries))
	for _, coupon := range r.entries {
		coupons = append(coupons, coupon)
	}

	absPath, err := filepath.Abs(r.filePath)
	if err != nil {
		return fmt.Errorf("unable to determine absolute path: %w", err)
	}

	// Open the file with write permissions, create it if it doesn't exist, truncate it
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("unable to open data file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // For pretty-printing

	if err := encoder.Encode(coupons); err != nil {
		return fmt.Errorf("error encoding JSON to file: %w", err)
	}

	return nil
}
