package domain

import (
	"time"

	"github.com/google/uuid"
)

type PdvIntegracao struct {
	ID                    uuid.UUID  `json:"id"`
	EmpresaID             uuid.UUID  `json:"empresa_id"`
	LojaID                uuid.UUID  `json:"loja_id"`
	Provedor              string     `json:"provedor"`
	Ambiente              string     `json:"ambiente"`
	URLBase               string     `json:"url_base"`
	APIKeyEncrypted       string     `json:"-"` // Never expose in JSON
	ClientID              string     `json:"client_id,omitempty"`
	ClientSecretEncrypted string     `json:"-"` // Never expose in JSON
	Ativo                 bool       `json:"ativo"`
	UltimaSincronizacao   *time.Time `json:"ultima_sincronizacao,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty"`
}

type CreatePdvIntegracaoInput struct {
	EmpresaID    uuid.UUID `json:"empresa_id" validate:"required"`
	LojaID       uuid.UUID `json:"loja_id" validate:"required"`
	Provedor     string    `json:"provedor" validate:"required,oneof=gula"`
	Ambiente     string    `json:"ambiente" validate:"required,oneof=sandbox production"`
	URLBase      string    `json:"url_base" validate:"required,url"`
	APIKey       string    `json:"api_key" validate:"required"`
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
}

type UpdatePdvIntegracaoInput struct {
	Provedor     *string `json:"provedor,omitempty" validate:"omitempty,oneof=gula"`
	Ambiente     *string `json:"ambiente,omitempty" validate:"omitempty,oneof=sandbox production"`
	URLBase      *string `json:"url_base,omitempty" validate:"omitempty,url"`
	APIKey       *string `json:"api_key,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`
	Ativo        *bool   `json:"ativo,omitempty"`
}

type PdvIntegracaoFilter struct {
	EmpresaID *uuid.UUID
	LojaID    *uuid.UUID
	Provedor  *string
	Ativo     *bool
	Limit     int
	Offset    int
}

// PdvIntegracaoResponse is the safe response without secrets
type PdvIntegracaoResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmpresaID           uuid.UUID  `json:"empresa_id"`
	LojaID              uuid.UUID  `json:"loja_id"`
	Provedor            string     `json:"provedor"`
	Ambiente            string     `json:"ambiente"`
	URLBase             string     `json:"url_base"`
	ClientID            string     `json:"client_id,omitempty"`
	Ativo               bool       `json:"ativo"`
	UltimaSincronizacao *time.Time `json:"ultima_sincronizacao,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (p *PdvIntegracao) ToResponse() *PdvIntegracaoResponse {
	return &PdvIntegracaoResponse{
		ID:                  p.ID,
		EmpresaID:           p.EmpresaID,
		LojaID:              p.LojaID,
		Provedor:            p.Provedor,
		Ambiente:            p.Ambiente,
		URLBase:             p.URLBase,
		ClientID:            p.ClientID,
		Ativo:               p.Ativo,
		UltimaSincronizacao: p.UltimaSincronizacao,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}
