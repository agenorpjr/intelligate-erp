package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/intelligate/erp/backend/internal/adapter/gula"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrPdvIntegracaoNaoEncontrada = errors.New("integração PDV não encontrada")
	ErrPdvIntegracaoJaExiste      = errors.New("integração PDV já existe para esta loja e provedor")
)

type PdvIntegracaoUsecase struct {
	pdvIntegracaoRepo domain.PdvIntegracaoRepository
	lojaRepo          domain.LojaRepository
	pdvOrigemRepo     domain.PdvOrigemRepository
	salesUsecase      *SalesUsecase
	auditoriaRepo     domain.AuditoriaRepository
}

func NewPdvIntegracaoUsecase(
	pdvIntegracaoRepo domain.PdvIntegracaoRepository,
	lojaRepo domain.LojaRepository,
	pdvOrigemRepo domain.PdvOrigemRepository,
	salesUsecase *SalesUsecase,
	auditoriaRepo domain.AuditoriaRepository,
) *PdvIntegracaoUsecase {
	return &PdvIntegracaoUsecase{
		pdvIntegracaoRepo: pdvIntegracaoRepo,
		lojaRepo:          lojaRepo,
		pdvOrigemRepo:     pdvOrigemRepo,
		salesUsecase:      salesUsecase,
		auditoriaRepo:     auditoriaRepo,
	}
}

func (uc *PdvIntegracaoUsecase) Criar(ctx context.Context, input domain.CreatePdvIntegracaoInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.PdvIntegracaoResponse, error) {
	// Verificar se loja existe
	loja, err := uc.lojaRepo.GetByID(ctx, input.LojaID)
	if err != nil {
		return nil, err
	}
	if loja == nil {
		return nil, ErrLojaNaoEncontrada
	}

	// Verificar se já existe integração para esta loja e provedor
	existing, err := uc.pdvIntegracaoRepo.GetByLojaProvedor(ctx, input.LojaID, input.Provedor)
	if err == nil && existing != nil {
		return nil, ErrPdvIntegracaoJaExiste
	}

	// TODO: Criptografar API Key e Client Secret antes de salvar
	// Por enquanto, armazenamos como está (em produção usar criptografia)
	integracao := &domain.PdvIntegracao{
		ID:                    uuid.New(),
		EmpresaID:             input.EmpresaID,
		LojaID:                input.LojaID,
		Provedor:              input.Provedor,
		Ambiente:              input.Ambiente,
		URLBase:               input.URLBase,
		APIKeyEncrypted:       input.APIKey, // TODO: encrypt
		ClientID:              input.ClientID,
		ClientSecretEncrypted: input.ClientSecret, // TODO: encrypt
		Ativo:                 true,
	}

	if err := uc.pdvIntegracaoRepo.Create(ctx, integracao); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "pdv_integracao",
		EntidadeID: integracao.ID,
		Acao:       "create",
		Depois:     stringPtr(toJSON(integracao.ToResponse())),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return integracao.ToResponse(), nil
}

func (uc *PdvIntegracaoUsecase) Listar(ctx context.Context, empresaID uuid.UUID, filter domain.PdvIntegracaoFilter) ([]*domain.PdvIntegracaoResponse, error) {
	integracoes, err := uc.pdvIntegracaoRepo.ListByEmpresa(ctx, empresaID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]*domain.PdvIntegracaoResponse, len(integracoes))
	for i, integracao := range integracoes {
		responses[i] = integracao.ToResponse()
	}
	return responses, nil
}

func (uc *PdvIntegracaoUsecase) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.PdvIntegracaoResponse, error) {
	integracao, err := uc.pdvIntegracaoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if integracao == nil {
		return nil, ErrPdvIntegracaoNaoEncontrada
	}
	return integracao.ToResponse(), nil
}

func (uc *PdvIntegracaoUsecase) Atualizar(ctx context.Context, id uuid.UUID, input domain.UpdatePdvIntegracaoInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.PdvIntegracaoResponse, error) {
	integracao, err := uc.pdvIntegracaoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if integracao == nil {
		return nil, ErrPdvIntegracaoNaoEncontrada
	}

	antes := toJSON(integracao.ToResponse())

	if input.Provedor != nil {
		integracao.Provedor = *input.Provedor
	}
	if input.Ambiente != nil {
		integracao.Ambiente = *input.Ambiente
	}
	if input.URLBase != nil {
		integracao.URLBase = *input.URLBase
	}
	if input.APIKey != nil {
		integracao.APIKeyEncrypted = *input.APIKey // TODO: encrypt
	}
	if input.ClientID != nil {
		integracao.ClientID = *input.ClientID
	}
	if input.ClientSecret != nil {
		integracao.ClientSecretEncrypted = *input.ClientSecret // TODO: encrypt
	}
	if input.Ativo != nil {
		integracao.Ativo = *input.Ativo
	}

	if err := uc.pdvIntegracaoRepo.Update(ctx, integracao); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &integracao.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "pdv_integracao",
		EntidadeID: integracao.ID,
		Acao:       "update",
		Antes:      stringPtr(antes),
		Depois:     stringPtr(toJSON(integracao.ToResponse())),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return integracao.ToResponse(), nil
}

