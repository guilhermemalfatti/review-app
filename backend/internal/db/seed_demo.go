package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type demoResident struct {
	email       string
	displayName string
}

type demoProvider struct {
	name     string
	category string
	phone    string
	notes    string
}

type demoReview struct {
	residentEmail string
	providerName  string
	anonymous     bool
	recommend     bool
	price         *int
	quality       *int
	deadline      *int
	comment       string
	daysAgo       int
}

func intPtr(v int) *int { return &v }

// SeedDemo inserts sample providers and reviews when SEED_DEMO is enabled.
// It is idempotent: existing providers (same name) and reviews (same user+provider) are skipped.
// The bootstrap Seed (condo + admin) is separate and always runs.
func SeedDemo(ctx context.Context, pool *pgxpool.Pool, condoID uuid.UUID) error {
	residents := []demoResident{
		{email: "ana.souza@example.com", displayName: "Ana Souza"},
		{email: "carlos.mendes@example.com", displayName: "Carlos Mendes"},
		{email: "lucia.ferreira@example.com", displayName: "Lúcia Ferreira"},
		{email: "roberto.alves@example.com", displayName: "Roberto Alves"},
		{email: "marina.costa@example.com", displayName: "Marina Costa"},
		{email: "paulo.nunes@example.com", displayName: "Paulo Nunes"},
		{email: "helena.dias@example.com", displayName: "Helena Dias"},
	}

	providers := []demoProvider{
		{name: "João Elétrica", category: "Eletricista", phone: "(51) 99811-2200", notes: "Instalações e reparos residenciais."},
		{name: "Hidráulica Silva", category: "Encanador", phone: "(51) 99722-3311", notes: "Vazamentos e troca de registros."},
		{name: "Pinturas Horizonte", category: "Pintor", phone: "(51) 99633-4422", notes: "Interna e externa, orçamento sem compromisso."},
		{name: "Pedreiro Zé Carlos", category: "Pedreiro", phone: "(51) 99544-5533", notes: "Reformas pequenas e muros."},
		{name: "Marcenaria do Vale", category: "Marceneiro", phone: "(51) 99455-6644", notes: "Armários sob medida."},
		{name: "Jardins da Serra", category: "Jardineiro", phone: "(51) 99366-7755", notes: "Manutenção mensal de jardins."},
		{name: "Limpeza Clara", category: "Limpeza", phone: "(51) 99277-8866", notes: "Faxina pós-obra e residencial."},
		{name: "Clima Sul", category: "Ar-condicionado", phone: "(51) 99188-9977", notes: "Instalação e manutenção."},
		{name: "Vidraçaria Cristal", category: "Outros", phone: "(51) 99099-1100", notes: "Box, espelhos e janelas."},
		{name: "Chaveiro Noturno", category: "Outros", phone: "(51) 98900-2211", notes: "Atendimento 24h."},
		{name: "Elétrica Rápida", category: "Eletricista", phone: "(51) 98811-3322", notes: "Emergências e quadro de luz."},
		{name: "Encanamentos Norte", category: "Encanador", phone: "(51) 98722-4433", notes: "Desentupimento e caixas d'água."},
		{name: "Cor e Luz Pinturas", category: "Pintor", phone: "(51) 98633-5544", notes: "Acabamento fino e texturas."},
		{name: "Verde Vivo", category: "Jardineiro", phone: "(51) 98544-6655", notes: "Poda e irrigação."},
		{name: "Fresh Clean", category: "Limpeza", phone: "(51) 98455-7766", notes: "Limpeza semanal de apartamentos."},
	}

	reviews := []demoReview{
		// João Elétrica
		{residentEmail: "ana.souza@example.com", providerName: "João Elétrica", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(4), comment: "Resolveu o curto na cozinha no mesmo dia. Muito cuidadoso e educado.", daysAgo: 12},
		{residentEmail: "carlos.mendes@example.com", providerName: "João Elétrica", anonymous: true, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(5), comment: "Pontual e deixou tudo limpo. Recomendo sem dúvida.", daysAgo: 28},
		{residentEmail: "paulo.nunes@example.com", providerName: "João Elétrica", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(2), deadline: intPtr(3), comment: "Cobrou caro e o problema voltou em uma semana. Fiquei decepcionado.", daysAgo: 50},
		{residentEmail: "helena.dias@example.com", providerName: "João Elétrica", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(5), comment: "Trocou as tomadas da sala com cuidado. Preço justo.", daysAgo: 6},

		// Hidráulica Silva
		{residentEmail: "lucia.ferreira@example.com", providerName: "Hidráulica Silva", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(4), comment: "Consertou o vazamento do banheiro sem quebrar o piso. Excelente.", daysAgo: 8},
		{residentEmail: "roberto.alves@example.com", providerName: "Hidráulica Silva", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(3), deadline: intPtr(2), comment: "Demorou dias para voltar com o orçamento. Serviço ok, mas atraso demais.", daysAgo: 40},
		{residentEmail: "marina.costa@example.com", providerName: "Hidráulica Silva", anonymous: true, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(5), comment: "Vieram rápido no fim de semana. Resolveram o cano estourado.", daysAgo: 14},
		{residentEmail: "ana.souza@example.com", providerName: "Hidráulica Silva", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(2), comment: "Deixaram sujeira e o registro ainda pingava. Precisei chamar outro.", daysAgo: 55},

		// Pinturas Horizonte
		{residentEmail: "marina.costa@example.com", providerName: "Pinturas Horizonte", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(5), comment: "Pintaram a sala em dois dias. Acabamento excelente, sem cheiro forte.", daysAgo: 5},
		{residentEmail: "ana.souza@example.com", providerName: "Pinturas Horizonte", anonymous: true, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(5), comment: "Bom preço e sem bagunça. Protegeram os móveis direitinho.", daysAgo: 19},
		{residentEmail: "carlos.mendes@example.com", providerName: "Pinturas Horizonte", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(2), comment: "Ficaram manchas na parede e atrasaram três dias. Não recomendo.", daysAgo: 70},
		{residentEmail: "helena.dias@example.com", providerName: "Pinturas Horizonte", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(4), comment: "Fachada do bloco ficou nova. Equipe organizada.", daysAgo: 21},

		// Pedreiro Zé Carlos
		{residentEmail: "carlos.mendes@example.com", providerName: "Pedreiro Zé Carlos", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(3), comment: "Fez o muro dos fundos. Trabalho sólido e honesto.", daysAgo: 60},
		{residentEmail: "lucia.ferreira@example.com", providerName: "Pedreiro Zé Carlos", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(4), comment: "Reformou o lavabo. Acabamento caprichado.", daysAgo: 25},
		{residentEmail: "roberto.alves@example.com", providerName: "Pedreiro Zé Carlos", anonymous: true, recommend: false, price: intPtr(2), quality: intPtr(2), deadline: intPtr(1), comment: "Sumiu no meio da obra e voltou só depois de muita cobrança. Péssimo.", daysAgo: 90},
		{residentEmail: "paulo.nunes@example.com", providerName: "Pedreiro Zé Carlos", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(4), comment: "Consertou a laje com calma. Explicou tudo antes de começar.", daysAgo: 17},

		// Marcenaria do Vale
		{residentEmail: "lucia.ferreira@example.com", providerName: "Marcenaria do Vale", anonymous: false, recommend: true, price: intPtr(3), quality: intPtr(5), deadline: intPtr(4), comment: "Armário da cozinha ficou perfeito. Um pouco caro, mas vale cada centavo.", daysAgo: 22},
		{residentEmail: "marina.costa@example.com", providerName: "Marcenaria do Vale", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(3), comment: "Estante sob medida linda. Atrasou um pouco, mas o resultado compensou.", daysAgo: 35},
		{residentEmail: "ana.souza@example.com", providerName: "Marcenaria do Vale", anonymous: true, recommend: false, price: intPtr(2), quality: intPtr(3), deadline: intPtr(2), comment: "Portas tortas e demora absurda. Tive que pedir ajuste várias vezes.", daysAgo: 80},

		// Jardins da Serra
		{residentEmail: "roberto.alves@example.com", providerName: "Jardins da Serra", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(5), comment: "Vêm todo mês. Jardim sempre em ordem e plantas saudáveis.", daysAgo: 3},
		{residentEmail: "helena.dias@example.com", providerName: "Jardins da Serra", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(5), comment: "Podaram sem destruir as flores. Muito atentos.", daysAgo: 16},
		{residentEmail: "carlos.mendes@example.com", providerName: "Jardins da Serra", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(3), comment: "Esqueceram de regar e algumas plantas morreram. Fraco.", daysAgo: 48},

		// Limpeza Clara
		{residentEmail: "marina.costa@example.com", providerName: "Limpeza Clara", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(5), comment: "Faxina pós-obra impecável. Parecia apartamento novo.", daysAgo: 15},
		{residentEmail: "paulo.nunes@example.com", providerName: "Limpeza Clara", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(4), comment: "Limparam vidros e azulejos com capricho. Pontuais.", daysAgo: 9},
		{residentEmail: "lucia.ferreira@example.com", providerName: "Limpeza Clara", anonymous: true, recommend: false, price: intPtr(2), quality: intPtr(2), deadline: intPtr(4), comment: "Passaram por cima da sujeira atrás do fogão. Não valeu o preço.", daysAgo: 38},

		// Clima Sul
		{residentEmail: "ana.souza@example.com", providerName: "Clima Sul", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(4), comment: "Instalaram o ar da suíte sem dor de cabeça. Explicaram a manutenção.", daysAgo: 33},
		{residentEmail: "roberto.alves@example.com", providerName: "Clima Sul", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(5), comment: "Limpeza do filtro rápida e barata. Voltarei todo ano.", daysAgo: 10},
		{residentEmail: "helena.dias@example.com", providerName: "Clima Sul", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(2), deadline: intPtr(2), comment: "Aparelho veio com barulho e o técnico sumiu. Péssimo pós-venda.", daysAgo: 62},

		// Vidraçaria Cristal
		{residentEmail: "ana.souza@example.com", providerName: "Vidraçaria Cristal", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(4), comment: "Box do banheiro novo, bem instalado e sem vazamento.", daysAgo: 18},
		{residentEmail: "carlos.mendes@example.com", providerName: "Vidraçaria Cristal", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(3), comment: "Espelho grande na entrada ficou ótimo. Recomendo.", daysAgo: 27},
		{residentEmail: "marina.costa@example.com", providerName: "Vidraçaria Cristal", anonymous: true, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(1), comment: "Vidro veio riscado e a troca demorou um mês. Experiência ruim.", daysAgo: 75},

		// Chaveiro Noturno
		{residentEmail: "carlos.mendes@example.com", providerName: "Chaveiro Noturno", anonymous: true, recommend: true, price: intPtr(3), quality: intPtr(5), deadline: intPtr(5), comment: "Vieram de madrugada quando travei a porta. Salvaram o dia.", daysAgo: 7},
		{residentEmail: "lucia.ferreira@example.com", providerName: "Chaveiro Noturno", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(5), comment: "Rápidos e honestos no preço da madrugada.", daysAgo: 20},
		{residentEmail: "paulo.nunes@example.com", providerName: "Chaveiro Noturno", anonymous: false, recommend: false, price: intPtr(1), quality: intPtr(3), deadline: intPtr(4), comment: "Cobrança abusiva na madrugada. Abriram a porta, mas o valor foi absurdo.", daysAgo: 44},

		// Elétrica Rápida
		{residentEmail: "lucia.ferreira@example.com", providerName: "Elétrica Rápida", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(5), comment: "Trocaram o disjuntor rapidinho. Atendimento claro.", daysAgo: 11},
		{residentEmail: "roberto.alves@example.com", providerName: "Elétrica Rápida", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(5), comment: "Emergência à noite, chegaram em 40 minutos. Nota 10.", daysAgo: 4},
		{residentEmail: "marina.costa@example.com", providerName: "Elétrica Rápida", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(2), deadline: intPtr(3), comment: "Diagnóstico errado e cobraram visita à toa. Evitem.", daysAgo: 58},

		// Encanamentos Norte
		{residentEmail: "carlos.mendes@example.com", providerName: "Encanamentos Norte", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(4), comment: "Desentupiram a cozinha sem drama. Preço combinado.", daysAgo: 9},
		{residentEmail: "ana.souza@example.com", providerName: "Encanamentos Norte", anonymous: true, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(5), comment: "Caixa d'água limpa e lacrada. Muito profissionais.", daysAgo: 30},
		{residentEmail: "helena.dias@example.com", providerName: "Encanamentos Norte", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(3), deadline: intPtr(2), comment: "Problema voltou em dois dias. Tive que pagar de novo. Não indico.", daysAgo: 66},

		// Cor e Luz Pinturas
		{residentEmail: "lucia.ferreira@example.com", providerName: "Cor e Luz Pinturas", anonymous: true, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(4), comment: "Textura da parede ficou linda. Detalhistas.", daysAgo: 26},
		{residentEmail: "paulo.nunes@example.com", providerName: "Cor e Luz Pinturas", anonymous: false, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(5), comment: "Quarto do menino ficou ótimo. Sem sujeira no corredor.", daysAgo: 13},
		{residentEmail: "roberto.alves@example.com", providerName: "Cor e Luz Pinturas", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(2), comment: "Cor ficou diferente do combinado e não quiseram refazer. Chateado.", daysAgo: 52},

		// Verde Vivo
		{residentEmail: "marina.costa@example.com", providerName: "Verde Vivo", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(3), comment: "Podaram demais as árvores. Ficou careca. Não chamaria de novo.", daysAgo: 45},
		{residentEmail: "ana.souza@example.com", providerName: "Verde Vivo", anonymous: false, recommend: false, price: intPtr(2), quality: intPtr(1), deadline: intPtr(2), comment: "Mataram minha roseira. Zero cuidado. Péssimo serviço.", daysAgo: 71},
		{residentEmail: "carlos.mendes@example.com", providerName: "Verde Vivo", anonymous: true, recommend: true, price: intPtr(4), quality: intPtr(4), deadline: intPtr(4), comment: "Para grama e sebes foi ok. Só não peçam poda fina.", daysAgo: 23},

		// Fresh Clean
		{residentEmail: "roberto.alves@example.com", providerName: "Fresh Clean", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(4), deadline: intPtr(5), comment: "Limpeza semanal confiável. Sempre no horário.", daysAgo: 2},
		{residentEmail: "helena.dias@example.com", providerName: "Fresh Clean", anonymous: false, recommend: true, price: intPtr(5), quality: intPtr(5), deadline: intPtr(5), comment: "Apartamento brilha. Equipe educada e discreta.", daysAgo: 8},
		{residentEmail: "paulo.nunes@example.com", providerName: "Fresh Clean", anonymous: false, recommend: false, price: intPtr(3), quality: intPtr(2), deadline: intPtr(4), comment: "Esqueceram o banheiro da suíte duas vezes. Cancelei o plano.", daysAgo: 41},
		{residentEmail: "lucia.ferreira@example.com", providerName: "Fresh Clean", anonymous: true, recommend: true, price: intPtr(4), quality: intPtr(5), deadline: intPtr(5), comment: "Ótima para quem trabalha fora. Vale a mensalidade.", daysAgo: 12},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo seed: %w", err)
	}
	defer tx.Rollback(ctx)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo12345"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	residentIDs := make(map[string]uuid.UUID, len(residents))
	for _, r := range residents {
		var id uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM users WHERE condo_id = $1 AND email = $2`, condoID, r.email).Scan(&id)
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `
				INSERT INTO users (condo_id, email, password_hash, display_name, role)
				VALUES ($1, $2, $3, $4, 'resident')
				RETURNING id
			`, condoID, r.email, string(passwordHash), r.displayName).Scan(&id)
			if err != nil {
				return fmt.Errorf("seed demo resident %s: %w", r.email, err)
			}
		} else if err != nil {
			return fmt.Errorf("lookup demo resident %s: %w", r.email, err)
		}
		residentIDs[r.email] = id
	}

	var adminID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM users WHERE condo_id = $1 AND role = 'admin' ORDER BY created_at ASC LIMIT 1
	`, condoID).Scan(&adminID)
	if err != nil {
		return fmt.Errorf("lookup admin for demo providers: %w", err)
	}

	providerIDs := make(map[string]uuid.UUID, len(providers))
	createdProviders := 0
	for _, p := range providers {
		var id uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT id FROM providers WHERE condo_id = $1 AND name = $2
		`, condoID, p.name).Scan(&id)
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `
				INSERT INTO providers (condo_id, name, category, phone, notes, status, created_by, reviewed_by, reviewed_at)
				VALUES ($1, $2, $3, $4, $5, 'approved', $6, $6, now())
				RETURNING id
			`, condoID, p.name, p.category, p.phone, p.notes, adminID).Scan(&id)
			if err != nil {
				return fmt.Errorf("seed demo provider %s: %w", p.name, err)
			}
			createdProviders++
		} else if err != nil {
			return fmt.Errorf("lookup demo provider %s: %w", p.name, err)
		}
		providerIDs[p.name] = id
	}

	createdReviews := 0
	for _, rev := range reviews {
		userID, ok := residentIDs[rev.residentEmail]
		if !ok {
			continue
		}
		providerID, ok := providerIDs[rev.providerName]
		if !ok {
			continue
		}

		var existing uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT id FROM reviews WHERE user_id = $1 AND provider_id = $2
		`, userID, providerID).Scan(&existing)
		if err == nil {
			continue
		}
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("lookup demo review: %w", err)
		}

		serviceDate := time.Now().UTC().AddDate(0, 0, -rev.daysAgo).Format("2006-01-02")
		_, err = tx.Exec(ctx, `
			INSERT INTO reviews (
				provider_id, user_id, is_anonymous, recommend,
				score_price, score_quality, score_deadline,
				comment, service_date, status, reviewed_by, reviewed_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::date,'approved',$10,now())
		`, providerID, userID, rev.anonymous, rev.recommend,
			rev.price, rev.quality, rev.deadline,
			rev.comment, serviceDate, adminID,
		)
		if err != nil {
			return fmt.Errorf("seed demo review %s/%s: %w", rev.residentEmail, rev.providerName, err)
		}
		createdReviews++
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo seed: %w", err)
	}

	slog.Info("demo seed complete", "providers", createdProviders, "reviews", createdReviews)
	return nil
}
