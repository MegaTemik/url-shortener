package service

import (
	"errors"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-playground/validator/v10"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
	ErrInvalidURL  = errors.New("invalid url")
)

type URLService struct {
	storage     URLStorage
	aliasLength int
}

type URLStorage interface {
	GetURL(alias string) (string, error)
	SaveURL(urlToSave string, alias string) error
	DeleteURL(alias string) (int64, error)
	UpdateURL(alias, newURL string) (int64, error)
}

type CreateURLInput struct {
	URL   string `validate:"required,url"`
	Alias string
}

type UpdateURLInput struct {
	Alias  string `validate:"required"`
	NewURL string `validate:"required,url"`
}

func NewURLService(storage URLStorage, aliasLength int) *URLService {
	return &URLService{storage: storage, aliasLength: aliasLength}
}

func (s *URLService) RegisterGetURL(alias string) (string, error) {
	url, err := s.storage.GetURL(alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return "", ErrURLNotFound
		}
	}
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLService) RegisterDeleteURL(alias string) (int64, error) {
	countDeleted, err := s.storage.DeleteURL(alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return 0, ErrURLNotFound
		}
	}
	if err != nil {
		return 0, err
	}
	return countDeleted, nil
}

func (s *URLService) RegisterSaveURL(inputReq CreateURLInput) (string, error) {
	if err := validator.New().Struct(inputReq); err != nil {
		return "", ErrInvalidURL
	}

	alias := inputReq.Alias
	if alias == "" {
		alias = random.NewRandomString(s.aliasLength)
	}

	err := s.storage.SaveURL(inputReq.URL, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			return "", ErrURLExists
		}
	}
	if err != nil {
		return "", err
	}
	return alias, nil
}

func (s *URLService) RegisterUpdateURL(inputReq UpdateURLInput) (int64, error) {
	if err := validator.New().Struct(inputReq); err != nil {
		return 0, ErrInvalidURL
	}

	alias := inputReq.Alias
	newURL := inputReq.NewURL

	countUpdated, err := s.storage.UpdateURL(alias, newURL)
	if errors.Is(err, storage.ErrURLNotFound) {
		return 0, ErrURLNotFound
	}

	if err != nil {
		return 0, err
	}

	return countUpdated, nil
}
