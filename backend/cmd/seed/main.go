package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/adapter/db"
	"github.com/intelligate/erp/backend/internal/config"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}
	defer dbpool.Close()

	// Test database connection
	ctx := context.Background()
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established")

	// Initialize repositories
	empresaRepo := db.NewEmpresaRepository(dbpool)
	lojaRepo := db.NewLojaRepository(dbpool)
	usuarioRepo := db.NewUsuarioRepository(dbpool)
	papelRepo := db.NewPapelRepository(dbpool)
	permissaoRepo := db.NewPermissaoRepository(dbpool)
	usuarioPapelRepo := db.NewUsuarioPapelRepository(dbpool)
	papelPermissaoRepo := db.NewPapelPermissaoRepository(dbpool)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(usuarioRepo, cfg.JWTSecret, cfg.JWTExpiry)

	// Run seed
	if err := runSeed(ctx, authUsecase, empresaRepo, lojaRepo, usuarioRepo, papelRepo, permissaoRepo, usuarioPapelRepo, papelPermissaoRepo); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}

	log.Println("Seed completed successfully!")
}

func runSeed(ctx context.Context, authUsecase *usecase.AuthUsecase, empresaRepo *db.EmpresaRepository, lojaRepo *db.LojaRepository, usuarioRepo *db.UsuarioRepository, papelRepo *db.PapelRepository, permissaoRepo *db.PermissaoRepository, usuarioPapelRepo *db.UsuarioPapelRepository, papelPermissaoRepo *db.PapelPermissaoRepository) error {
	// 1. Create or get empresa
	empresa, err := getOrCreateEmpresa(ctx, empresaRepo)
	if err != nil {
		return err
	}
	log.Printf("Empresa: %s (ID: %s)", empresa.RazaoSocial, empresa.ID)

	// 2. Create or get loja
	loja, err := getOrCreateLoja(ctx, lojaRepo, empresa.ID)
	if err != nil {
		return err
	}
	log.Printf("Loja: %s (ID: %s)", loja.Nome, loja.ID)

	// 3. Create or get permissions
	permissoes, err := getOrCreatePermissoes(ctx, permissaoRepo)
	if err != nil {
		return err
	}
	log.Printf("Permissões criadas: %d", len(permissoes))

	// 4. Create or get papel ADMIN
	papelAdmin, err := getOrCreatePapelAdmin(ctx, papelRepo, empresa.ID, permissaoRepo, papelPermissaoRepo, permissoes)
	if err != nil {
		return err
	}
	log.Printf("Papel ADMIN: %s (ID: %s)", papelAdmin.Nome, papelAdmin.ID)

	// 5. Create or get admin user
	adminUser, err := getOrCreateAdminUser(ctx, usuarioRepo, authUsecase, empresa.ID)
	if err != nil {
		return err
	}
	log.Printf("Usuário admin: %s (ID: %s)", adminUser.Email, adminUser.ID)

	// 6. Associate user with ADMIN role
	err = associateUserWithRole(ctx, usuarioPapelRepo, adminUser.ID, papelAdmin.ID, loja.ID)
	if err != nil {
		return err
	}
	log.Printf("Usuário associado ao papel ADMIN na loja %s", loja.Nome)

	return nil
}

func getOrCreateEmpresa(ctx context.Context, repo *db.EmpresaRepository) (*domain.Empresa, error) {
	// Try to find by CNPJ
	cnpj := "12.345.678/0001-90"
	empresa, err := repo.GetByCNPJ(ctx, cnpj)
	if err == nil && empresa != nil {
		log.Println("Empresa já existe, pulando criação")
		return empresa, nil
	}

	// Create new empresa
	nomeFantasia := "IntelliGate"
	empresa = &domain.Empresa{
		CNPJ:         cnpj,
		RazaoSocial:  "IntelliGate ERP Ltda",
		NomeFantasia: &nomeFantasia,
		Ativo:        true,
	}
	if err := repo.Create(ctx, empresa); err != nil {
		return nil, err
	}
	return empresa, nil
}

