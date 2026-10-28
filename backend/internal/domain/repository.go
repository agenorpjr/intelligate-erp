package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EmpresaRepository interface {
	Create(ctx context.Context, e *Empresa) error
	GetByID(ctx context.Context, id uuid.UUID) (*Empresa, error)
	GetByCNPJ(ctx context.Context, cnpj string) (*Empresa, error)
	List(ctx context.Context, limit, offset int) ([]*Empresa, error)
	Update(ctx context.Context, e *Empresa) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LojaRepository interface {
	Create(ctx context.Context, l *Loja) error
	GetByID(ctx context.Context, id uuid.UUID) (*Loja, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Loja, error)
	Update(ctx context.Context, l *Loja) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsuarioRepository interface {
	Create(ctx context.Context, u *Usuario) error
	GetByID(ctx context.Context, id uuid.UUID) (*Usuario, error)
	GetByEmail(ctx context.Context, email string) (*Usuario, error)
	GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*Usuario, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Usuario, error)
	Update(ctx context.Context, u *Usuario) error
	UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error
	UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PapelRepository interface {
	Create(ctx context.Context, p *Papel) error
	GetByID(ctx context.Context, id uuid.UUID) (*Papel, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Papel, error)
	Update(ctx context.Context, p *Papel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PermissaoRepository interface {
	Create(ctx context.Context, p *Permissao) error
	GetByID(ctx context.Context, id uuid.UUID) (*Permissao, error)
	GetByChave(ctx context.Context, chave string) (*Permissao, error)
	List(ctx context.Context, limit, offset int) ([]*Permissao, error)
}

type UsuarioPapelRepository interface {
	Create(ctx context.Context, up *UsuarioPapel) error
	GetByID(ctx context.Context, id uuid.UUID) (*UsuarioPapel, error)
	ListByUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*UsuarioPapelComDetalhes, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsuarioPapelComDetalhes struct {
	ID        uuid.UUID  `json:"id"`
	UsuarioID uuid.UUID  `json:"usuario_id"`
	PapelID   uuid.UUID  `json:"papel_id"`
	LojaID    *uuid.UUID `json:"loja_id,omitempty"`
	PapelNome string     `json:"papel_nome"`
	LojaNome  *string    `json:"loja_nome,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type PapelPermissaoRepository interface {
	Create(ctx context.Context, papelID, permissaoID uuid.UUID) error
	Delete(ctx context.Context, papelID, permissaoID uuid.UUID) error
	ListPermissoesByPapel(ctx context.Context, papelID uuid.UUID) ([]*Permissao, error)
	ListPermissoesByUsuario(ctx context.Context, usuarioID, lojaID uuid.UUID) ([]*Permissao, error)
}

type AuditoriaRepository interface {
	Create(ctx context.Context, a *Auditoria) error
	List(ctx context.Context, filter AuditoriaFilter) ([]*Auditoria, error)
}

// Sales domain repositories
type PdvOrigemRepository interface {
	Create(ctx context.Context, p *PdvOrigem) error
	GetByID(ctx context.Context, id uuid.UUID) (*PdvOrigem, error)
	GetByCodigo(ctx context.Context, empresaID uuid.UUID, codigo string) (*PdvOrigem, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*PdvOrigem, error)
	Update(ctx context.Context, p *PdvOrigem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoriaCanonicaRepository interface {
	Create(ctx context.Context, c *CategoriaCanonica) error
	GetByID(ctx context.Context, id uuid.UUID) (*CategoriaCanonica, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*CategoriaCanonica, error)
	Update(ctx context.Context, c *CategoriaCanonica) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProdutoCanonicoRepository interface {
	Create(ctx context.Context, p *ProdutoCanonico) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProdutoCanonico, error)
	GetBySKU(ctx context.Context, empresaID uuid.UUID, sku string) (*ProdutoCanonico, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*ProdutoCanonico, error)
	Update(ctx context.Context, p *ProdutoCanonico) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PedidoCanonicoRepository interface {
	Create(ctx context.Context, p *PedidoCanonico) error
	GetByID(ctx context.Context, id uuid.UUID) (*PedidoCanonico, error)
	GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*PedidoCanonico, error)
	List(ctx context.Context, filter PedidoCanonicoFilter) ([]*PedidoCanonico, error)
	Update(ctx context.Context, p *PedidoCanonico) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PedidoItemCanonicoRepository interface {
	Create(ctx context.Context, i *PedidoItemCanonico) error
	GetByID(ctx context.Context, id uuid.UUID) (*PedidoItemCanonico, error)
	ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*PedidoItemCanonico, error)
	Update(ctx context.Context, i *PedidoItemCanonico) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error
}

type PedidoPagamentoCanonicoRepository interface {
	Create(ctx context.Context, p *PedidoPagamentoCanonico) error
	GetByID(ctx context.Context, id uuid.UUID) (*PedidoPagamentoCanonico, error)
	ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*PedidoPagamentoCanonico, error)
	Update(ctx context.Context, p *PedidoPagamentoCanonico) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error
}

type EventoIngestaoRepository interface {
	Create(ctx context.Context, e *EventoIngestao) error
	GetByID(ctx context.Context, id uuid.UUID) (*EventoIngestao, error)
	GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*EventoIngestao, error)
	List(ctx context.Context, filter EventoIngestaoFilter) ([]*EventoIngestao, error)
	Update(ctx context.Context, e *EventoIngestao) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type VendasResumoRepository interface {
	GetResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time) (*VendasResumo, error)
}

// CMV domain repositories
type InsumoRepository interface {
	Create(ctx context.Context, i *Insumo) error
	GetByID(ctx context.Context, id uuid.UUID) (*Insumo, error)
	GetByNome(ctx context.Context, empresaID uuid.UUID, nome string) (*Insumo, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*Insumo, error)
	Update(ctx context.Context, i *Insumo) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateCusto(ctx context.Context, c *InsumoCusto) error
	GetCustoVigente(ctx context.Context, insumoID uuid.UUID, dataVigencia time.Time) (*InsumoCusto, error)
	GetCustoAtual(ctx context.Context, insumoID uuid.UUID) (*InsumoCusto, error)
	ListCustosByInsumo(ctx context.Context, insumoID uuid.UUID, limit, offset int) ([]*InsumoCusto, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (*Insumo, error)
	UpdateCustoUnitario(ctx context.Context, id uuid.UUID, custoUnitario float64) error
}

type ProdutoRepository interface {
	Create(ctx context.Context, p *Produto) error
	GetByID(ctx context.Context, id uuid.UUID) (*Produto, error)
	GetBySKUExterno(ctx context.Context, empresaID uuid.UUID, skuExterno string) (*Produto, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*Produto, error)
	Update(ctx context.Context, p *Produto) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FichaTecnicaRepository interface {
	Create(ctx context.Context, ft *FichaTecnica) error
	GetByID(ctx context.Context, id uuid.UUID) (*FichaTecnica, error)
	GetByProdutoLoja(ctx context.Context, produtoID, lojaID uuid.UUID) (*FichaTecnica, error)
	ListByProduto(ctx context.Context, produtoID uuid.UUID, ativo *bool, limit, offset int) ([]*FichaTecnica, error)
	ListByLoja(ctx context.Context, lojaID uuid.UUID, ativo *bool, limit, offset int) ([]*FichaTecnica, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*FichaTecnica, error)
	Update(ctx context.Context, ft *FichaTecnica) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateComponente(ctx context.Context, c *FichaTecnicaComponente) error
	GetComponentes(ctx context.Context, fichaTecnicaID uuid.UUID) ([]*FichaTecnicaComponente, error)
	UpdateComponente(ctx context.Context, c *FichaTecnicaComponente) error
	DeleteComponente(ctx context.Context, id uuid.UUID) error
	DeleteComponentesByFicha(ctx context.Context, fichaTecnicaID uuid.UUID) error
	GetCMVResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (*CMVResumo, error)
	GetCMVProdutosDetalhe(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time, limit, offset int) ([]*CMVProdutoDetalhe, error)
	GetProdutosSemFichaTecnica(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time, limit, offset int) ([]*CMVProdutoDetalhe, error)
}

// PDV Integração Repository
type PdvIntegracaoRepository interface {
	Create(ctx context.Context, p *PdvIntegracao) error
	GetByID(ctx context.Context, id uuid.UUID) (*PdvIntegracao, error)
	GetByLojaProvedor(ctx context.Context, lojaID uuid.UUID, provedor string) (*PdvIntegracao, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, filter PdvIntegracaoFilter) ([]*PdvIntegracao, error)
	Update(ctx context.Context, p *PdvIntegracao) error
	UpdateUltimaSincronizacao(ctx context.Context, id uuid.UUID, ultimaSincronizacao time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Compras domain repositories
type FornecedorRepository interface {
	Create(ctx context.Context, f *Fornecedor) error
	GetByID(ctx context.Context, id uuid.UUID) (*Fornecedor, error)
	GetByCNPJCPF(ctx context.Context, empresaID uuid.UUID, cnpjCpf string) (*Fornecedor, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*Fornecedor, error)
	ListByFilter(ctx context.Context, empresaID uuid.UUID, ativo *bool, nome, cnpjCpf *string, limit, offset int) ([]*Fornecedor, error)
	Update(ctx context.Context, f *Fornecedor) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type OrdemCompraRepository interface {
	Create(ctx context.Context, oc *OrdemCompra) error
	GetByID(ctx context.Context, id uuid.UUID) (*OrdemCompra, error)
	GetWithItens(ctx context.Context, id uuid.UUID) (*OrdemCompra, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*OrdemCompra, error)
	ListByLoja(ctx context.Context, lojaID uuid.UUID, limit, offset int) ([]*OrdemCompra, error)
	ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *OrdemCompraStatus, dataInicio, dataFim *time.Time, limit, offset int) ([]*OrdemCompra, error)
	Update(ctx context.Context, oc *OrdemCompra) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrdemCompraStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateItem(ctx context.Context, item *OrdemCompraItem) error
	GetItem(ctx context.Context, id uuid.UUID) (*OrdemCompraItem, error)
	ListItensByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) ([]*OrdemCompraItem, error)
	UpdateItem(ctx context.Context, item *OrdemCompraItem) error
	UpdateItemQuantidadeRecebida(ctx context.Context, itemID uuid.UUID, quantidadeRecebida float64) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	DeleteItensByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) error
}

type RecebimentoCompraRepository interface {
	Create(ctx context.Context, rc *RecebimentoCompra) error
	GetByID(ctx context.Context, id uuid.UUID) (*RecebimentoCompra, error)
	ListByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) ([]*RecebimentoCompra, error)
	ListByFilter(ctx context.Context, ordemCompraID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time, limit, offset int) ([]*RecebimentoCompra, error)
	CreateItem(ctx context.Context, item *RecebimentoCompraItem) error
	ListItensByRecebimento(ctx context.Context, recebimentoID uuid.UUID) ([]*RecebimentoCompraItem, error)
}

type ContaPagarRepository interface {
	Create(ctx context.Context, cp *ContaPagar) error
	GetByID(ctx context.Context, id uuid.UUID) (*ContaPagar, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*ContaPagar, error)
	ListByFornecedor(ctx context.Context, fornecedorID uuid.UUID, limit, offset int) ([]*ContaPagar, error)
	ListByStatus(ctx context.Context, empresaID uuid.UUID, status ContaPagarStatus, limit, offset int) ([]*ContaPagar, error)
	ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *ContaPagarStatus, dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim *time.Time, limit, offset int) ([]*ContaPagar, error)
	Update(ctx context.Context, cp *ContaPagar) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status ContaPagarStatus, dataPagamento *time.Time, formaPagamento *string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