func (uc *PdvIntegracaoUsecase) Deletar(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	integracao, err := uc.pdvIntegracaoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if integracao == nil {
		return ErrPdvIntegracaoNaoEncontrada
	}

	antes := toJSON(integracao.ToResponse())

	if err := uc.pdvIntegracaoRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &integracao.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "pdv_integracao",
		EntidadeID: id,
		Acao:       "delete",
		Antes:      stringPtr(antes),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

// SincronizarVendas busca vendas no PDV externo e ingere no modelo canônico
func (uc *PdvIntegracaoUsecase) SincronizarVendas(ctx context.Context, integracaoID uuid.UUID, dataInicio, dataFim time.Time, usuarioID uuid.UUID, ip, userAgent string) (*SincronizacaoResult, error) {
	integracao, err := uc.pdvIntegracaoRepo.GetByID(ctx, integracaoID)
	if err != nil {
		return nil, err
	}
	if integracao == nil {
		return nil, ErrPdvIntegracaoNaoEncontrada
	}

	if !integracao.Ativo {
		return nil, errors.New("integração não está ativa")
	}

	// Buscar PDV de origem associado à loja
	pdvOrigem, err := uc.pdvOrigemRepo.GetByCodigo(ctx, integracao.EmpresaID, integracao.Provedor)
	if err != nil || pdvOrigem == nil {
		return nil, ErrPdvOrigemNaoEncontrado
	}

	// Criar cliente Gula
	gulaClient := gula.NewGulaClient(gula.ConfigGula{
		URLBase:      integracao.URLBase,
		APIKey:       integracao.APIKeyEncrypted, // TODO: decrypt in production
		ClientID:     integracao.ClientID,
		ClientSecret: integracao.ClientSecretEncrypted, // TODO: decrypt in production
		Ambiente:     integracao.Ambiente,
		Timeout:      30 * time.Second,
	})

	// Testar conexão
	if err := gulaClient.TestarConexao(ctx); err != nil {
		return nil, fmt.Errorf("falha ao conectar com Gula Menu: %w", err)
	}

	// Buscar vendas no Gula
	vendasGula, err := gulaClient.BuscarVendas(ctx, dataInicio, dataFim)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar vendas no Gula: %w", err)
	}

	result := &SincronizacaoResult{
		IntegracaoID:     integracaoID,
		DataInicio:       dataInicio,
		DataFim:          dataFim,
		TotalEncontrados: len(vendasGula),
		TotalProcessados: 0,
		TotalErros:       0,
		Erros:            []string{},
	}

	// Processar cada venda
	for _, venda := range vendasGula {
		// Mapear para payload de ingestão
		payload, idempotencyKey, err := venda.MapearParaSalesEvent(pdvOrigem.ID)
		if err != nil {
			result.TotalErros++
			result.Erros = append(result.Erros, fmt.Sprintf("Venda %s: erro no mapeamento: %v", venda.NumeroExterno, err))
			continue
		}

		// Ingerir via SalesUsecase
		input := IngestSalesEventInput{
			EmpresaID:      integracao.EmpresaID,
			LojaID:         integracao.LojaID,
			PdvOrigemID:    pdvOrigem.ID,
			TipoEvento:     "venda",
			Payload:        payload,
			IdempotencyKey: idempotencyKey,
		}

		_, err = uc.salesUsecase.IngestSalesEvent(ctx, input)
		if err != nil {
			result.TotalErros++
			result.Erros = append(result.Erros, fmt.Sprintf("Venda %s: erro na ingestão: %v", venda.NumeroExterno, err))
			continue
		}

		result.TotalProcessados++
	}

	// Atualizar última sincronização
	if err := uc.pdvIntegracaoRepo.UpdateUltimaSincronizacao(ctx, integracaoID, time.Now()); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &integracao.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "pdv_integracao",
		EntidadeID: integracaoID,
		Acao:       "sync",
		Depois:     stringPtr(toJSON(result)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return result, nil
}

type SincronizacaoResult struct {
	IntegracaoID     uuid.UUID `json:"integracao_id"`
	DataInicio       time.Time `json:"data_inicio"`
	DataFim          time.Time `json:"data_fim"`
	TotalEncontrados int       `json:"total_encontrados"`
	TotalProcessados int       `json:"total_processados"`
	TotalErros       int       `json:"total_erros"`
	Erros            []string  `json:"erros,omitempty"`
}