func getOrCreateLoja(ctx context.Context, repo *db.LojaRepository, empresaID uuid.UUID) (*domain.Loja, error) {
	// Try to find by codigo
	codigo := "LOJA001"
	lojas, err := repo.ListByEmpresa(ctx, empresaID, 100, 0)
	if err == nil {
		for _, l := range lojas {
			if l.Codigo == codigo {
				log.Println("Loja já existe, pulando criação")
				return l, nil
			}
		}
	}

	// Create new loja
	endereco := "Rua Principal, 123 - Centro"
	loja := &domain.Loja{
		EmpresaID:   empresaID,
		Codigo:      codigo,
		Nome:        "Loja Principal",
		Endereco:    &endereco,
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	if err := repo.Create(ctx, loja); err != nil {
		return nil, err
	}
	return loja, nil
}

func getOrCreatePermissoes(ctx context.Context, repo *db.PermissaoRepository) (map[string]*domain.Permissao, error) {
	permissoesNecessarias := []struct {
		chave     string
		descricao string
	}{
		{"loja:create", "Criar lojas"},
		{"loja:read", "Visualizar lojas"},
		{"loja:update", "Atualizar lojas"},
		{"loja:delete", "Excluir lojas"},
		{"usuario:create", "Criar usuários"},
		{"usuario:read", "Visualizar usuários"},
		{"usuario:update", "Atualizar usuários"},
		{"usuario:delete", "Excluir usuários"},
		{"papel:create", "Criar papéis"},
		{"papel:read", "Visualizar papéis"},
		{"papel:update", "Atualizar papéis"},
		{"papel:delete", "Excluir papéis"},
		{"permissao:read", "Visualizar permissões"},
		{"empresa:read", "Visualizar empresa"},
		{"empresa:update", "Atualizar empresa"},
		{"auditoria:read", "Visualizar auditoria"},
	}

	result := make(map[string]*domain.Permissao)
	for _, p := range permissoesNecessarias {
		permissao, err := repo.GetByChave(ctx, p.chave)
		if err == nil && permissao != nil {
			result[p.chave] = permissao
			continue
		}

		descricao := p.descricao
		permissao = &domain.Permissao{
			Chave:     p.chave,
			Descricao: &descricao,
		}
		if err := repo.Create(ctx, permissao); err != nil {
			return nil, err
		}
		result[p.chave] = permissao
	}
	return result, nil
}

func getOrCreatePapelAdmin(ctx context.Context, papelRepo *db.PapelRepository, empresaID uuid.UUID, permissaoRepo *db.PermissaoRepository, papelPermissaoRepo *db.PapelPermissaoRepository, permissoes map[string]*domain.Permissao) (*domain.Papel, error) {
	// Try to find existing ADMIN role
	papeis, err := papelRepo.ListByEmpresa(ctx, empresaID, 100, 0)
	if err == nil {
		for _, p := range papeis {
			if p.Nome == "ADMIN" {
				log.Println("Papel ADMIN já existe, pulando criação")
				return p, nil
			}
		}
	}

	// Create ADMIN role
	descricao := "Administrador do sistema com acesso total"
	papelAdmin := &domain.Papel{
		EmpresaID: empresaID,
		Nome:      "ADMIN",
		Descricao: &descricao,
		Sistema:   true,
	}
	if err := papelRepo.Create(ctx, papelAdmin); err != nil {
		return nil, err
	}

	// Associate all permissions with ADMIN role
	for _, permissao := range permissoes {
		if err := papelPermissaoRepo.Create(ctx, papelAdmin.ID, permissao.ID); err != nil {
			return nil, err
		}
	}

	return papelAdmin, nil
}

func getOrCreateAdminUser(ctx context.Context, usuarioRepo *db.UsuarioRepository, authUsecase *usecase.AuthUsecase, empresaID uuid.UUID) (*domain.Usuario, error) {
	email := "admin@intelligate.com"

	// Try to find existing admin user
	usuario, err := usuarioRepo.GetByEmailAndEmpresa(ctx, email, empresaID)
	if err == nil && usuario != nil {
		log.Println("Usuário admin já existe, pulando criação")
		return usuario, nil
	}

	// Hash password
	senhaHash, err := authUsecase.HashSenha("admin123")
	if err != nil {
		return nil, err
	}

	// Create admin user
	usuario = &domain.Usuario{
		EmpresaID: empresaID,
		Email:     email,
		Nome:      "Administrador",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	if err := usuarioRepo.Create(ctx, usuario); err != nil {
		return nil, err
	}
	return usuario, nil
}

func associateUserWithRole(ctx context.Context, repo *db.UsuarioPapelRepository, usuarioID, papelID, lojaID uuid.UUID) error {
	// Check if association already exists
	associacoes, err := repo.ListByUsuario(ctx, usuarioID)
	if err == nil {
		for _, a := range associacoes {
			if a.PapelID == papelID && a.LojaID != nil && *a.LojaID == lojaID {
				log.Println("Associação usuário-papel-loja já existe, pulando")
				return nil
			}
		}
	}

	// Create association
	associacao := &domain.UsuarioPapel{
		UsuarioID: usuarioID,
		PapelID:   papelID,
		LojaID:    &lojaID,
	}
	return repo.Create(ctx, associacao)
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
